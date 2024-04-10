package openfalcon

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.io/ssslkj123/ephemeral-exporter/types"
	"k8s.io/klog"
)

type OpenFalconMetric struct {
	Endpoint    string  `json:"endpoint"`
	Metric      string  `json:"metric"`
	Timestamp   int64   `json:"timestamp"`
	Value       float64 `json:"value"`
	CounterType string  `json:"counterType"`
	Step        int64   `json:"step"`
	Tags        string  `json:"tags"`
}

func GetPodEphemeralStorageForOpenfalconMetrics(pods []types.PodEphemeralStorage) []OpenFalconMetric {
	metrics, ts := []OpenFalconMetric{}, time.Now().Unix()

	for _, v := range pods {
		// fmt.Printf("namespace=%s, name=%s, usage_bytes=%d, limit_bytes=%d, used_percent=%.2f%%\n", v.Namespace, v.Name, v.UsageBytes, v.LimitBytes, v.UsedPercent)
		for metricName, metricValue := range map[string]float64{
			"pod_ephemeral_storage_limit_bytes":   float64(v.LimitBytes),
			"pod_ephemeral_storage_usage_bytes":   float64(v.UsageBytes),
			"pod_ephemeral_storage_usage_percent": float64(v.UsedPercent),
		} {
			metrics = append(metrics, OpenFalconMetric{
				Endpoint:    v.Name,
				Metric:      metricName,
				Timestamp:   ts,
				Value:       metricValue,
				CounterType: "GAUGE",
				Step:        60,
				Tags:        fmt.Sprintf("cluster=%s, env=%s,namespace=%s", os.Getenv("cluster"), os.Getenv("env"), v.Namespace),
			})
		}
	}
	return metrics
}

func UpdateMetrics(pods []types.PodEphemeralStorage) {
	// openfalcon_push_url : http://172.31.201.206:1988/v1/push
	client := resty.New()
	client.OnError(func(req *resty.Request, err error) {
		klog.Info(err)
		if v, ok := err.(*resty.ResponseError); ok {
			klog.Error(v.Err)
		}
	}).
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			return 0, errors.New("quota Exceeded")
		})

	metrics := GetPodEphemeralStorageForOpenfalconMetrics(pods)

	// klog.Info(metrics)
	body, _ := json.Marshal(metrics)
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(fmt.Sprintf("http://%s:1988/v1/push", os.Getenv("NodeIP")))
	if err != nil {
		klog.Error(err.Error())
	}
	if resp.IsSuccess() && resp.String() == "success" {
		klog.Info("指标上报成功")
	} else {
		klog.Errorf("指标上报失败,请检查上报body, %s\n", body)
	}
}
