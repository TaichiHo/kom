package example

import (
	"os"
	"path/filepath"

	"github.com/weibaohui/kom/callbacks"
	"github.com/weibaohui/kom/kom"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

func Connect() {
	callbacks.RegisterInit()

	defaultKubeConfig := os.Getenv("KUBECONFIG")
	if defaultKubeConfig == "" {
		defaultKubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	_, err := kom.Clusters().RegisterByPathWithID(defaultKubeConfig, "default")
	if err != nil {
		klog.Errorf("register cluster api configs error %v", err)
	}
	err = kom.Clusters().RegisterClusterAPIConfigs()
	if err != nil {
		klog.Errorf("register cluster api configs error %v", err)
	}
	kom.Clusters().Show()
}
