package callbacks

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Delete(k *kom.Kubectl) error {
	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context
	forceDelete := stmt.ForceDelete // Add force delete flag

	// Modify delete options to support force delete
	deleteOptions := metav1.DeleteOptions{}
	if forceDelete {
		background := metav1.DeletePropagationBackground
		deleteOptions.PropagationPolicy = &background
		deleteOptions.GracePeriodSeconds = utils.Int64Ptr(0)
	}

	var err error
	if name == "" {
		err = fmt.Errorf("Name must be specified when deleting an object")
		return err
	}
	if namespaced {
		if ns == "" {
			ns = metav1.NamespaceDefault
		}

		err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Delete(ctx, name, deleteOptions)
	} else {
		err = stmt.Kubectl.DynamicClient().Resource(gvr).Delete(ctx, name, deleteOptions)
	}

	if err != nil {
		return err
	}
	stmt.RowsAffected = 1
	return nil
}
