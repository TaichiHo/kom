package kom

import (
	"fmt"
	"slices"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
)

// Get pod-related Services
func (p *pod) LinkedService() ([]*v1.Service, error) {
	// Search process
	// Get target Pod details:

	// Get metadata.labels using Pod API.
	// Determine Pod's Namespace.
	// Get all Services in the Namespace:

	// Use kubectl get services -n {namespace} or call API /api/v1/namespaces/{namespace}/services.
	// Match each Service's selector:

	// For each Service:
	// Extract its spec.selector.
	// Check if Pod has all these labels with matching values.
	// If all label conditions are met, record this Service as associated with the Pod.
	// Return results:

	// Return all matching Service names and related information.
	var pod *v1.Pod
	err := p.kubectl.WithCache(p.kubectl.Statement.CacheTTL).Get(&pod).Error
	if err != nil {
		return nil, fmt.Errorf("get pod %s/%s error %v", p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, err.Error())
	}
	if pod == nil {
		return nil, fmt.Errorf("get pod %s/%s error %v", p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, "pod is nil")
	}
	podLabels := pod.GetLabels()

	if len(podLabels) == 0 {
		return nil, nil
	}

	var services []*v1.Service
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.Service{}).
		Namespace(p.kubectl.Statement.Namespace).
		RemoveManagedFields().
		List(&services).Error

	if err != nil {
		return nil, fmt.Errorf("get service error %v", err.Error())
	}

	var result []*v1.Service
	for _, svc := range services {
		serviceLabels := svc.Spec.Selector
		// If empty, indicates no specific pod selector, skip this svc
		if len(serviceLabels) == 0 {
			continue
		}
		// Iterate through selector
		// All kv pairs in serviceLabels must exist in podLabels with matching values
		if utils.CompareMapContains(serviceLabels, podLabels) {
			result = append(result, svc)
		}
	}
	return result, nil
}

func (p *pod) LinkedEndpoints() ([]*v1.Endpoints, error) {

	services, err := p.LinkedService()
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, nil
	}
	// endpoints 与 svc 同名
	// 1.获取service 名称
	// 2.获取endpoints
	// 3.返回endpoints

	var names []string
	for _, svc := range services {
		names = append(names, svc.Name)
	}

	var endpoints []*v1.Endpoints

	err = p.kubectl.newInstance().
		WithContext(p.kubectl.Statement.Context).
		Resource(&v1.Endpoints{}).
		Namespace(p.kubectl.Statement.Namespace).
		Where("metadata.name in " + utils.StringListToSQLIn(names)).
		RemoveManagedFields().
		List(&endpoints).Error
	if err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (p *pod) LinkedPVC() ([]*v1.PersistentVolumeClaim, error) {

	var pod v1.Pod
	err := p.kubectl.Get(&pod).Error
	if err != nil {
		return nil, err
	}
	// 找打pvc 名称列表
	var pvcNames []string
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcNames = append(pvcNames, volume.PersistentVolumeClaim.ClaimName)
		}
	}

	if len(pvcNames) == 0 {
		return nil, nil
	}

	// 找出同ns下pvc的列表，过滤pvcNames
	var pvcList []*v1.PersistentVolumeClaim
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.PersistentVolumeClaim{}).
		Namespace(p.kubectl.Statement.Namespace).
		Where("metadata.name in " + utils.StringListToSQLIn(pvcNames)).
		RemoveManagedFields().
		List(&pvcList).Error
	if err != nil {
		return nil, err
	}

	for i := range pvcList {
		pvc := pvcList[i]
		var pvcMounts []*PodMount

		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvc.Name {

				for _, container := range pod.Spec.Containers {
					for _, volumeMount := range container.VolumeMounts {
						if volumeMount.Name == volume.Name {
							pm := PodMount{
								Name:      volume.PersistentVolumeClaim.ClaimName,
								MountPath: volumeMount.MountPath,
								SubPath:   volumeMount.SubPath,
								ReadOnly:  volumeMount.ReadOnly,
							}
							pvcMounts = append(pvcMounts, &pm)
						}
					}
				}

			}
		}

		if len(pvcMounts) > 0 {
			if pvc.Annotations == nil {
				pvc.Annotations = make(map[string]string)
			}
			pvc.Annotations["pvcMounts"] = utils.ToJSON(pvcMounts)
		}
	}

	return pvcList, nil
}
func (p *pod) LinkedPV() ([]*v1.PersistentVolume, error) {

	pvcList, err := p.LinkedPVC()
	if err != nil {
		return nil, err
	}
	if len(pvcList) == 0 {
		return nil, nil
	}
	var pvNames []string
	for _, pvc := range pvcList {
		pvNames = append(pvNames, pvc.Spec.VolumeName)
	}
	// 找出同ns下pvc的列表，过滤pvcNames
	var pvList []*v1.PersistentVolume
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.PersistentVolume{}).
		Namespace(p.kubectl.Statement.Namespace).
		Where("metadata.name in " + utils.StringListToSQLIn(pvNames)).
		RemoveManagedFields().
		List(&pvList).Error
	if err != nil {
		return nil, err
	}
	return pvList, nil
}

