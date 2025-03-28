package callbacks

import (
	"fmt"
	"reflect"

	"github.com/weibaohui/kom/kom"
)

func GetLogs(k *kom.Kubectl) error {

	stmt := k.Statement
	ns := stmt.Namespace
	name := stmt.Name
	options := stmt.PodLogOptions
	options.Container = stmt.ContainerName
	ctx := stmt.Context

	// If there's only one container, it doesn't need to be set
	// if stmt.ContainerName == "" {
	// 	return fmt.Errorf("Please call ContainerName() method to set Pod container name")
	// }

	// Use reflection to get the value of dest
	destValue := reflect.ValueOf(stmt.Dest)

	// Ensure dest is a pointer
	if destValue.Kind() != reflect.Ptr {
		// Handle error: dest is not a pointer
		return fmt.Errorf("Target container must be a pointer type")
	}

	stream, err := k.Client().CoreV1().Pods(ns).GetLogs(name, options).Stream(ctx)
	if err != nil {
		return err
	}
	// Assign stream to dest
	destValue.Elem().Set(reflect.ValueOf(stream))
	return nil
}
