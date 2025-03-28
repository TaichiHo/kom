package kom

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
)

type statefulSet struct {
	kubectl *Kubectl
}

func (s *statefulSet) Restart() error {
	return s.kubectl.Ctl().Rollout().Restart()
}
func (s *statefulSet) Scale(replicas int32) error {
	return s.kubectl.Ctl().Scale(replicas)
}

func (s *statefulSet) Stop() error {
	return s.kubectl.Ctl().Scaler().Stop()
}
func (s *statefulSet) Restore() error {
	return s.kubectl.Ctl().Scaler().Restore()
}

func (s *statefulSet) ManagedPods() ([]*corev1.Pod, error) {
	// First find the StatefulSet
	var sts v1.StatefulSet
	err := s.kubectl.WithCache(s.kubectl.Statement.CacheTTL).Resource(&sts).Get(&sts).Error

	if err != nil {
		return nil, err
	}
	// Get pods through StatefulSet
	var podList []*corev1.Pod
	err = s.kubectl.newInstance().WithCache(s.kubectl.Statement.CacheTTL).Resource(&corev1.Pod{}).
		Namespace(s.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", sts.GetName(), "StatefulSet")).
		List(&podList).Error
	return podList, err
}
func (s *statefulSet) ManagedPod() (*corev1.Pod, error) {
	podList, err := s.ManagedPods()
	if err != nil {
		return nil, err
	}
	if len(podList) > 0 {
		return podList[0], nil
	}
	return nil, fmt.Errorf("no Pod found under StatefulSet[%s]", s.kubectl.Statement.Name)
}
func (s *statefulSet) HPAList() ([]*autoscalingv2.HorizontalPodAutoscaler, error) {
	// Get pods through StatefulSet
	var list []*autoscalingv2.HorizontalPodAutoscaler
	err := s.kubectl.newInstance().WithCache(s.kubectl.Statement.CacheTTL).
		GVK("autoscaling", "v2", "HorizontalPodAutoscaler").
		Namespace(s.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("spec.scaleTargetRef.name='%s' and spec.scaleTargetRef.kind='%s'", s.kubectl.Statement.Name, "StatefulSet")).
		List(&list).Error
	return list, err
}