func (p *pod) LinkedIngress() ([]*networkingv1.Ingress, error) {

	var pod v1.Pod
	err := p.kubectl.Get(&pod).Error
	if err != nil {
		return nil, err
	}
	services, err := p.LinkedService()
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	var servicesName []string
	for _, svc := range services {
		servicesName = append(servicesName, svc.Name)
	}

	// 获取ingress
	// Ingress 通过 spec.rules 或 spec.defaultBackend 中的 service.name 指定关联的 Service。
	// 遍历services，获取ingress
	var ingressList []networkingv1.Ingress
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&networkingv1.Ingress{}).
		Namespace(p.kubectl.Statement.Namespace).
		WithCache(p.kubectl.Statement.CacheTTL).
		RemoveManagedFields().
		List(&ingressList).Error
	if err != nil {
		return nil, err
	}

	// 过滤ingressList，只保留与services关联的ingress
	var result []*networkingv1.Ingress
	for _, ingress := range ingressList {
		if slices.Contains(servicesName, ingress.Spec.Rules[0].Host) {
			result = append(result, &ingress)
		}
	}
	// 遍历 Ingress 检查关联
	for _, ingress := range ingressList {
		if ingress.Spec.DefaultBackend != nil {
			if ingress.Spec.DefaultBackend.Service != nil && ingress.Spec.DefaultBackend.Service.Name != "" {
				if slices.Contains(servicesName, ingress.Spec.DefaultBackend.Service.Name) {
					result = append(result, &ingress)
				}
			}
		}

		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil && path.Backend.Service.Name != "" {

						backName := path.Backend.Service.Name
						if slices.Contains(servicesName, backName) {
							result = append(result, &ingress)
						}
					}
				}

			}

		}
	}

	return result, nil
}

