package utils

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func GetKubeConfigFromFile() *rest.Config {
	// loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// loadingRules.ExplicitPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
	// configOverrides := &clientcmd.ConfigOverrides{CurrentContext: "ack-test", Timeout: "3s"}
	// kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	// config, err := kubeConfig.ClientConfig()
	// if err != nil {
	// 	panic(err.Error())
	// }

	kubeConfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func GetKubeConfigInCluster() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return config
}

func GetKubeConfig() *rest.Config {
	// 当环境变量存在 KUBERNETES_SERVICE_HOST 与 KUBERNETES_SERVICE_PORT 时表示部署在集群，使用InCluster方式获取Config
	if _, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST"); ok {
		return GetKubeConfigInCluster()
	} else {
		return GetKubeConfigFromFile()
	}
}

func GetMetricClientSet() *metricsv.Clientset {
	clientset, err := metricsv.NewForConfig(GetKubeConfig())
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func GetClientSet() *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(GetKubeConfig())
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
