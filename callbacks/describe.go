package callbacks

import (
	"fmt"
	"reflect"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/kom/describe"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

func Describe(k *kom.Kubectl) error {

	stmt := k.Statement
	ns := stmt.Namespace
	name := stmt.Name
	gvk := k.Statement.GVK
	namespaced := stmt.Namespaced

	if stmt.GVK.Empty() {
		return fmt.Errorf("Please call GVK() method to set GroupVersionKind")
	}

	// Reflection check
	destValue := reflect.ValueOf(stmt.Dest)

	// Ensure dest is a pointer to a byte slice
	if !(destValue.Kind() == reflect.Ptr && destValue.Elem().Kind() == reflect.Slice) || destValue.Elem().Type().Elem().Kind() != reflect.Uint8 {
		return fmt.Errorf("Please ensure dest is a pointer to a byte slice. Define var s []byte and use &s")
	}

	if namespaced {
		if stmt.AllNamespace {
			ns = metav1.NamespaceAll
		} else {
			if ns == "" {
				ns = metav1.NamespaceDefault
			}
		}
	} else {
		ns = metav1.NamespaceNone
	}

	var output string
	var err error
	// Execute describe
	m := k.Status().DescriberMap()
	gk := schema.GroupKind{
		Group: gvk.Group,
		Kind:  gvk.Kind,
	}
	// First look in the built-in describerMap
	if d, ok := m[gk]; ok {
		output, err = d.Describe(ns, name, describe.DescriberSettings{
			ShowEvents: true,
		})
		if err != nil {
			return fmt.Errorf("DescriberMap describe %s/%s error: %v", gvk.String(), name, err)
		}
	} else {
		// No built-in descriptor
		mapping := &meta.RESTMapping{
			Resource: k.Statement.GVR,
		}
		if gd, b := describe.GenericDescriberFor(mapping, k.RestConfig()); b {
			output, err = gd.Describe(ns, name, describe.DescriberSettings{
				ShowEvents: true,
			})
			if err != nil {
				return fmt.Errorf("GenericDescriber describe %s/%s error: %v", gvk.String(), name, err)
			}
		}
	}

	// Write result to tx.Statement.Dest
	if destBytes, ok := k.Statement.Dest.(*[]byte); ok {
		// Directly assign using outBuf.Bytes()
		*destBytes = []byte(output)
		klog.V(8).Infof("Describe result %s", *destBytes)
	} else {
		return fmt.Errorf("dest is not a *[]byte")
	}
	return nil
}
