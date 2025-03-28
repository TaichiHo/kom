package kom

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"github.com/weibaohui/kom/kom/describe"
	"github.com/weibaohui/kom/kom/doc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var clusterInstances *ClusterInstances

// ClusterInstances manages multiple cluster instances
type ClusterInstances struct {
	clusters             map[string]*ClusterInst
	callbackRegisterFunc func(cluster *ClusterInst) func() // Callback function for registering parameters
}

// ClusterInst represents a single cluster instance
type ClusterInst struct {
	ID            string                       // Cluster ID
	Kubectl       *Kubectl                     // kom
	Client        *kubernetes.Clientset        // Kubernetes client
	Config        *rest.Config                 // REST config
	DynamicClient *dynamic.DynamicClient       // Dynamic client
	apiResources  []*metav1.APIResource        // Currently registered k8s resources
	crdList       []*unstructured.Unstructured // Currently registered k8s CRDs //TODO Update periodically or via Watch
	callbacks     *callbacks                   // Callbacks
	docs          *doc.Docs                    // Documentation
	serverVersion *version.Info                // Server version
	describerMap  map[schema.GroupKind]describe.ResourceDescriber
	Cache         *ristretto.Cache[string, any]
	openAPISchema *openapi_v2.Document // OpenAPI schema
}

// Clusters returns the cluster instances manager
func Clusters() *ClusterInstances {
	return clusterInstances
}

// Initialize
func init() {
	clusterInstances = &ClusterInstances{
		clusters: make(map[string]*ClusterInst),
	}
}

// DefaultCluster gets the default cluster, simplifying the calling method
func DefaultCluster() *Kubectl {
	return Clusters().DefaultCluster().Kubectl
}

// Cluster gets a cluster by ID
func Cluster(id string) *Kubectl {
	var cluster *ClusterInst
	if id == "" {
		cluster = Clusters().DefaultCluster()
	} else {
		cluster = Clusters().GetClusterById(id)
	}
	if cluster == nil {
		return nil
	}
	return cluster.Kubectl
}

// RegisterInCluster registers an InCluster configuration
func (c *ClusterInstances) RegisterInCluster() (*Kubectl, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("InCluster Error %v", err)
	}
	return c.RegisterByConfigWithID(config, "InCluster")
}

// SetRegisterCallbackFunc sets the callback registration function
func (c *ClusterInstances) SetRegisterCallbackFunc(callback func(cluster *ClusterInst) func()) {
	c.callbackRegisterFunc = callback
}

// RegisterByPath registers a cluster using a kubeconfig file path
func (c *ClusterInstances) RegisterByPath(path string) (*Kubectl, error) {
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPath Error %s %v", path, err)
	}
	return c.RegisterByConfig(config)
}

// RegisterByString registers a cluster using the string content of a kubeconfig file
func (c *ClusterInstances) RegisterByString(str string) (*Kubectl, error) {
	config, err := clientcmd.Load([]byte(str))
	if err != nil {
		return nil, fmt.Errorf("RegisterByString Error,content=:\n%s\n,err:%v", str, err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return c.RegisterByConfig(restConfig)
}

// RegisterByStringWithID registers a cluster using the string content of a kubeconfig file with a specific ID
func (c *ClusterInstances) RegisterByStringWithID(str string, id string) (*Kubectl, error) {
	config, err := clientcmd.Load([]byte(str))
	if err != nil {
		return nil, fmt.Errorf("RegisterByStringWithID Error content=\n%s\n,id:%s,err:%v", str, id, err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return c.RegisterByConfigWithID(restConfig, id)
}

// RegisterByPathWithID registers a cluster using a kubeconfig file path with a specific ID
func (c *ClusterInstances) RegisterByPathWithID(path string, id string) (*Kubectl, error) {
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPathWithID Error path:%s,id:%s,err:%v", path, id, err)
	}
	return c.RegisterByConfigWithID(config, id)
}

// RegisterByConfig registers a cluster using a REST config
func (c *ClusterInstances) RegisterByConfig(config *rest.Config) (*Kubectl, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	host := config.Host

	return c.RegisterByConfigWithID(config, host)
}

// RegisterByConfigWithID registers a cluster using a REST config with a specific ID
func (c *ClusterInstances) RegisterByConfigWithID(config *rest.Config, id string) (*Kubectl, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	config.QPS = 200
	config.Burst = 2000
	cluster, exists := clusterInstances.clusters[id]
	if exists {
		return cluster.Kubectl, nil
	} else {
		// Initialize when key doesn't exist
		k := initKubectl(config, id)
		cluster = &ClusterInst{
			ID:      id,
			Kubectl: k,
			Config:  config,
		}
		clusterInstances.clusters[id] = cluster

		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("RegisterByConfigWithID Error %s %v", id, err)
		}
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("RegisterByConfigWithID Error %s %v", id, err)
		}
		cluster.Client = client               // Kubernetes client
		cluster.DynamicClient = dynamicClient // Dynamic client
		// Cache
		cluster.apiResources = k.initializeAPIResources()       // API resources
		cluster.crdList = k.initializeCRDList(time.Minute * 10) // CRD list with 10-minute cache
		cluster.callbacks = k.initializeCallbacks()             // Callbacks
		cluster.serverVersion = k.initializeServerVersion()     // Server version
		cluster.openAPISchema = k.getOpenAPISchema()
		cluster.docs = doc.InitTrees(k.getOpenAPISchema()) // Documentation
		cluster.describerMap = k.initializeDescriberMap()  // Initialize describers
		if c.callbackRegisterFunc != nil {                 // Register callback method
			c.callbackRegisterFunc(cluster)
		}

		cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
			NumCounters: 1e7,     // number of keys to track frequency of (10M)
			MaxCost:     1 << 30, // maximum cost of cache (1GB)
			BufferItems: 64,      // number of keys per Get buffer
		})
		cluster.Cache = cache
		return k, nil
	}
}

