package kom

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type scale struct {
	kubectl *Kubectl
}

func (s *scale) Scale(replicas int32) error {

	kind := s.kubectl.Statement.GVK.Kind
	klog.V(8).Infof("scale Kind=%s", kind)
	klog.V(8).Infof("scale Resource=%s", s.kubectl.Statement.GVR.Resource)
	klog.V(8).Infof("scale %s/%s", s.kubectl.Statement.Namespace, s.kubectl.Statement.Name)

	// Currently supported resources for restart method:
	// Deployment
	// StatefulSet
	// ReplicaSet
	// ReplicationController

	if !isSupportedKind(kind, []string{"Deployment", "StatefulSet", "ReplicationController", "ReplicaSet"}) {
		s.kubectl.Error = fmt.Errorf("%s %s/%s Scale is not supported", kind, s.kubectl.Statement.Namespace, s.kubectl.Statement.Name)
		return s.kubectl.Error
	}

	var item interface{}
	patchData := fmt.Sprintf("{\"spec\":{\"replicas\":%d}}", replicas)
	err := s.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	if err != nil {
		s.kubectl.Error = fmt.Errorf("%s %s/%s scale error %v", kind, s.kubectl.Statement.Namespace, s.kubectl.Statement.Name, err)
		return err
	}
	return s.kubectl.Error
}

// Stop stops the deployment
// Before stopping, records the current number of replicas in the deployment's annotation
// kom.restore.replicas
func (s *scale) Stop() error {
	kind := s.kubectl.Statement.GVK.Kind
	if !isSupportedKind(kind, []string{"Deployment", "StatefulSet", "ReplicationController", "ReplicaSet"}) {
		s.kubectl.Error = fmt.Errorf("%s %s/%s Scale is not supported", kind, s.kubectl.Statement.Namespace, s.kubectl.Statement.Name)
		return s.kubectl.Error
	}
	var item unstructured.Unstructured
	err := s.kubectl.Get(&item).Error
	if err != nil {
		return err
	}

	replicas, found, err := unstructured.NestedInt64(item.Object, "spec", "replicas")
	if err != nil {
		return fmt.Errorf("Error fetching replicas: %v\n", err)
	}
	if !found {
		return fmt.Errorf("spec.replicas not found\n")
	}

	if replicas == 0 {
		// Already stopped
		return nil
	}
	patchData := fmt.Sprintf(`{
	"spec": {
		"replicas": %d
	},
	"metadata": {
		"annotations": {
			"kom.restore.replicas": "%d"
		}
	}
}`, 0, replicas)
	err = s.kubectl.Patch(&item, types.MergePatchType, patchData).Error

	if err != nil {
		return fmt.Errorf("stop %s/%s error %v", item.GetNamespace(), item.GetName(), err)
	}
	return nil

}

// Restore restores the deployment
// If kom.restore.replicas annotation exists in the deployment
// Sets the value of kom.restore.replicas as the deployment's replicas
// If not found, sets it to 1
func (s *scale) Restore() error {
	kind := s.kubectl.Statement.GVK.Kind
	if !isSupportedKind(kind, []string{"Deployment", "StatefulSet", "ReplicationController", "ReplicaSet"}) {
		s.kubectl.Error = fmt.Errorf("%s %s/%s Scale is not supported", kind, s.kubectl.Statement.Namespace, s.kubectl.Statement.Name)
		return s.kubectl.Error
	}
	var item unstructured.Unstructured
	err := s.kubectl.Get(&item).Error
	if err != nil {
		return err
	}
	annotations, found, err := unstructured.NestedStringMap(item.Object, "metadata", "annotations")
	if err != nil {
		return fmt.Errorf("error fetching annotations: %v\n", err)
	}
	if !found {
		return fmt.Errorf("annotations not found")
	}
	targetReplicas := int32(1)

	if restoreReplicas, exists := annotations["kom.restore.replicas"]; exists {
		if i, err := strconv.ParseInt(restoreReplicas, 10, 64); err == nil {
			targetReplicas = int32(i)
		}
	}

	patchData := fmt.Sprintf(`{
	"spec": {
		"replicas": %d
	},
	"metadata": {
		"annotations": {
			"kom.restore.replicas": null
		}
	}
}`, targetReplicas)
	err = s.kubectl.Patch(&item, types.MergePatchType, patchData).Error

	if err != nil {
		return fmt.Errorf("stop %s/%s error %v", item.GetNamespace(), item.GetName(), err)
	}
	return nil

}
