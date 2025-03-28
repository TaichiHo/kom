package kom

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// Resource types supported by rollout
var rolloutSupportedKinds = []string{"Deployment", "StatefulSet", "DaemonSet", "ReplicaSet"}

type rollout struct {
	kubectl *Kubectl
}

func (d *rollout) logInfo(action string) {
	kind := d.kubectl.Statement.GVK.Kind
	resource := d.kubectl.Statement.GVR.Resource
	namespace := d.kubectl.Statement.Namespace
	name := d.kubectl.Statement.Name
	klog.V(8).Infof("%s Kind=%s", action, kind)
	klog.V(8).Infof("%s Resource=%s", action, resource)
	klog.V(8).Infof("%s %s/%s", action, namespace, name)
}
func (d *rollout) handleError(kind string, namespace string, name string, action string, err error) error {
	if err != nil {
		d.kubectl.Error = fmt.Errorf("%s %s/%s %s error %v", kind, namespace, name, action, err)
		return err
	}
	return nil
}
func (d *rollout) checkResourceKind(kind string, supportedKinds []string) error {
	if !isSupportedKind(kind, supportedKinds) {
		d.kubectl.Error = fmt.Errorf("%s %s/%s operation is not supported", kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name)
		return d.kubectl.Error
	}
	return nil
}

func (d *rollout) Restart() error {

	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Restart")

	if err := d.checkResourceKind(kind, rolloutSupportedKinds); err != nil {
		return err
	}

	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kom.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().Format(time.DateTime))
	err := d.kubectl.Patch(&item, types.StrategicMergePatchType, patchData).Error
	return d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "restarting", err)
}
func (d *rollout) Pause() error {
	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Pause")

	if err := d.checkResourceKind(kind, []string{"Deployment"}); err != nil {
		return err
	}

	var item interface{}
	patchData := `{"spec":{"paused":true}}`
	err := d.kubectl.Patch(&item, types.StrategicMergePatchType, patchData).Error
	return d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "pause", err)

}
func (d *rollout) Resume() error {
	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Resume")

	if err := d.checkResourceKind(kind, []string{"Deployment"}); err != nil {
		return err
	}

	var item interface{}
	patchData := `{"spec":{"paused":null}}`
	err := d.kubectl.Patch(&item, types.StrategicMergePatchType, patchData).Error
	return d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "resume", err)

}

