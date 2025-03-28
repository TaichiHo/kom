package kom

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
)

type deploy struct {
	kubectl *Kubectl
}

func (d *deploy) Stop() error {
	return d.kubectl.Ctl().Scaler().Stop()
}
func (d *deploy) Restore() error {
	return d.kubectl.Ctl().Scaler().Restore()
}
func (d *deploy) Restart() error {
	return d.kubectl.Ctl().Rollout().Restart()
}
func (d *deploy) Scale(replicas int32) error {
	return d.kubectl.Ctl().Scale(replicas)
}
func (d *deploy) HPAList() ([]*autoscalingv2.HorizontalPodAutoscaler, error) {
	// Get pods through ReplicaSet
	var list []*autoscalingv2.HorizontalPodAutoscaler
	err := d.kubectl.newInstance().WithCache(d.kubectl.Statement.CacheTTL).
		GVK("autoscaling", "v2", "HorizontalPodAutoscaler").
		Resource(&autoscalingv2.HorizontalPodAutoscaler{}).
		Namespace(d.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("spec.scaleTargetRef.name='%s' and spec.scaleTargetRef.kind='%s'", d.kubectl.Statement.Name, "Deployment")).
		List(&list).Error
	return list, err
}
func (d *deploy) ManagedPods() ([]*corev1.Pod, error) {
	// First find the ReplicaSet
	rs, err := d.ManagedLatestReplicaSet()
	if err != nil {
		return nil, err
	}
	// Get pods through ReplicaSet
	var podList []*corev1.Pod
	err = d.kubectl.newInstance().WithCache(d.kubectl.Statement.CacheTTL).Resource(&corev1.Pod{}).
		Namespace(d.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", rs.GetName(), "ReplicaSet")).
		List(&podList).Error
	return podList, err
}
func (d *deploy) ManagedPod() (*corev1.Pod, error) {
	podList, err := d.ManagedPods()
	if err != nil {
		return nil, err
	}
	if len(podList) > 0 {
		return podList[0], nil
	}
	return nil, fmt.Errorf("no Pod found under Deployment[%s]", d.kubectl.Statement.Name)
}

// ManagedLatestReplicaSet returns the ReplicaSet of the latest deployment version
func (d *deploy) ManagedLatestReplicaSet() (*v1.ReplicaSet, error) {
	var item v1.Deployment
	err := d.kubectl.WithCache(d.kubectl.Statement.CacheTTL).Resource(&item).Get(&item).Error

	if err != nil {
		return nil, err
	}

	var rsList []*v1.ReplicaSet
	err = d.kubectl.newInstance().
		WithCache(d.kubectl.Statement.CacheTTL).
		Resource(&v1.ReplicaSet{}).
		Namespace(d.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", d.kubectl.Statement.Name, "Deployment")).
		List(&rsList).Error
	if err != nil {
		return nil, err
	}

	// First check how many ReplicaSets there are, if there's only one, that's the one
	if len(rsList) == 1 {
		return rsList[0], nil
	}

	// If there are multiple ReplicaSets, need to filter by revision
	// Look for annotation on Deployment
	// metadata:
	//   annotations:
	//     deployment.kubernetes.io/revision: "50"
	var revision string
	for k, v := range item.GetAnnotations() {
		if strings.HasPrefix(k, "deployment.kubernetes.io/revision") {
			// Found revision
			// deployment.kubernetes.io/revision: "50"
			// 50
			revision = v
			break
		}
	}

	for _, rs := range rsList {
		if rs.Annotations["deployment.kubernetes.io/revision"] == revision {
			return rs, nil
		}
	}
	return nil, fmt.Errorf("no latest ReplicaSet found under Deployment[%s]", item.GetName())
}

func (d *deploy) ReplaceImageTag(targetContainerName string, tag string) (*v1.Deployment, error) {
	var item v1.Deployment
	err := d.kubectl.Resource(&item).Get(&item).Error

	if err != nil {
		return nil, err
	}

	for i := range item.Spec.Template.Spec.Containers {
		c := &item.Spec.Template.Spec.Containers[i]
		if c.Name == targetContainerName {
			c.Image = replaceImageTag(c.Image, tag)
		}
	}
	err = d.kubectl.Resource(&item).Update(&item).Error
	return &item, err
}

// replaceImageTag replaces the tag of an image
func replaceImageTag(imageName, newTag string) string {
	// Check if image name contains a tag
	if strings.Contains(imageName, ":") {
		// Split image name and tag by ":"
		parts := strings.Split(imageName, ":")
		// Replace old tag with new tag
		return parts[0] + ":" + newTag
	} else {
		// If image name has no tag, directly add new tag
		return imageName + ":" + newTag
	}
}
