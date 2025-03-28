package kom

import (
	"context"
	"fmt"

	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// listResources lists all objects of the specified resource type
func (k *Kubectl) listResources(ctx context.Context, kind string, ns string) (resources []*unstructured.Unstructured, err error) {
	gvr, namespaced := k.Tools().GetGVRByKind(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("unsupported resource type: %s", kind)
	}

	listOptions := metav1.ListOptions{}

	var list *unstructured.UnstructuredList
	if namespaced {
		list, err = k.DynamicClient().Resource(gvr).Namespace(ns).List(ctx, listOptions)
	} else {
		list, err = k.DynamicClient().Resource(gvr).List(ctx, listOptions)
	}
	if err != nil {
		return nil, err
	}
	for _, item := range list.Items {
		obj := item.DeepCopy()
		utils.RemoveManagedFields(obj)
		resources = append(resources, obj)
	}

	return resources, nil
}