// GetClusterById gets a cluster instance by ID
func (c *ClusterInstances) GetClusterById(id string) *ClusterInst {
	cluster, exists := c.clusters[id]
	if !exists {
		return nil
	}
	return cluster
}

// RemoveClusterById removes a cluster by ID
func (c *ClusterInstances) RemoveClusterById(id string) {
	delete(c.clusters, id)
}

// AllClusters returns all cluster instances
func (c *ClusterInstances) AllClusters() map[string]*ClusterInst {
	return c.clusters
}

// DefaultCluster returns a default ClusterInst instance.
// Returns nil when the clusters list is empty.
// First tries to return the instance with ID "InCluster",
// then tries to return the instance with ID "default".
// If neither exists, returns any instance from the clusters list.
func (c *ClusterInstances) DefaultCluster() *ClusterInst {
	// Check if clusters list is empty
	if len(c.clusters) == 0 {
		return nil
	}

	// Try to get cluster instance with ID "InCluster"
	id := "InCluster"
	cluster, exists := c.clusters[id]
	if exists {
		return cluster
	}

	// Try to get cluster instance with ID "default"
	id = "default"
	cluster, exists = c.clusters[id]
	if exists {
		return cluster
	}

	// If neither instance exists, return any instance from the clusters list
	for _, v := range c.clusters {
		return v
	}

	// Return nil if clusters list is empty (theoretically should have returned by now)
	return nil
}

// RegisterClusterAPIConfigs discovers and registers Cluster API managed clusters
func (c *ClusterInstances) RegisterClusterAPIConfigs() error {
	// Get the default cluster's dynamic client to discover CAPI clusters
	defaultCluster := c.DefaultCluster()
	if defaultCluster == nil {
		return fmt.Errorf("no default cluster available to discover CAPI clusters")
	}

	// Define supported CAPI versions
	capiVersions := []schema.GroupVersionResource{
		{
			Group:    "cluster.x-k8s.io",
			Version:  "v1beta1",
			Resource: "clusters",
		},
		{
			Group:    "cluster.x-k8s.io",
			Version:  "v1alpha3",
			Resource: "clusters",
		},
	}

	// Track if we found any clusters across all versions
	foundClusters := false

	// Iterate through each CAPI version
	for _, capiGVR := range capiVersions {
		// List all namespaces with CAPI clusters for this version
		clusters, err := defaultCluster.DynamicClient.Resource(capiGVR).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			klog.V(4).Infof("CAPI version %s not available: %v", capiGVR.Version, err)
			continue // Try next version if this one isn't available
		}

		// If we found at least one cluster in any version
		if len(clusters.Items) > 0 {
			foundClusters = true
		}

		for _, cluster := range clusters.Items {
			clusterName := cluster.GetName()
			namespace := cluster.GetNamespace()

			// Generate a unique cluster ID including the version
			clusterID := fmt.Sprintf("capi-%s-%s-%s", capiGVR.Version, namespace, clusterName)

			// Check if cluster is already registered
			if c.GetClusterById(clusterID) != nil {
				klog.V(4).Infof("Cluster %s in namespace %s (version %s) already registered", clusterName, namespace, capiGVR.Version)
				continue
			}

			// Try to get the kubeconfig secret
			secret, err := defaultCluster.Client.CoreV1().Secrets(namespace).Get(
				context.Background(),
				fmt.Sprintf("%s-kubeconfig", clusterName),
				metav1.GetOptions{},
			)
			if err != nil {
				klog.Errorf("Failed to get kubeconfig secret for cluster %s in namespace %s: %v", clusterName, namespace, err)
				continue
			}

			// Get kubeconfig data from secret
			kubeconfigData, ok := secret.Data["value"]
			if !ok {
				klog.Errorf("Kubeconfig secret for cluster %s in namespace %s has no 'value' key", clusterName, namespace)
				continue
			}

			// Register the cluster
			_, err = c.RegisterByStringWithID(string(kubeconfigData), clusterID)
			if err != nil {
				klog.Errorf("Failed to register cluster %s in namespace %s: %v", clusterName, namespace, err)
				continue
			}

			klog.Infof("Successfully registered CAPI cluster %s from namespace %s with ID %s (version %s)",
				clusterName, namespace, clusterID, capiGVR.Version)
		}
	}

	if !foundClusters {
		klog.V(4).Info("No Cluster API clusters found in any supported version")
	}

	return nil
}

// Show displays information about all clusters
func (c *ClusterInstances) Show() {
	klog.Infof("Show Clusters\n")
	for k, v := range c.clusters {
		if v.serverVersion == nil {
			klog.Infof("%s=nil\n", k)
			continue
		}
		klog.Infof("%s[%s,%s]=%s\n", k, v.serverVersion.Platform, v.serverVersion.GitVersion, v.Config.Host)
	}
}
