package kom

import (
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/random"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/kom/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type node struct {
	kubectl *Kubectl
}

// Cordon marks the node as unschedulable.
// The core functionality of the cordon command is to mark a node as Unschedulable.
// In this state, the scheduler will not assign new pods to this node.
func (d *node) Cordon() error {
	var item interface{}
	patchData := `{"spec":{"unschedulable":true}}`
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return err
}

// UnCordon node
// The uncordon command is the reverse operation of cordon,
// used to restore a node from unschedulable state to schedulable state.
func (d *node) UnCordon() error {
	var item interface{}
	patchData := `{"spec":{"unschedulable":null}}`
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return err
}

// Taint adds a taint to a node to indicate that certain pods should not run on it.
// The syntax for the taint command is: kubectl taint node <node-name> <key>=<value>:<effect>
// where <key>, <value>, and <effect> represent the taint's key, value, and scope respectively.
// The effect can be NoSchedule, PreferNoSchedule, or NoExecute.
// The taint command adds a key-value pair named 'taints' to the node's metadata.annotations field.
// Example:
// Taint("dedicated2=special-user:NoSchedule")
// Taint("dedicated2:NoSchedule")
func (d *node) Taint(str string) error {
	taint, err := parseTaint(str)
	if err != nil {
		return err
	}
	var original *corev1.Node
	err = d.kubectl.Get(&original).Error
	if err != nil {
		return err
	}
	taints := original.Spec.Taints
	if taints == nil || len(taints) == 0 {
		taints = []corev1.Taint{*taint}
	} else {
		taints = append(taints, *taint)
	}

	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"taints":%s}}`, utils.ToJSON(taints))
	err = d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return err
}

// UnTaint removes a taint from the node
func (d *node) UnTaint(str string) error {
	taint, err := parseTaint(str)
	if err != nil {
		return err
	}
	var original *corev1.Node
	err = d.kubectl.Get(&original).Error
	if err != nil {
		return err
	}
	taints := original.Spec.Taints
	if taints == nil || len(taints) == 0 {
		return fmt.Errorf("taint %s not found", str)
	}

	taints = slice.Filter(taints, func(index int, item corev1.Taint) bool {
		return item.Key != taint.Key
	})
	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"taints":%s}}`, utils.ToJSON(taints))
	err = d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return err
}

// AllNodeLabels gets all labels from all nodes
func (d *node) AllNodeLabels() (map[string]string, error) {
	var list []*corev1.Node
	err := d.kubectl.newInstance().Resource(&corev1.Node{}).WithCache(d.kubectl.Statement.CacheTTL).
		List(&list).Error
	if err != nil {
		return nil, err
	}
	var labels map[string]string
	for _, n := range list {
		if len(n.Labels) > 0 {
			labels = maputil.Merge(labels, n.Labels)
		}
	}
	return labels, nil
}

