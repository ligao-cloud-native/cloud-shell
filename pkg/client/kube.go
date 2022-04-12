package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func NewRestClient(masterUrl, kubeConfigPath string) (*rest.Config, *kubernetes.Clientset) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeConfigPath)
	if err != nil {
		klog.Fatalf("Failed to build config, err: %v", err)
	}
	//kubeConfig.QPS = float32(config.Config.KubeAPIConfig.QPS)
	//kubeConfig.Burst = int(config.Config.KubeAPIConfig.Burst)
	//kubeConfig.ContentType = "application/json"

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Fatalf("Failed to create kube client, err: %v", err)
	}

	return kubeConfig, kubeClient
}