// Status
// Common field extraction:
//
// spec.replicas: Desired number of replicas.
// status.replicas: Current number of running replicas.
// status.updatedReplicas: Number of replicas updated to latest version.
// status.readyReplicas: Number of replicas that passed health checks and are ready to serve.
// status.unavailableReplicas: Number of currently unavailable replicas.
// Deployment:
//
// Completion conditions: updatedReplicas == spec.replicas, readyReplicas == spec.replicas, and unavailableReplicas == 0.
// StatefulSet:
//
// Completion conditions: updatedReplicas == spec.replicas and readyReplicas == spec.replicas.
// DaemonSet:
//
// Specific fields:
// status.desiredNumberScheduled: Number of nodes that should run the daemon pod.
// status.updatedNumberScheduled: Number of nodes that are running the updated daemon pod.
// status.numberReady: Number of nodes that are running the daemon pod and are ready.
// status.numberUnavailable: Number of nodes that are running the daemon pod but are not ready.
// Completion conditions: updatedNumberScheduled == desiredNumberScheduled, numberReady == desiredNumberScheduled, and numberUnavailable == 0.
// ReplicaSet:
//
// Completion conditions: readyReplicas == spec.replicas.
// Return status:
//
// Returns success message when rollout is complete.
// Returns progress information when update is in progress.
func (d *rollout) Status() (string, error) {
	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Status")

	if err := d.checkResourceKind(kind, rolloutSupportedKinds); err != nil {
		return "", err
	}

	var item unstructured.Unstructured
	err := d.kubectl.Get(&item).Error
	if err != nil {
		return "", d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "status", err)
	}

	// Extract replicas configuration
	specReplicas, _, _ := unstructured.NestedInt64(item.Object, "spec", "replicas")
	updatedReplicas, _, _ := unstructured.NestedInt64(item.Object, "status", "updatedReplicas")
	readyReplicas, _, _ := unstructured.NestedInt64(item.Object, "status", "readyReplicas")
	unavailableReplicas, _, _ := unstructured.NestedInt64(item.Object, "status", "unavailableReplicas")

	switch kind {
	case "Deployment":
		// Check if Deployment rollout is complete
		if updatedReplicas == specReplicas && readyReplicas == specReplicas && unavailableReplicas == 0 {
			return "Deployment successfully rolled out", nil
		}
		return fmt.Sprintf("Deployment rollout in progress: %d of %d updated, %d ready", updatedReplicas, specReplicas, readyReplicas), nil

	case "StatefulSet":
		// Check if StatefulSet rollout is complete
		if updatedReplicas == specReplicas && readyReplicas == specReplicas {
			return "StatefulSet successfully rolled out", nil
		}
		return fmt.Sprintf("StatefulSet rollout in progress: %d of %d updated, %d ready", updatedReplicas, specReplicas, readyReplicas), nil

	case "DaemonSet":
		desiredNumberScheduled, _, _ := unstructured.NestedInt64(item.Object, "status", "desiredNumberScheduled")
		updatedNumberScheduled, _, _ := unstructured.NestedInt64(item.Object, "status", "updatedNumberScheduled")
		numberReady, _, _ := unstructured.NestedInt64(item.Object, "status", "numberReady")
		numberUnavailable, _, _ := unstructured.NestedInt64(item.Object, "status", "numberUnavailable")

		// Check if DaemonSet rollout is complete
		if updatedNumberScheduled == desiredNumberScheduled && numberReady == desiredNumberScheduled && numberUnavailable == 0 {
			return "DaemonSet successfully rolled out", nil
		}
		return fmt.Sprintf("DaemonSet rollout in progress: %d of %d updated, %d ready", updatedNumberScheduled, desiredNumberScheduled, numberReady), nil

	case "ReplicaSet":
		// Check if ReplicaSet rollout is complete
		if readyReplicas == specReplicas {
			return "ReplicaSet successfully rolled out", nil
		}
		return fmt.Sprintf("ReplicaSet rollout in progress: %d of %d ready", readyReplicas, specReplicas), nil

	default:
		return "", fmt.Errorf("unsupported kind: %s", kind)
	}
}

