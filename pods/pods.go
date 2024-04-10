package pods

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/ssslkj123/ephemeral-exporter/types"
	"github.com/ssslkj123/ephemeral-exporter/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func GetPodEphemeralStorageUsage() []types.PodEphemeralStorage {
	start := time.Now()
	client := resty.New()

	podStatsSummary := &types.StatsSummaryForPodEphemeralStorage{}
	_, err := client.R().SetResult(podStatsSummary).Get(fmt.Sprintf("http://%s:10255/stats/summary", os.Getenv("NodeIP")))
	if err != nil {
		klog.Error(err.Error())
	}

	pods := []types.PodEphemeralStorage{}
	for _, pod := range podStatsSummary.Pods {
		// 临时存储空间使用量大于0的Pod才纳入处理
		if pod.EphemeralStorage.UsedBytes > 0 {
			pods = append(pods, types.PodEphemeralStorage{
				Name:       pod.PodRef.Name,
				Namespace:  pod.PodRef.Namespace,
				UsageBytes: pod.EphemeralStorage.UsedBytes,
			})
		}
	}
	klog.Infof("GetPodEphemeralStorageUsage: node is %s pods numbers is %d took %s", os.Getenv("NodeIP"), len(pods), time.Since(start))
	// klog.Info(pods)
	return pods
}

func GetPodEphemeralStorage() []types.PodEphemeralStorage {
	start := time.Now()
	result := []types.PodEphemeralStorage{}
	// 节点上使用临时存储空间的Pod
	podUsedDatas := GetPodEphemeralStorageUsage()

	// 没有找到使用临时存储空间Pod时立即返回
	if len(podUsedDatas) == 0 {
		return podUsedDatas
	}

	// 节点上所有Pod
	client := utils.GetClientSet()
	pods, _ := client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", os.Getenv("NodeName"))})

	for _, podUsedData := range podUsedDatas {
		for _, pod := range pods.Items {
			// 根据namespace+name匹配Pod
			if podUsedData.Namespace == pod.GetNamespace() && podUsedData.Name == pod.GetName() {
				limitValue := pod.Spec.Containers[0].Resources.Limits.StorageEphemeral().Value()
				// Pod只有定义了resources.limits.ephemeral-storage才处理，没有定义的就是无限使用
				if limitValue > 0 {
					// fmt.Printf("namespace=%s, name=%s, usage_bytes=%d, limit_bytes=%d\n", podUsedData.Namespace, podUsedData.Name, podUsedData.UsageBytes, limitValue)
					value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(podUsedData.UsageBytes)*100/float64(limitValue)), 64)
					podUsedData.UsedPercent = value
					podUsedData.LimitBytes = limitValue
					result = append(result, podUsedData)
				}
				break
			}
		}
	}
	klog.Infof("GetPodEphemeralStorage: node is %s pods number is %d took %s", os.Getenv("NodeIP"), len(result), time.Since(start))
	return result
}