// PodMount contains Pod mount information
// Mount types include: configMap, secret
type PodMount struct {
	Name      string `json:"name,omitempty"`
	MountPath string `json:"mountPath,omitempty"`
	SubPath   string `json:"subPath,omitempty"`
	Mode      *int32 `json:"mode,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

// LinkedConfigMap gets Pod-related ConfigMaps
func (p *pod) LinkedConfigMap() ([]*v1.ConfigMap, error) {
	var item *v1.Pod
	err := p.kubectl.Get(&item).Error
	if err != nil {
		return nil, err
	}
	// Find configmap name list
	var configMapNames []string
	for _, volume := range item.Spec.Volumes {
		if volume.ConfigMap != nil {
			configMapNames = append(configMapNames, volume.ConfigMap.Name)
		}
	}
	if len(configMapNames) == 0 {
		return nil, nil
	}
	// Find configmap list in the same namespace, filter by configMapNames
	var configMapList []*v1.ConfigMap
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.ConfigMap{}).
		Namespace(p.kubectl.Statement.Namespace).
		RemoveManagedFields().
		Where("metadata.name in " + utils.StringListToSQLIn(configMapNames)).
		List(&configMapList).Error
	if err != nil {
		return nil, err
	}

	// item.Spec.Containers.volumeMounts
	// item.Spec.Volumes
	// Through iterating secretNames, we can find volumeName in pod.Spec.Volumes.
	// Through volumeName, we can find volumeMounts in pod.Spec.Containers.volumeMounts, extract mode
	// Extract mountPath, subPath from volumeMounts

	for i := range configMapList {
		configMap := configMapList[i]
		var configMapMounts []*PodMount

		configMapName := configMap.Name
		for _, volume := range item.Spec.Volumes {
			if volume.ConfigMap != nil && volume.ConfigMap.Name == configMapName {

				for _, container := range item.Spec.Containers {
					for _, volumeMount := range container.VolumeMounts {
						if volumeMount.Name == volume.Name {
							cm := PodMount{
								Name:      volume.ConfigMap.Name,
								MountPath: volumeMount.MountPath,
								SubPath:   volumeMount.SubPath,
								ReadOnly:  volumeMount.ReadOnly,
								Mode:      volume.ConfigMap.DefaultMode,
							}
							configMapMounts = append(configMapMounts, &cm)
						}
					}
				}

			}
		}

		if len(configMapMounts) > 0 {
			if configMap.Annotations == nil {
				configMap.Annotations = make(map[string]string)
			}
			configMap.Annotations["configMapMounts"] = utils.ToJSON(configMapMounts)
		}
	}

	return configMapList, nil
}

// LinkedSecret gets Pod-related Secrets
func (p *pod) LinkedSecret() ([]*v1.Secret, error) {
	var item *v1.Pod
	err := p.kubectl.Get(&item).Error
	if err != nil {
		return nil, err
	}
	// Find secret name list
	var secretNames []string
	for _, volume := range item.Spec.Volumes {
		if volume.Secret != nil {
			secretNames = append(secretNames, volume.Secret.SecretName)
		}
	}
	if len(secretNames) == 0 {
		return nil, nil
	}
	// Find secret list in the same namespace, filter by secretNames
	var secretList []*v1.Secret
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.Secret{}).
		Namespace(p.kubectl.Statement.Namespace).
		RemoveManagedFields().
		Where("metadata.name in " + utils.StringListToSQLIn(secretNames)).
		List(&secretList).Error
	if err != nil {
		return nil, err
	}

	// item.Spec.Containers.volumeMounts
	// item.Spec.Volumes
	// Through iterating secretNames, we can find volumeName in pod.Spec.Volumes.
	// Through volumeName, we can find volumeMounts in pod.Spec.Containers.volumeMounts, extract mode
	// Extract mountPath, subPath from volumeMounts

	for i := range secretList {
		secret := secretList[i]
		var secretMounts []*PodMount

		secretName := secret.Name
		for _, volume := range item.Spec.Volumes {
			if volume.Secret != nil && volume.Secret.SecretName == secretName {

				for _, container := range item.Spec.Containers {
					for _, volumeMount := range container.VolumeMounts {
						if volumeMount.Name == volume.Name {
							sm := PodMount{
								Name:      volume.Secret.SecretName,
								MountPath: volumeMount.MountPath,
								SubPath:   volumeMount.SubPath,
								ReadOnly:  volumeMount.ReadOnly,
								Mode:      volume.Secret.DefaultMode,
							}
							secretMounts = append(secretMounts, &sm)
						}
					}
				}

			}
		}

		if len(secretMounts) > 0 {
			if secret.Annotations == nil {
				secret.Annotations = make(map[string]string)
			}
			secret.Annotations["secretMounts"] = utils.ToJSON(secretMounts)
		}
	}

	return secretList, nil
}

// Env contains three values per line: container name, ENV name, ENV value
type Env struct {
	ContainerName string `json:"containerName,omitempty"`
	EnvName       string `json:"envName,omitempty"`
	EnvValue      string `json:"envValue,omitempty"`
}

func (p *pod) LinkedEnv() ([]*Env, error) {
	// 先获取容器列表，然后获取容器的环境变量，然后组装到Env结构体中

	// 先获取pod，从pod中读取容器列表
	var item *v1.Pod
	err := p.kubectl.Get(&item).Error
	if err != nil {
		return nil, err
	}

	var envs []*Env

	// 获取容器名称列表
	for _, container := range item.Spec.Containers {

		// 进到容器中执行ENV命令，获取输出字符串
		var result []byte
		err = p.kubectl.newInstance().Resource(&v1.Pod{}).
			WithContext(p.kubectl.Statement.Context).
			Namespace(p.kubectl.Statement.Namespace).
			Name(p.kubectl.Statement.Name).Ctl().Pod().
			ContainerName(container.Name).
			Command("env").
			Execute(&result).Error
		if err != nil {
			klog.V(6).Infof("get %s/%s/%s env error %v", p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, container.Name, err.Error())
			return nil, err
		}

		// 解析result，获取ENV名称和ENV值
		envArrays := strings.Split(string(result), "\n")
		for _, envline := range envArrays {
			envArray := strings.Split(envline, "=")
			if len(envArray) != 2 {
				continue
			}
			envs = append(envs, &Env{ContainerName: container.Name, EnvName: envArray[0], EnvValue: envArray[1]})
		}
	}

	return envs, nil
}

// LinkedEnvFromPod extracts env definitions from pod definition
func (p *pod) LinkedEnvFromPod() ([]*Env, error) {
	// First get pod, read container list from pod
	var pod v1.Pod
	err := p.kubectl.Get(&pod).Error
	if err != nil {
		return nil, err
	}
	var envs []*Env
	for _, container := range pod.Spec.Containers {

		for _, env := range container.Env {

			envHolder := &Env{ContainerName: container.Name, EnvName: env.Name, EnvValue: env.Value}
			if envHolder.EnvValue != "" {
				envs = append(envs, envHolder)
				continue
			}

			// ref has multiple cases that need to be checked
			// FieldRef\ResourceFieldRef\ConfigMapKeyRef\SecretKeyRef
			// Get values for these four cases, should be one of the four
			// Get value of env.ValueFrom.FieldRef.FieldPath
			if env.ValueFrom != nil && env.ValueFrom.FieldRef != nil && env.ValueFrom.FieldRef.FieldPath != "" {
				envHolder.EnvValue = fmt.Sprintf("[Field] %s", env.ValueFrom.FieldRef.FieldPath)
			}

			// 		 - name: CPU_REQUEST
			//     valueFrom:
			//       resourceFieldRef:
			//         containerName: multi-env-container
			//         resource: requests.cpu
			//   - name: MEMORY_LIMIT
			//     valueFrom:
			//       resourceFieldRef:
			//         containerName: multi-env-container
			//         resource: limits.memory
			if env.ValueFrom != nil && env.ValueFrom.ResourceFieldRef != nil && env.ValueFrom.ResourceFieldRef.Resource != "" {
				envHolder.EnvValue = fmt.Sprintf("[Container] %s/%s", env.ValueFrom.ResourceFieldRef.ContainerName, env.ValueFrom.ResourceFieldRef.Resource)
			}

			// configMapKeyRef:
			// name: my-env-configmap
			// key: env.list
			if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Key != "" {
				envHolder.EnvValue = fmt.Sprintf("[ConfigMap] %s/%s", env.ValueFrom.ConfigMapKeyRef.Name, env.ValueFrom.ConfigMapKeyRef.Key)
			}

			// secretKeyRef:
			// name: db-credentials
			// key: DB_PASSWORD
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Key != "" {
				envHolder.EnvValue = fmt.Sprintf("[Secret] %s/%s", env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key)
			}
			envs = append(envs, envHolder)

		}

		for _, envFrom := range container.EnvFrom {

			if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name != "" {
				envs = append(envs,
					&Env{
						ContainerName: container.Name,
						EnvName:       envFrom.ConfigMapRef.Name,
						EnvValue:      fmt.Sprintf("[ConfigMap] %s", envFrom.ConfigMapRef.Name),
					},
				)
			}
			if envFrom.SecretRef != nil && envFrom.SecretRef.Name != "" {
				envs = append(envs,
					&Env{
						ContainerName: container.Name,
						EnvName:       envFrom.SecretRef.Name,
						EnvValue:      fmt.Sprintf("[Secret] %s", envFrom.SecretRef.Name),
					},
				)
			}
		}
	}
	return envs, nil
}

type SelectedNode struct {
	Reason  string `json:"reason,omitempty"`    // Selection type: NodeSelector/NodeAffinity/Tolerations/NodeName
	Name    string `json:"node_name,omitempty"` // Node name
	Current bool   `json:"current,omitempty"`   // Whether it is the current node
}

// LinkedNode schedulable hosts
// Currently does not support host filtering based on CPU and memory resource constraints
func (p *pod) LinkedNode() ([]*SelectedNode, error) {

	var selectedNodeList []*SelectedNode

	var item *v1.Pod
	err := p.kubectl.Get(&item).Error
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("")
	}
	var nodeList []*v1.Node
	err = p.kubectl.newInstance().Resource(&v1.Node{}).
		List(&nodeList).Error
	if err != nil {
		return nil, err
	}

	// 1. NodeSelector
	// This configuration means Pod can only be scheduled to nodes with label disktype=ssd.
	// All labels in NodeSelector must be satisfied on the Node
	if item.Spec.NodeSelector != nil {
		nodeList = slice.Filter(nodeList, func(index int, n *v1.Node) bool {
			labels := n.Labels

			if utils.CompareMapContains(item.Spec.NodeSelector, labels) {
				selectedNodeList = append(selectedNodeList, &SelectedNode{
					Reason: "NodeSelector",
					Name:   n.Name,
				})
				return true
			}
			return false
		})
	}

	// 2. nodeAffinity
	// requiredDuringSchedulingIgnoredDuringExecution
	if item.Spec.Affinity != nil && item.Spec.Affinity.NodeAffinity != nil && item.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		terms := item.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		for _, term := range terms {
			if term.MatchExpressions != nil && len(term.MatchExpressions) > 0 {

				for _, exp := range term.MatchExpressions {
					nodeList = slice.Filter(nodeList, func(index int, n *v1.Node) bool {
						labels := n.Labels

						if utils.MatchNodeSelectorRequirement(labels, exp) {
							for _, selectedNode := range selectedNodeList {
								if selectedNode.Name == n.Name {
									return false
								}
							}
							selectedNodeList = append(selectedNodeList, &SelectedNode{
								Reason: "NodeAffinity",
								Name:   n.Name,
							})
							return true
						}
						return false
					})
				}

			}
		}

	}

	// Taints and tolerations
	// Only one toleration needs to be satisfied.
	// If node has taints, need to check
	// If node has no taints, no need to check
	if item.Spec.Tolerations != nil && len(item.Spec.Tolerations) > 0 {

		nodeList = slice.Filter(nodeList, func(index int, n *v1.Node) bool {
			// If node has no taints, no need to check
			if n.Spec.Taints == nil || len(n.Spec.Taints) == 0 {
				return true
			}

			for _, t := range n.Spec.Taints {
				if isTaintTolerated(t, item.Spec.Tolerations) {
					for _, selectedNode := range selectedNodeList {
						if selectedNode.Name == n.Name {
							return false
						}
					}
					selectedNodeList = append(selectedNodeList, &SelectedNode{
						Reason: "Tolerations",
						Name:   n.Name,
					})
					return true
				}
			}
			return false
		})
	}
	// Finally, if nodeName is configured, only this one is valid
	if item.Spec.NodeName != "" {
		nodeList = slice.Filter(nodeList, func(index int, n *v1.Node) bool {
			if n.Name == item.Spec.NodeName {

				// Check if it existed before, if yes, don't add again
				// Because once matched, it needs to be filled in pod.spec.NodeName
				// Cannot distinguish if scheduling succeeded due to nodeSelector or nodeAffinity
				for _, selectedNode := range selectedNodeList {
					if selectedNode.Name == item.Spec.NodeName {
						return false
					}
				}
				// First time
				selectedNodeList = append(selectedNodeList, &SelectedNode{
					Reason: "NodeName",
					Name:   n.Name,
				})
				return true
			}
			return false
		})

	}

	// Set whether it is the current node
	for _, selectedNode := range selectedNodeList {
		if selectedNode.Name == item.Spec.NodeName {
			selectedNode.Current = true
		}
	}

	return selectedNodeList, nil
}

// matchTaintAndToleration checks if a single taint matches a single toleration.
func matchTaintAndToleration(taint v1.Taint, toleration v1.Toleration) bool {
	// Check Effect (must match or be empty in toleration)
	if toleration.Effect != "" && toleration.Effect != taint.Effect {
		return false
	}

	// Check Operator and Key
	switch toleration.Operator {
	case "Equal":
		// Key and Value must match
		if toleration.Key != taint.Key {
			return false
		}
		if toleration.Value != taint.Value {
			return false
		}
	case "Exists":
		// Only Key needs to match
		if toleration.Key != "" && toleration.Key != taint.Key {
			return false
		}
	default:
		// Invalid operator
		return false
	}

	return true
}

// isTaintTolerated checks if a taint on a node is tolerated by any toleration of a pod.
func isTaintTolerated(taint v1.Taint, tolerations []v1.Toleration) bool {
	for _, toleration := range tolerations {
		if matchTaintAndToleration(taint, toleration) {
			return true
		}
	}
	return false
}