// Drain node
// Drain is typically used when a node needs maintenance.
// It not only marks the node as unschedulable but also evicts all pods from the node one by one.
func (d *node) Drain() error {
	// TODO: Add handling for --force flag, which forces eviction of all pods even if they don't satisfy PDB
	name := d.kubectl.Statement.Name

	// Step 1: Mark the node as unschedulable
	klog.V(8).Infof("node/%s  cordoned\n", name)
	err := d.Cordon()
	if err != nil {
		klog.V(8).Infof("node/%s  cordon error %v\n", name, err.Error())
		return err
	}

	// Step 2: Get all pods on the node
	var podList []*corev1.Pod
	err = d.kubectl.newInstance().Resource(&corev1.Pod{}).
		WithFieldSelector(fmt.Sprintf("spec.nodeName=%s", name)).
		List(&podList).Error
	if err != nil {
		klog.V(8).Infof("list pods in node/%s  error %v\n", name, err.Error())
		return err
	}

	// Step 3: Evict all evictable pods
	for _, pod := range podList {
		if isDaemonSetPod(pod) || isMirrorPod(pod) {
			// Ignore DaemonSet and Mirror Pods
			klog.V(8).Infof("ignore evict pod  %s/%s  \n", pod.Namespace, pod.Name)
			continue
		}
		klog.V(8).Infof("pod/%s eviction started", pod.Name)

		// Evict Pod
		err := d.evictPod(pod)
		if err != nil {
			klog.V(8).Infof("failed to evict pod %s: %v", pod.Name, err)
			return fmt.Errorf("failed to evict pod %s: %v", pod.Name, err)
		}
		klog.V(8).Infof("pod/%s evictied", pod.Name)
	}

	// Step 4: Wait for all pods to be evicted
	err = wait.PollImmediate(2*time.Second, 5*time.Minute, func() (bool, error) {
		var podList []*corev1.Pod
		err = d.kubectl.newInstance().Resource(&corev1.Pod{}).
			WithFieldSelector(fmt.Sprintf("spec.nodeName=%s", name)).
			List(&podList).Error
		if err != nil {
			klog.V(8).Infof("list pods in node/%s  error %v\n", name, err.Error())
			return false, err
		}
		for _, pod := range podList {
			if isDaemonSetPod(pod) || isMirrorPod(pod) {
				// Ignore DaemonSet and Mirror Pods
				klog.V(8).Infof("ignore evict pod  %s/%s  \n", pod.Namespace, pod.Name)
				continue
			}
			klog.V(8).Infof("pod/%s eviction started", pod.Name)

			// Evict Pod
			err := d.evictPod(pod)
			if err != nil {
				return false, fmt.Errorf("failed to evict pod %s: %v", pod.Name, err)
			}
			klog.V(8).Infof("pod/%s evictied", pod.Name)
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("timeout waiting for pods to be evicted: %w", err)
	}

	klog.V(8).Infof("node/%s drained", name)
	return nil
}

// CreateNodeShell gets a node shell
// Requires nsenter to be present in the container
func (d *node) CreateNodeShell(image ...string) (namespace, podName, containerName string, err error) {
	// Get node
	runImage := "alpine:latest"
	if len(image) > 0 {
		runImage = image[0]
	}
	namespace = "kube-system"
	containerName = "shell"
	podName = fmt.Sprintf("node-shell-%s", strings.ToLower(random.RandString(8)))
	var yaml = `
apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
spec:
  containers:
  - args:
    - -t
    - "1"
    - -m
    - -u
    - -i
    - -n
    - sleep
    - "14000"
    command:
    - nsenter
    image: %s
    name: %s
    imagePullPolicy: IfNotPresent
    securityContext:
      privileged: true
  hostIPC: true
  hostNetwork: true
  hostPID: true
  restartPolicy: Never
  nodeName: %s	
  tolerations:
  - operator: Exists
`
	yaml = fmt.Sprintf(yaml, podName, namespace, runImage, containerName, d.kubectl.Statement.Name)

	ret := d.kubectl.Applier().Apply(yaml)
	// [Pod/node-shell-xqrbqqvt created]
	// Check if contains "created"
	klog.V(6).Infof("%s Node Shell creation result %s", d.kubectl.Statement.Name, ret)

	// Creation successful
	if len(ret) > 0 && strings.Contains(ret[0], "created") {
		// Wait for startup or timeout, use default timeout if not specified
		err = d.waitPodReady(namespace, podName, d.kubectl.Statement.CacheTTL)
		return
	}

	// Creation failed
	err = fmt.Errorf("node shell creation failed %s", ret)
	return
}
func (d *node) waitPodReady(ns, podName string, ttl time.Duration) error {
	var p *v1.Pod
	if ttl == 0 {
		// Set a default wait time
		ttl = 30
	}
	timeout := time.After(ttl * time.Second)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for Pod to start")
		case <-ticker.C:
			err := d.kubectl.newInstance().Resource(&v1.Pod{}).Name(podName).Namespace(ns).Get(&p).Error
			if err != nil {
				klog.V(6).Infof("waiting for Pod %s/%s to be created...", ns, podName)
				continue
			}

			if p == nil {
				klog.V(6).Infof("Pod %s/%s not created", ns, podName)
				continue
			}

			if len(p.Status.ContainerStatuses) == 0 {
				klog.V(6).Infof("Pod %s/%s container status not ready", ns, podName)
				continue
			}

			// Check if all containers are Ready
			allContainersReady := true
			for _, status := range p.Status.ContainerStatuses {
				if !status.Ready {
					allContainersReady = false
					klog.V(6).Infof("container %s in Pod %s/%s not ready", status.Name, ns, podName)
					break
				}
			}

			if allContainersReady {
				klog.V(6).Infof("all containers in Pod %s/%s are ready", ns, podName)
				break
			}
		}

		// If all containers are Ready, exit loop
		if p != nil && len(p.Status.ContainerStatuses) > 0 {
			allReady := true
			for _, status := range p.Status.ContainerStatuses {
				if !status.Ready {
					allReady = false
					break
				}
			}
			if allReady {
				break
			}
		}

		klog.V(6).Infof("continue waiting for Pod %s/%s to be fully ready...", ns, podName)
	}

	return nil
}

