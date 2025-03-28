package callbacks

import (
	"fmt"
	"reflect"

	"github.com/weibaohui/kom/kom"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func Watch(k *kom.Kubectl) error {

	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	ctx := stmt.Context
	namespaceList := stmt.NamespaceList

	opts := stmt.ListOptions
	listOptions := metav1.ListOptions{}
	if len(opts) > 0 {
		listOptions = opts[0]
	}

	destValue := reflect.ValueOf(stmt.Dest)

	// Ensure dest is a pointer to an interface
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Interface {
		return fmt.Errorf("stmt.Dest must be a pointer to watch.Interface")
	}

	// Ensure dest's actual type implements the watch.Interface interface
	if !destValue.Elem().Type().Implements(reflect.TypeOf((*watch.Interface)(nil)).Elem()) {
		return fmt.Errorf("stmt.Dest must implement watch.Interface interface")
	}

	var watcher watch.Interface
	var err error

	if namespaced {
		if stmt.AllNamespace || len(namespaceList) > 1 {
			// All namespaces or multiple namespaces provided
			// client-go doesn't support cross-namespace queries, so get all and filter later
			ns = metav1.NamespaceAll
		} else {
			// Not all namespaces and no multiple namespaces provided
			if ns == "" {
				ns = metav1.NamespaceDefault
			}
		}

		watcher, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Watch(ctx, listOptions)
	} else {
		watcher, err = stmt.Kubectl.DynamicClient().Resource(gvr).Watch(ctx, listOptions)
	}
	if err != nil {
		return err
	}

	// Assign watcher to dest
	destValue.Elem().Set(reflect.ValueOf(watcher))

	return nil
}
