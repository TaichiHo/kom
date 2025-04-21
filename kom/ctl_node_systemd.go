package kom

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// getClusterNamespace finds the namespace of a cluster managed by Cluster API
func getClusterNamespace(clusterName string, dynamicClient dynamic.Interface, ctx context.Context) (string, error) {
	if clusterName == "" {
		return "", fmt.Errorf("cluster name is empty")
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

	// Try to find the cluster in each CAPI version
	for _, capiGVR := range capiVersions {
		// List all namespaces with CAPI clusters for this version
		clusters, err := dynamicClient.Resource(capiGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			klog.V(4).Infof("CAPI version %s not available: %v", capiGVR.Version, err)
			continue // Try next version if this one isn't available
		}

		// Look for our cluster
		for _, cluster := range clusters.Items {
			if cluster.GetName() == clusterName {
				return cluster.GetNamespace(), nil
			}
		}
	}

	return "", fmt.Errorf("failed to find namespace for cluster %s", clusterName)
}

// getSSHKeyFromSecret gets the SSH key from the admin cluster's secret
func (d *node) getSSHKeyFromSecret() (string, error) {
	// Get the cluster name from the kubectl ID
	clusterName := d.kubectl.ID
	if clusterName == "" {
		return "", fmt.Errorf("cluster name is empty")
	}

	// Get in-cluster config for admin cluster access
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get in-cluster config: %v", err)
	}

	// Create a client for the admin cluster
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create admin cluster client: %v", err)
	}

	// Create a dynamic client for the admin cluster
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create admin cluster dynamic client: %v", err)
	}

	// Find the cluster's namespace
	clusterNamespace, err := getClusterNamespace(clusterName, dynamicClient, d.kubectl.Statement.Context)
	if err != nil {
		return "", err
	}

	// Get the secret name
	secretName := fmt.Sprintf("%s-ssh-key", clusterName)

	// Get the secret from the admin cluster in the cluster's namespace
	secret, err := client.CoreV1().Secrets(clusterNamespace).Get(d.kubectl.Statement.Context, secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get SSH key secret from admin cluster in namespace %s: %v", clusterNamespace, err)
	}

	// Get the SSH key content
	sshKeyData, ok := secret.Data["ssh-privatekey"]
	if !ok {
		return "", fmt.Errorf("SSH key not found in secret")
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "ssh-key-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tmpFile.Close()

	// Write the SSH key to the temporary file
	if _, err := tmpFile.Write(sshKeyData); err != nil {
		return "", fmt.Errorf("failed to write SSH key to temporary file: %v", err)
	}

	// Set the correct permissions
	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		return "", fmt.Errorf("failed to set SSH key file permissions: %v", err)
	}

	return tmpFile.Name(), nil
}

// SystemdServiceStatus gets the status of a systemd service on a node
func (d *node) SystemdServiceStatus(serviceName string) (string, error) {
	// Get SSH key from secret
	sshKeyFile, err := d.getSSHKeyFromSecret()
	if err != nil {
		return "", err
	}
	defer os.Remove(sshKeyFile) // Clean up the temporary file

	// Get node IP
	node, err := d.getNodeWithCache(d.getCacheTTL())
	if err != nil {
		return "", fmt.Errorf("failed to get node info: %v", err)
	}

	// Find node IP
	var nodeIP string
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			nodeIP = addr.Address
			break
		}
	}
	if nodeIP == "" {
		return "", fmt.Errorf("failed to find node IP")
	}

	// Execute SSH command
	cmd := exec.Command("ssh", "-i", sshKeyFile, "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", nodeIP), "systemctl", "status", serviceName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute SSH command: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

// RestartSystemdService restarts a systemd service on a node
func (d *node) RestartSystemdService(serviceName string) error {
	// Get SSH key from secret
	sshKeyFile, err := d.getSSHKeyFromSecret()
	if err != nil {
		return err
	}
	defer os.Remove(sshKeyFile) // Clean up the temporary file

	// Get node IP
	node, err := d.getNodeWithCache(d.getCacheTTL())
	if err != nil {
		return fmt.Errorf("failed to get node info: %v", err)
	}

	// Find node IP
	var nodeIP string
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			nodeIP = addr.Address
			break
		}
	}
	if nodeIP == "" {
		return fmt.Errorf("failed to find node IP")
	}

	// Execute SSH command
	cmd := exec.Command("ssh", "-i", sshKeyFile, "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", nodeIP), "systemctl", "restart", serviceName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute SSH command: %v, output: %s", err, string(output))
	}

	// Check if restart was successful
	if strings.Contains(string(output), "Failed to restart") {
		return fmt.Errorf("service restart failed: %s", string(output))
	}

	return nil
}

// JournalLogs gets the journal logs for a systemd service on a node
func (d *node) JournalLogs(serviceName string, lines int) (string, error) {
	// Get SSH key from secret
	sshKeyFile, err := d.getSSHKeyFromSecret()
	if err != nil {
		return "", err
	}
	defer os.Remove(sshKeyFile) // Clean up the temporary file

	// Get node IP
	node, err := d.getNodeWithCache(d.getCacheTTL())
	if err != nil {
		return "", fmt.Errorf("failed to get node info: %v", err)
	}

	// Find node IP
	var nodeIP string
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			nodeIP = addr.Address
			break
		}
	}
	if nodeIP == "" {
		return "", fmt.Errorf("failed to find node IP")
	}

	// Execute SSH command
	cmd := exec.Command("ssh", "-i", sshKeyFile, "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("root@%s", nodeIP), "journalctl", "-u", serviceName, "-n", fmt.Sprintf("%d", lines))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute SSH command: %v, output: %s", err, string(output))
	}

	return string(output), nil
}
