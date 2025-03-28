package metadata

import "strings"

type ResourceInfo struct {
	Group      string
	Version    string
	Kind       string
	Namespaced bool
}

var resourceMap = map[string]ResourceInfo{
	// Namespace-level resources
	"pod":                            {Group: "", Version: "v1", Kind: "Pod", Namespaced: true},
	"deployment":                     {Group: "apps", Version: "v1", Kind: "Deployment", Namespaced: true},
	"statefulset":                    {Group: "apps", Version: "v1", Kind: "StatefulSet", Namespaced: true},
	"daemonset":                      {Group: "apps", Version: "v1", Kind: "DaemonSet", Namespaced: true},
	"replicaset":                     {Group: "apps", Version: "v1", Kind: "ReplicaSet", Namespaced: true},
	"service":                        {Group: "", Version: "v1", Kind: "Service", Namespaced: true},
	"configmap":                      {Group: "", Version: "v1", Kind: "ConfigMap", Namespaced: true},
	"secret":                         {Group: "", Version: "v1", Kind: "Secret", Namespaced: true},
	"ingress":                        {Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespaced: true},
	"networkpolicy":                  {Group: "networking.k8s.io", Version: "v1", Kind: "NetworkPolicy", Namespaced: true},
	"role":                           {Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role", Namespaced: true},
	"rolebinding":                    {Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding", Namespaced: true},
	"serviceaccount":                 {Group: "", Version: "v1", Kind: "ServiceAccount", Namespaced: true},
	"persistentvolumeclaim":          {Group: "", Version: "v1", Kind: "PersistentVolumeClaim", Namespaced: true},
	"horizontalpodautoscaler":        {Group: "autoscaling", Version: "v2", Kind: "HorizontalPodAutoscaler", Namespaced: true},
	"cronjob":                        {Group: "batch", Version: "v1", Kind: "CronJob", Namespaced: true},
	"job":                            {Group: "batch", Version: "v1", Kind: "Job", Namespaced: true},
	"node":                           {Group: "", Version: "v1", Kind: "Node", Namespaced: false},
	"namespace":                      {Group: "", Version: "v1", Kind: "Namespace", Namespaced: false},
	"persistentvolume":               {Group: "", Version: "v1", Kind: "PersistentVolume", Namespaced: false},
	"clusterrole":                    {Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole", Namespaced: false},
	"clusterrolebinding":             {Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding", Namespaced: false},
	"storageclass":                   {Group: "storage.k8s.io", Version: "v1", Kind: "StorageClass", Namespaced: false},
	"customresourcedefinition":       {Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition", Namespaced: false},
	"mutatingwebhookconfiguration":   {Group: "admissionregistration.k8s.io", Version: "v1", Kind: "MutatingWebhookConfiguration", Namespaced: false},
	"validatingwebhookconfiguration": {Group: "admissionregistration.k8s.io", Version: "v1", Kind: "ValidatingWebhookConfiguration", Namespaced: false},
}

// GetResourceInfo returns resource information based on the resource type string
func GetResourceInfo(resourceType string) (ResourceInfo, bool) {
	resourceType = strings.ToLower(resourceType)
	if info, exists := resourceMap[resourceType]; exists {
		return info, true
	}
	return ResourceInfo{}, false
}

// IsNamespaced determines if a resource is namespace-scoped
func IsNamespaced(resourceType string) bool {
	resourceType = strings.ToLower(resourceType)
	if info, exists := resourceMap[resourceType]; exists {
		return info.Namespaced
	}
	return false
}
