package myutil

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func GetClientSet() (*kubernetes.Clientset, error) {
	var err error
	var config *rest.Config
	var kubeconfig *string
	if home := homeDire(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(可选)输入kubeconfig文件得绝对路径")

	} else {
		kubeconfig = flag.String("kubeconfig", "", "输入kubeconfig文件得绝对路径")
	}
	flag.Parse()

	// 首先使用inCluster模式 Pod 中
	config, err = rest.InClusterConfig()
	if err != nil {
		// 使用Kubeconfig创建集群配置
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}
	// 创建ClientSet对象
	return kubernetes.NewForConfig(config)
}

func homeDire() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