// CreateKubectlShell creates a shell for kubectl operations
// Requires nsenter to be present in the container
// CreateKubectlShell creates a Pod for running kubectl and passes in kubeconfig content
func (d *node) CreateKubectlShell(kubeconfig string, image ...string) (namespace, podName, containerName string, err error) {
	// Default kubectl image
	runImage := "bitnami/kubectl:latest"
	if len(image) > 0 {
		runImage = image[0]
	}

	namespace = "kube-system"
	containerName = "shell"
	podName = fmt.Sprintf("kubectl-shell-%s", strings.ToLower(random.RandString(8)))

	// Replace newlines in kubeconfig string with \n
	kubeconfigEscaped := strings.Replace(kubeconfig, "\n", `\n`, -1)
	// Use template string to create YAML configuration
	podTemplate := `
apiVersion: v1
kind: Pod
metadata:
  name: {{.PodName}}
  namespace: {{.Namespace}}
spec:
  initContainers:
  - name: init-container
    image: {{.RunImage}}
    imagePullPolicy: IfNotPresent
    command: ['sh', '-c', 'echo -e "{{.Kubeconfig}}" > /.kube/config || (echo "Failed to write kubeconfig" && exit 1)']
    volumeMounts:
     - name: kube-config
       mountPath: /.kube
  containers:
  - name: {{.ContainerName}}
    image: {{.RunImage}}
    command: ['tail', '-f', '/dev/null']
    env:
     - name: KUBECONFIG
       value: /.kube/config
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: kube-config
      mountPath: /.kube
  nodeName: {{.NodeName}}	
  tolerations:
  - operator: Exists	  
  volumes:
  - name: kube-config
    emptyDir: {}
`

	// 准备模板数据
	data := map[string]interface{}{
		"PodName":       podName,
		"Namespace":     namespace,
		"RunImage":      runImage,
		"Kubeconfig":    kubeconfigEscaped,
		"ContainerName": containerName,
		"NodeName":      d.kubectl.Statement.Name, // 使用 node 上的 kubectl name
	}

	// 创建模板并执行填充
	tmpl, err := template.New("podConfig").Parse(podTemplate)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse template: %w", err)
	}

	// 生成最终的 YAML 配置
	var yaml string
	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to execute template: %w", err)
	}
	yaml = buf.String()

	klog.V(6).Infof("Generated YAML:\n%s", yaml)
	// Call kubectl's Applier method to apply the generated YAML config
	ret := d.kubectl.Applier().Apply(yaml)

	// Check if creation was successful
	klog.V(6).Infof("%s kubectl Shell creation result %s", d.kubectl.Statement.Name, ret)

	// If the return result contains "created" string, consider creation successful
	if len(ret) > 0 && strings.Contains(ret[0], "created") {
		// Wait for startup or timeout, use default timeout if not specified
		err = d.waitPodReady(namespace, podName, d.kubectl.Statement.CacheTTL)

		return
	}

	// Creation failed
	err = fmt.Errorf("kubectl shell creation failed %s", ret)
	return
}

