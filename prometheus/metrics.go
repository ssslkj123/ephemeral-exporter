package prometheus

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ssslkj123/ephemeral-exporter/types"
)

var (
	metricLimit        *prometheus.GaugeVec
	metricUsage        *prometheus.GaugeVec
	metricUsagePercent *prometheus.GaugeVec
)

func InitPrometheusMetrics(r *prometheus.Registry) {
	metricLimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pod_ephemeral_storage_limit_bytes",
		Help: "Ephemeral Storage for Pod Limit Bytes",
	}, []string{
		"node", "namespace", "pod", "env", "cluster",
	})

	metricUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pod_ephemeral_storage_usage_bytes",
		Help: "Ephemeral Storage for Pod Used Bytes",
	}, []string{
		"node", "namespace", "pod", "env", "cluster",
	})

	metricUsagePercent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pod_ephemeral_storage_usage_percent",
		Help: "Ephemeral Storage for Pod Used Percent",
	}, []string{
		"node", "namespace", "pod", "env", "cluster",
	})

	r.MustRegister(metricLimit)
	r.MustRegister(metricUsage)
	r.MustRegister(metricUsagePercent)
}

func UpdateMetrics(pods []types.PodEphemeralStorage) {
	for _, v := range pods {
		// fmt.Printf("namespace=%s, name=%s, usage_bytes=%d, limit_bytes=%d, used_percent=%.2f%%\n", v.Namespace, v.Name, v.UsageBytes, v.LimitBytes, v.UsedPercent)
		metricLimit.With(prometheus.Labels{
			"node":      os.Getenv("NodeIP"),
			"cluster":   os.Getenv("cluster"),
			"env":       os.Getenv("env"),
			"namespace": v.Namespace,
			"pod":       v.Name,
		}).Set(float64(v.LimitBytes))

		metricUsage.With(prometheus.Labels{
			"node":      os.Getenv("NodeIP"),
			"cluster":   os.Getenv("cluster"),
			"env":       os.Getenv("env"),
			"namespace": v.Namespace,
			"pod":       v.Name,
		}).Set(float64(v.UsageBytes))

		metricUsagePercent.With(prometheus.Labels{
			"node":      os.Getenv("NodeIP"),
			"cluster":   os.Getenv("cluster"),
			"env":       os.Getenv("env"),
			"namespace": v.Namespace,
			"pod":       v.Name,
		}).Set(float64(v.UsedPercent))
	}
}
