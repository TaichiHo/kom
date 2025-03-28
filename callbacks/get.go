package callbacks

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func Get(k *kom.Kubectl) error {
	var err error
	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context
	conditions := stmt.Filter.Conditions
	// If where conditions are set, List should be used because SQL queries return a list, even if it has only one element
	if len(conditions) > 0 {
		return fmt.Errorf("Please use List for SQL queries, if you need to get a single resource, get it from the List")
	}
	if name == "" {
		err = fmt.Errorf("Name must be specified when getting an object")
		return err
	}

	cacheKey := fmt.Sprintf("%s/%s/%s/%s/%s", ns, name, gvr.Group, gvr.Resource, gvr.Version)
	res, err := utils.GetOrSetCache(stmt.Kubectl.ClusterCache(), cacheKey, stmt.CacheTTL, func() (ret *unstructured.Unstructured, err error) {
		if namespaced {
			if ns == "" {
				ns = metav1.NamespaceDefault
			}
			ret, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
		} else {
			ret, err = stmt.Kubectl.DynamicClient().Resource(gvr).Get(ctx, name, metav1.GetOptions{})
		}
		return
	})
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