// Check if Pod is created by DaemonSet
func isDaemonSetPod(pod *corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

// Check if Pod is a Mirror Pod
func isMirrorPod(pod *corev1.Pod) bool {
	_, exists := pod.Annotations[corev1.MirrorPodAnnotationKey]
	return exists
}

// Evict Pod
func (d *node) evictPod(pod *corev1.Pod) error {
	klog.V(8).Infof("evicting pod %s/%s \n", pod.Namespace, pod.Name)
	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
	}
	err := d.kubectl.Client().PolicyV1().Evictions(pod.Namespace).Evict(context.TODO(), eviction)

	// err := d.kubectl.newInstance().Resource(eviction).Create(eviction).Error
	if err != nil {
		return err
	}
	klog.V(8).Infof(" pod %s/%s evicted\n", pod.Namespace, pod.Name)
	return nil
}

// ParseTaint parses a taint string into a corev1.Taint structure.
func parseTaint(taintStr string) (*corev1.Taint, error) {
	// Split the input string into key-value-effect
	var key, value, effect string
	parts := strings.Split(taintStr, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid taint format: %s", taintStr)
	}
	keyValue := parts[0]
	effect = parts[1]

	// Check the effect
	if effect != string(corev1.TaintEffectNoSchedule) &&
		effect != string(corev1.TaintEffectPreferNoSchedule) &&
		effect != string(corev1.TaintEffectNoExecute) {
		return nil, fmt.Errorf("invalid taint effect: %s", effect)
	}

	// Parse the key and value
	keyValueParts := strings.SplitN(keyValue, "=", 2)
	key = keyValueParts[0]
	if len(keyValueParts) == 2 {
		value = keyValueParts[1]
	}

	// Return the Taint structure
	return &corev1.Taint{
		Key:    key,
		Value:  value,
		Effect: corev1.TaintEffect(effect),
	}, nil
}
func (d *node) RunningPods() ([]*corev1.Pod, error) {
	// Get all pods running on this node
	var podList []*corev1.Pod
	err := d.kubectl.newInstance().WithCache(d.kubectl.Statement.CacheTTL).Resource(&corev1.Pod{}).
		AllNamespace().
		Where(fmt.Sprintf("spec.nodeName='%s'", d.kubectl.Statement.Name)).
		List(&podList).Error
	return podList, err
}
func (d *node) TotalRequestsAndLimits() (map[corev1.ResourceName]resource.Quantity, map[corev1.ResourceName]resource.Quantity) {
	// Get all pods running on this node
	podList, _ := d.RunningPods()
	return getPodsTotalRequestsAndLimits(podList)
}

