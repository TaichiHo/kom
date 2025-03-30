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
	if _, err := kom.Clusters().RegisterInCluster(); err != nil {
		klog.Warningf("register in-cluster error: %v", err)
	}
	if _, err := kom.Clusters().RegisterByPathWithID(defaultKubeConfig, "default"); err != nil {
		klog.Warningf("register default cluster error: %v", err)
	}
	if err := kom.Clusters().RegisterClusterAPIConfigs(); err != nil {
		klog.Warningf("register cluster api configs error: %v", err)
	}
	kom.Clusters().Show()
}
