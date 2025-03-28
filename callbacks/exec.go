package callbacks

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

func ExecuteCommand(k *kom.Kubectl) error {

	stmt := k.Statement
	ns := stmt.Namespace
	name := stmt.Name
	command := stmt.Command
	args := stmt.Args
	containerName := stmt.ContainerName
	ctx := stmt.Context

	if stmt.ContainerName == "" {
		return fmt.Errorf("Please call ContainerName() method to set Pod container name")
	}
	if stmt.Command == "" {
		return fmt.Errorf("Please call Command() method to set command")
	}

	// Reflection check
	destValue := reflect.ValueOf(stmt.Dest)

	// Ensure dest is a pointer to a byte slice
	if !(destValue.Kind() == reflect.Ptr && destValue.Elem().Kind() == reflect.Slice) || destValue.Elem().Type().Elem().Kind() != reflect.Uint8 {
		return fmt.Errorf("Please ensure dest is a pointer to a byte slice. Define var s []byte and use &s")
	}

	var err error
	cmd := []string{command}
	cmd = append(cmd, args...)
	klog.V(8).Infof("Execute %s %v in [%s/%s:%s]\n", command, args, ns, name, containerName)

	req := k.Client().CoreV1().RESTClient().
		Post().
		Namespace(ns).
		Resource("pods").
		Name(name).
		SubResource("exec").
		Param("container", containerName).
		Param("command", cmd[0])

	for _, arg := range cmd[1:] {
		req.Param("command", arg)
	}

	req.Param("tty", "false").
		Param("stdin", fmt.Sprintf("%v", stmt.Stdin != nil)).
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := createExecutor(req.URL(), k.RestConfig())
	if err != nil {
		return fmt.Errorf("error creating executor: %v", err)
	}

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	options := &remotecommand.StreamOptions{
		Stdout: &outBuf,
		Stderr: &errBuf,
		Tty:    false,
	}
	if stmt.Stdin != nil {
		options.Stdin = stmt.Stdin
	}
	err = executor.StreamWithContext(ctx, *options)
	if err != nil {
		s := errBuf.String()
		klog.V(8).Infof("Error executing command: %v", err)
		if strings.Contains(s, "Invalid argument") {
			return fmt.Errorf("System parameter error: %v", s)
		}
		return fmt.Errorf("error executing command: %v %v", err, s)
	}

	// Write result to tx.Statement.Dest
	if destBytes, ok := k.Statement.Dest.(*[]byte); ok {
		// Directly assign using outBuf.Bytes()
		*destBytes = outBuf.Bytes()
		klog.V(8).Infof("Execute result %s", *destBytes)
	} else {
		return fmt.Errorf("dest is not a *[]byte")
	}
	return nil
}

func createExecutor(url *url.URL, config *rest.Config) (remotecommand.Executor, error) {

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", url)
	if err != nil {
		return nil, err
	}
	// Fallback executor is default, unless feature flag is explicitly disabled.
	// WebSocketExecutor must be "GET" method as described in RFC 6455 Sec. 4.1 (page 17).
	websocketExec, err := remotecommand.NewWebSocketExecutor(config, "GET", url.String())
	if err != nil {
		return nil, err
	}
	exec, err = remotecommand.NewFallbackExecutor(websocketExec, exec, func(err error) bool {
		return httpstream.IsUpgradeFailure(err) || httpstream.IsHTTPSProxyError(err)
	})
	if err != nil {
		return nil, err
	}
	return exec, nil
}