// ResourceUsage gets the node's resource usage, including resource requests and limits, and current usage percentage
func (d *node) ResourceUsage() *ResourceUsageResult {
	// Get node information
	node, err := d.getNodeWithCache(d.getCacheTTL())
	if err != nil {
		return nil
	}

	// Get all pods running on this node
	podList, _ := d.RunningPods()

	// Calculate total requests and limits
	reqs, limits := getPodsTotalRequestsAndLimits(podList)

	// Calculate usage ratios
	fractions := make(map[corev1.ResourceName]ResourceUsageFraction)
	for _, resourceName := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory, corev1.ResourceEphemeralStorage} {
		allocatable := node.Status.Allocatable[resourceName]
		if allocatable.IsZero() {
			continue
		}

		reqQuantity := reqs[resourceName]
		limitQuantity := limits[resourceName]
		allocatableQuantity := allocatable

		fraction := ResourceUsageFraction{
			RequestFraction: float64(reqQuantity.AsDec().UnscaledBig().Int64()) / float64(allocatableQuantity.AsDec().UnscaledBig().Int64()) * 100,
			LimitFraction:   float64(limitQuantity.AsDec().UnscaledBig().Int64()) / float64(allocatableQuantity.AsDec().UnscaledBig().Int64()) * 100,
		}
		fractions[resourceName] = fraction
	}

	return &ResourceUsageResult{
		Requests:       reqs,
		Limits:         limits,
		Allocatable:    node.Status.Allocatable,
		UsageFractions: fractions,
	}
}
func (d *node) ResourceUsageTable() []*ResourceUsageRow {
	result := d.ResourceUsage()
	tableData, err := convertToTableData(result)
	if err != nil {
		return nil
	}
	return tableData
}

// IPUsage calculates the node's IP count status, returns total node IPs, used count, and available count
func (d *node) IPUsage() (total, used, available int) {
	// Get node information
	node, err := d.getNodeWithCache(d.getCacheTTL())
	if err != nil {
		return
	}

	// Get all pods running on this node
	podList, _ := d.RunningPods()

	// Get maximum number of pods
	maxPods := node.Status.Allocatable[corev1.ResourcePods]
	total = int(maxPods.Value())

	// Get number of pods currently running
	used = len(podList)

	// Calculate remaining available pods
	available = total - used

	return
}

// getCacheTTL gets cache time
// Default 5 seconds
func (d *node) getCacheTTL(defaultCacheTime ...time.Duration) time.Duration {
	// If cache time is specified in Statement, use that value
	if d.kubectl.Statement.CacheTTL > 0 {
		return d.kubectl.Statement.CacheTTL
	}
	// If default cache time is provided, use that value
	if len(defaultCacheTime) > 0 {
		return defaultCacheTime[0]
	}
	// Otherwise, use 10 seconds as default
	return time.Second * 10
}

// PodCount calculates the number of Pods on the node and the node's Pod count limit
func (d *node) PodCount() (total, used, available int) {
	// Get node information
	node, err := d.getNodeWithCache(d.getCacheTTL())
	if err != nil {
		return
	}

	// Get all pods running on this node
	podList, _ := d.RunningPods()

	// Get maximum number of pods
	maxPods := node.Status.Allocatable[corev1.ResourcePods]
	total = int(maxPods.Value())

	// Get number of pods currently running
	used = len(podList)

	// Calculate remaining available pods
	available = total - used

	return
}

// getNodeWithCache gets node with cache
func (d *node) getNodeWithCache(cacheTime time.Duration) (*corev1.Node, error) {
	// Get node information from cache
	var node *corev1.Node
	err := d.kubectl.WithCache(cacheTime).Resource(&corev1.Node{}).Get(&node).Error
	return node, err
}
func getPodsTotalRequestsAndLimits(podList []*corev1.Pod) (reqs map[corev1.ResourceName]resource.Quantity, limits map[corev1.ResourceName]resource.Quantity) {
	reqs = map[corev1.ResourceName]resource.Quantity{}
	limits = map[corev1.ResourceName]resource.Quantity{}

	for _, pod := range podList {
		for _, container := range pod.Spec.Containers {
			// Add container requests to total
			for name, quantity := range container.Resources.Requests {
				if value, ok := reqs[name]; !ok {
					reqs[name] = quantity.DeepCopy()
				} else {
					value.Add(quantity)
					reqs[name] = value
				}
			}
			// Add container limits to total
			for name, quantity := range container.Resources.Limits {
				if value, ok := limits[name]; !ok {
					limits[name] = quantity.DeepCopy()
				} else {
					value.Add(quantity)
					limits[name] = value
				}
			}
		}
	}
	return
}