type RolloutHistory struct {
	Kind              string                  `json:"kind,omitempty"`
	Name              string                  `json:"name,omitempty"`
	Namespace         string                  `json:"namespace,omitempty"`
	Revision          string                  `json:"revision,omitempty"`
	CreationTimestamp metav1.Time             `json:"creationTimestamp"`
	GVK               schema.GroupVersionKind `json:"gvk,omitempty"`
	ExtraInfo         map[string]string       `json:"extraInfo,omitempty"`
	Containers        []ContainerInfo         `json:"containers,omitempty"`
}
type ContainerInfo struct {
	Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`
}

func (d *rollout) History() ([]RolloutHistory, error) {
	kind := d.kubectl.Statement.GVK.Kind
	name := d.kubectl.Statement.Name
	ns := d.kubectl.Statement.Namespace
	d.logInfo("History")

	// Check if the resource type is supported
	if err := d.checkResourceKind(kind, []string{"Deployment", "StatefulSet", "DaemonSet"}); err != nil {
		return nil, err
	}

	var item unstructured.Unstructured
	err := d.kubectl.Get(&item).Error
	if err != nil {
		return nil, d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "history", err)
	}

	switch kind {
	case "Deployment":
		// Get Deployment's spec.selector.matchLabels
		labels, found, err := unstructured.NestedMap(item.Object, "spec", "selector", "matchLabels")
		if err != nil || !found {
			return nil, fmt.Errorf("failed to get matchLabels from Deployment: %s/%s %v", ns, name, err)
		}

		// Construct labelSelector string by concatenating all labels
		labelSelector := ""
		for key, value := range labels {
			labelSelector += fmt.Sprintf("%s=%s,", key, value)
		}
		// Remove the last comma
		if len(labelSelector) > 0 {
			labelSelector = labelSelector[:len(labelSelector)-1]
		}

		// Query all ReplicaSets associated with the Deployment
		var rsList []*v1.ReplicaSet

		err = d.kubectl.newInstance().Resource(&v1.ReplicaSet{}).
			Namespace(ns).
			WithLabelSelector(labelSelector).List(&rsList).Error
		if err != nil {
			return nil, fmt.Errorf("failed to list ReplicaSets for Deployment %s/%s: %v", ns, name, err)
		}
		rsList = slice.Filter(rsList, func(index int, item *v1.ReplicaSet) bool {
			for _, owner := range item.OwnerReferences {
				if owner.Kind == kind && owner.Name == name {
					return true
				}
			}
			return false
		})
		// If no ReplicaSets found, there's no history
		if len(rsList) == 0 {
			return nil, fmt.Errorf("no history found for Deployment %s/%s", ns, name)
		}

		var historyEntries []RolloutHistory
		for _, rs := range rsList {
			revision := rs.Annotations["deployment.kubernetes.io/revision"]
			var containers []ContainerInfo
			if rs.Spec.Template.Spec.Containers != nil {
				cs := rs.Spec.Template.Spec.Containers
				for _, c := range cs {
					containers = append(containers, ContainerInfo{
						Name:  c.Name,
						Image: c.Image,
					})
				}
			}

			historyEntries = append(historyEntries, RolloutHistory{
				Kind:              "ReplicaSet",
				Name:              rs.GetName(),
				Namespace:         rs.GetNamespace(),
				GVK:               rs.GetObjectKind().GroupVersionKind(),
				CreationTimestamp: rs.GetCreationTimestamp(),
				Revision:          revision,
				Containers:        containers,
			})
		}
		return historyEntries, nil

	case "StatefulSet":

		var versionList []*v1.ControllerRevision
		err = d.kubectl.newInstance().Resource(&v1.ControllerRevision{}).
			Namespace(ns).
			List(&versionList).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get controllerrevisions for StatefulSet: %s/%s %v", ns, name, err)
		}

		versionList = d.filterByOwner(versionList, kind, name)

		if len(versionList) == 0 {
			return nil, fmt.Errorf("no history found for StatefulSet %s/%s", ns, name)
		}

		var historyEntries []RolloutHistory
		for _, rv := range versionList {
			var containers []ContainerInfo
			var stsTemplate v1.StatefulSet
			if err = json.Unmarshal(rv.Data.Raw, &stsTemplate); err == nil {
				if stsTemplate.Spec.Template.Spec.Containers != nil {
					cs := stsTemplate.Spec.Template.Spec.Containers
					for _, c := range cs {
						containers = append(containers, ContainerInfo{
							Name:  c.Name,
							Image: c.Image,
						})
					}
				}

			}
			historyEntries = append(historyEntries, RolloutHistory{
				Kind:              "ControllerRevision",
				Name:              rv.GetName(),
				Namespace:         rv.GetNamespace(),
				GVK:               rv.GetObjectKind().GroupVersionKind(),
				CreationTimestamp: rv.GetCreationTimestamp(),
				Revision:          fmt.Sprintf("%d", rv.Revision),
				Containers:        containers,
			})
		}
		return historyEntries, nil
	case "DaemonSet":

		var versionList []*v1.ControllerRevision
		err = d.kubectl.newInstance().Resource(&v1.ControllerRevision{}).
			Namespace(ns).
			List(&versionList).Error
		if err != nil {

			return nil, fmt.Errorf("failed to get controllerrevisions for DaemonSet: %s/%s %v", ns, name, err)
		}
		versionList = d.filterByOwner(versionList, kind, name)
		if len(versionList) == 0 {
			return nil, fmt.Errorf("no history found for DaemonSet %s/%s", ns, name)
		}

		var historyEntries []RolloutHistory
		for _, rv := range versionList {

			var containers []ContainerInfo
			var dsTemplate v1.DaemonSet
			if err = json.Unmarshal(rv.Data.Raw, &dsTemplate); err == nil {
				if dsTemplate.Spec.Template.Spec.Containers != nil {
					cs := dsTemplate.Spec.Template.Spec.Containers
					for _, c := range cs {
						containers = append(containers, ContainerInfo{
							Name:  c.Name,
							Image: c.Image,
						})
					}
				}

			}

			historyEntries = append(historyEntries, RolloutHistory{
				Kind:              "ControllerRevision",
				Name:              rv.GetName(),
				Namespace:         rv.GetNamespace(),
				CreationTimestamp: rv.GetCreationTimestamp(),
				GVK:               rv.GetObjectKind().GroupVersionKind(),
				Revision:          fmt.Sprintf("%d", rv.Revision),
				Containers:        containers,
			})
		}
		return historyEntries, nil

	default:
		return nil, fmt.Errorf("unsupported kind: %s", kind)
	}
}

func (d *rollout) filterByOwner(versionList []*v1.ControllerRevision, kind string, name string) []*v1.ControllerRevision {
	versionList = slice.Filter(versionList, func(index int, item *v1.ControllerRevision) bool {
		for _, owner := range item.OwnerReferences {
			if owner.Kind == kind && owner.Name == name {
				return true
			}
		}
		return false
	})
	return versionList
}
func (d *rollout) Undo(toVersions ...int) (string, error) {
	kind := d.kubectl.Statement.GVK.Kind
	name := d.kubectl.Statement.Name
	namespace := d.kubectl.Statement.Namespace
	toVersion := 0
	if len(toVersions) > 0 {
		toVersion = toVersions[0]
	}
	d.logInfo("Undo")

	// Check if the resource type is supported
	if err := d.checkResourceKind(kind, []string{"Deployment", "DaemonSet", "StatefulSet"}); err != nil {
		return "", err
	}

	var item unstructured.Unstructured
	err := d.kubectl.Get(&item).Error
	if err != nil {
		return "", d.handleError(kind, namespace, name, "Undo", err)
	}

	// Call different rollback methods based on resource type
	switch kind {
	case "Deployment":
		err = d.rollbackDeployment(toVersion)
	case "StatefulSet":
		err = d.rollbackStatefulSet(toVersion)
	case "DaemonSet":
		err = d.rollbackDaemonSet(toVersion)
	default:
		return "", fmt.Errorf("unsupported kind: %s", kind)
	}

	if err != nil {
		return "", d.handleError(kind, namespace, name, "Undo", err)
	}

	return fmt.Sprintf("%s/%s rolled back successfully", kind, name), nil
}

func (d *rollout) rollbackDeployment(toVersion int) error {
	kind := d.kubectl.Statement.GVK.Kind
	name := d.kubectl.Statement.Name
	ns := d.kubectl.Statement.Namespace
	var deploy v1.Deployment
	err := d.kubectl.Resource(&deploy).
		WithLabelSelector("app=" + name).
		Get(&deploy).Error
	if err != nil {
		return fmt.Errorf(" rollbackDeployment get deployment  err %v ", err)
	}

	if toVersion == 0 {
		// If version not specified, rollback to previous version
		revision, err := ExtractDeploymentRevision(deploy.Annotations)
		if err != nil {
			return fmt.Errorf(" rollbackDeployment get deployment revision err %v ", err)
		}
		toVersion = revision - 1
	}

	var rsList []v1.ReplicaSet
	err = d.kubectl.newInstance().Resource(&v1.ReplicaSet{}).
		Namespace(ns).
		List(&rsList).Error
	if err != nil {
		return fmt.Errorf(" rollbackDeployment get rs list err %v ", err)
	}
	var vrs *v1.ReplicaSet
	for _, rs := range rsList {
		owners := rs.OwnerReferences
		if owners != nil && len(owners) > 0 {
			for _, owner := range owners {
				if owner.Kind == kind && owner.Name == name {

					if v, err := ExtractDeploymentRevision(rs.Annotations); err == nil && v == toVersion {
						vrs = &rs
						break
					}
				}
			}
		}
		if vrs != nil {
			break
		}
	}
	if vrs == nil {
		return fmt.Errorf("rollbackDeployment get rs [%s %s] err : not found ", kind, name)
	}
	spec := vrs.Spec.Template.Spec

	deploy.Spec.Template.Spec = spec
	err = d.kubectl.Resource(&deploy).Update(&deploy).Error
	if err != nil {
		return fmt.Errorf(" rollbackDeployment rollout undo deployment  err %v ", err)
	}

	return nil
}

// ExtractDeploymentRevision extracts the value of deployment.kubernetes.io/revision from annotations and converts it to int
func ExtractDeploymentRevision(annotations map[string]string) (int, error) {
	const revisionKey = "deployment.kubernetes.io/revision"

	// Check if annotations is nil
	if annotations == nil {
		return 0, errors.New("annotations is nil")
	}

	// Get revision value
	revisionStr, exists := annotations[revisionKey]
	if !exists {
		return 0, fmt.Errorf("annotation %q not found", revisionKey)
	}

	// Convert to int
	revision, err := strconv.Atoi(revisionStr)
	if err != nil {
		return 0, fmt.Errorf("failed to convert %q to int: %v", revisionStr, err)
	}

	return revision, nil
}

func (d *rollout) rollbackDaemonSet(toVersion int) error {
	// Find the specified version from ControllerVersion list
	// Convert Revision.Data.Raw to DaemonSet
	// Extract DaemonSet's Spec.Template.Spec, assign it to the original DaemonSet, update
	// Complete rollback

	kind := d.kubectl.Statement.GVK.Kind
	name := d.kubectl.Statement.Name
	ns := d.kubectl.Statement.Namespace

	var ds v1.DaemonSet
	err := d.kubectl.Resource(&ds).
		WithLabelSelector("app=" + name).
		Get(&ds).Error
	if err != nil {
		return fmt.Errorf("rollbackDaemonSet get daemonset err %v", err)
	}
	var versionList []*v1.ControllerRevision
	err = d.kubectl.newInstance().Resource(&v1.ControllerRevision{}).
		Namespace(ns).
		List(&versionList).Error
	if err != nil {
		return fmt.Errorf("rollbackDaemonSet list controllerrevisions err %v", err)
	}
	// If version not specified, rollback to previous version
	if toVersion == 0 {
		// Find the latest ControllerRevision to determine version
		// Find the maximum version
		var latestRevision int64 = 0
		for _, revision := range versionList {
			for _, owner := range revision.OwnerReferences {
				if owner.Kind == kind && owner.Name == name {
					// Select target version as latest version
					// Determine the latest version
					if revision.Revision > latestRevision {
						latestRevision = revision.Revision
					}
				}
			}
		}

		toVersion = int(latestRevision - 1)
		if toVersion <= 0 {
			// Add protection: if there are changes, version number must be greater than 0, minimum is 1
			toVersion = 1
		}
	}

	// Get target version's ControllerRevision and extract PodTemplateSpec

	// Find target version's ControllerRevision
	var targetRevision *v1.ControllerRevision

	for _, revision := range versionList {
		for _, owner := range revision.OwnerReferences {
			if owner.Kind == kind && owner.Name == name {
				if int(revision.Revision) == toVersion {
					targetRevision = revision
					break
				}
			}
		}
		if targetRevision != nil {
			break
		}
	}

	if targetRevision == nil {
		return fmt.Errorf("rollbackDaemonSet get target revision %d for %s not found", toVersion, name)
	}

	// Extract target version's PodTemplateSpec
	var dsTemplate v1.DaemonSet
	err = json.Unmarshal(targetRevision.Data.Raw, &dsTemplate)
	if err != nil {
		return fmt.Errorf("rollbackDaemonSet unmarshal controllerrevision data err %v", err)
	}

	// Update current DaemonSet with target version's template
	ds.Spec.Template.Spec = dsTemplate.Spec.Template.Spec

	// Update DaemonSet
	err = d.kubectl.Resource(&ds).Update(&ds).Error
	if err != nil {
		return fmt.Errorf("rollbackDaemonSet update daemonset err %v", err)
	}

	return nil
}
func (d *rollout) rollbackStatefulSet(toVersion int) error {
	// Find the specified version from ControllerVersion list
	// Convert Revision.Data.Raw to StatefulSet
	// Extract StatefulSet's Spec.Template.Spec, assign it to the original StatefulSet, update
	// Complete rollback

	kind := d.kubectl.Statement.GVK.Kind
	name := d.kubectl.Statement.Name
	ns := d.kubectl.Statement.Namespace

	var sts v1.StatefulSet
	err := d.kubectl.Resource(&sts).
		WithLabelSelector("app=" + name).
		Get(&sts).Error
	if err != nil {
		return fmt.Errorf("rollbackStatefulSet get StatefulSet err %v", err)
	}
	var versionList []*v1.ControllerRevision
	err = d.kubectl.newInstance().Resource(&v1.ControllerRevision{}).
		Namespace(ns).
		List(&versionList).Error
	if err != nil {
		return fmt.Errorf("rollbackStatefulSet list controllerrevisions err %v", err)
	}
	// If version not specified, rollback to previous version
	if toVersion == 0 {
		// Find the latest ControllerRevision to determine version
		// Find the maximum version
		var latestRevision int64 = 0
		for _, revision := range versionList {
			for _, owner := range revision.OwnerReferences {
				if owner.Kind == kind && owner.Name == name {
					// Select target version as latest version
					// Determine the latest version
					if revision.Revision > latestRevision {
						latestRevision = revision.Revision
					}
				}
			}
		}

		toVersion = int(latestRevision - 1)
		if toVersion <= 0 {
			// Add protection: if there are changes, version number must be greater than 0, minimum is 1
			toVersion = 1
		}
	}

	// Get target version's ControllerRevision and extract PodTemplateSpec

	// Find target version's ControllerRevision
	var targetRevision *v1.ControllerRevision

	for _, revision := range versionList {
		for _, owner := range revision.OwnerReferences {
			if owner.Kind == kind && owner.Name == name {
				if int(revision.Revision) == toVersion {
					targetRevision = revision
					break
				}
			}
		}
		if targetRevision != nil {
			break
		}
	}

	if targetRevision == nil {
		return fmt.Errorf("rollbackStatefulSet get target revision %d for %s not found", toVersion, name)
	}

	// Extract target version's PodTemplateSpec
	var stsTemplate v1.StatefulSet
	err = json.Unmarshal(targetRevision.Data.Raw, &stsTemplate)
	if err != nil {
		return fmt.Errorf("rollbackStatefulSet unmarshal controllerrevision data err %v", err)
	}

	// Update current StatefulSet with target version's template
	sts.Spec.Template.Spec = stsTemplate.Spec.Template.Spec

	// Update StatefulSet
	err = d.kubectl.Resource(&sts).Update(&sts).Error
	if err != nil {
		return fmt.Errorf("rollbackStatefulSet update statefulset err %v", err)
	}

	return nil
}
