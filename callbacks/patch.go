package callbacks

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func Patch(k *kom.Kubectl) error {

	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context
	patchType := stmt.PatchType
	patchData := stmt.PatchData

	var res *unstructured.Unstructured
	var err error
	if name == "" {
		err = fmt.Errorf("Name must be specified when patching an object")
		return err
	}
	if namespaced {
		if ns == "" {
			ns = metav1.NamespaceDefault
		}
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Patch(ctx, name, patchType, []byte(patchData), metav1.PatchOptions{})
	} else {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Patch(ctx, name, patchType, []byte(patchData), metav1.PatchOptions{})
	}
	if err != nil {
		return err
	}

	stmt.RowsAffected = 1
	if stmt.RemoveManagedFields {
		utils.RemoveManagedFields(res)
	}
	// Convert unstructured back to original object
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, stmt.Dest)
	if err != nil {
		return err
	}

	return nil
}
