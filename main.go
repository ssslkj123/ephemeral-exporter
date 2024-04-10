package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"net/http"

	sdkprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cron "github.com/robfig/cron/v3"
	"github.io/ssslkj123/ephemeral-exporter/openfalcon"
	"github.io/ssslkj123/ephemeral-exporter/pods"
	"github.io/ssslkj123/ephemeral-exporter/prometheus"
	"k8s.io/klog"
)

func PreCheck() bool {
	// 依赖pod.spec.container.env当中定义的以下4个环境变量
	// 1.集群所处环境 env=""
	// 2.kubernetes集群名 cluster=""
	// 3.NodeName:
	//- name: NodeName
	// 	valueFrom:
	// 	fieldRef:
	// 	  fieldPath: spec.nodeName
	// 4.NodeIP:
	// - name: NodeIP
	//   valueFrom:
	// 	fieldRef:
	// 	  fieldPath: status.hostIP
	// envSlice := []string{"env", "cluster", "NodeIP", "NodeName"}
	// envMap := make(map[string]string, 4)
	// for _, v := range envSlice {
	// 	envMap[v] = os.Getenv(v)
	// 	if _, ok := envMap[v]; !ok {
	// 		klog.Errorf("Environment variable is missing %v", envMap[v])
	// 		return false
	// 	}
	// }

	envCheck := []string{"env", "cluster", "NodeIP", "NodeName"}
	for _, v := range envCheck {
		if _, ok := os.LookupEnv(v); !ok {
			// klog.Error("The environment variable env must provide")
			klog.Errorf("Environment variable is missing !")
			return false
		}
	}
	return true
}

func UpdateMetricsByCronTask() {

	//  等待组实现定时任务
	var wg sync.WaitGroup
	wg.Add(1)

	c := cron.New()
	c.AddFunc("@every 1m", func() {
		klog.Infof("每%s的定时任务启动\n", "1m")
		start := time.Now()
		// 获取Pod指标数据
		pods := pods.GetPodEphemeralStorage()
		klog.Infof("完成%d个Pod指标数据获取, 耗时%s\n", len(pods), time.Since(start))

		// pods 长度大于0时才包装指标上报
		if len(pods) > 0 {
			// 更新Prometheus Metrics指标
			prometheus.UpdateMetrics(pods)
			klog.Infof("完成Prometheus指标更新, 耗时%s\n", time.Since(start))

			// 更新OpenFalcon指标
			openfalcon.UpdateMetrics(pods)
			klog.Infof("完成OpenFalcon指标更新, 耗时%s\n", time.Since(start))
		}
		klog.Infof("每%s的定时任务结束\n", "1m")
	})

	c.Start()

	wg.Wait()
	// select {}
}

func main() {
	// 环境变量检测
	if !PreCheck() {
		panic(errors.New("Failed pre-check, please provide necessary environment variables"))
	}

	// 初始化Prometheus Metrics
	r := sdkprometheus.NewRegistry()
	prometheus.InitPrometheusMetrics(r)

	port := "9200"
	if customPort := os.Getenv("METRIC_PORT"); customPort != "" {
		port = customPort
	}
	uri := "/metrics"
	if customUri := os.Getenv("METRIC_URI"); customUri != "" {
		uri = customUri
	}
	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{EnableOpenMetrics: false})

	// 更新指标
	go UpdateMetricsByCronTask()

	// 启动http服务
	go http.Handle(uri, handler)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		klog.Errorf("Listener Failed : %s\n", err.Error())
		panic(err.Error())
	}
}
