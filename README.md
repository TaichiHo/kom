# Kom - Kubernetes Operations Manager

[English](README.md) | [中文](README_cn.md)

[![kom](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](https://github.com/weibaohui/kom/blob/master/LICENSE)


## Introduction

`kom` is a tool designed for Kubernetes operations, serving as an SDK-level wrapper for kubectl and client-go. It provides a range of functionalities to manage Kubernetes resources, including creating, updating, deleting, and retrieving resources. The project supports operations on various Kubernetes resource types and can handle Custom Resource Definitions (CRD).
By using `kom`, you can easily perform actions such as creating, reading, updating, and deleting resources, fetching logs, and managing files within PODs.

## **Features**

1. Easy to Use: `kom` offers a comprehensive set of features, including create, update, delete, retrieve, and list operations for both built-in and CRD resources.
2. Multi-Cluster Support: With RegisterCluster, you can easily manage multiple Kubernetes clusters.
3. Chained Invocation: `kom` provides chained invocation, making resource operations simpler and more intuitive.
4. CRD Support: `kom` supports custom resource definitions (CRDs), allowing you to define and manage custom resources effortlessly.
5. Callback Mechanism Support: Enables easy extension of business logic without tight coupling to Kubernetes operations.
6. POD File Operations: Supports file management within PODs, making it easy to upload, download, and delete files.
7. Support the encapsulation of high-frequency operations, such as restarting deployments and scaling (expanding or shrinking capacity).
8. Support SQL queries for k8s resources.`select * from pod where `metadata.namespace`='kube-system' or `metadata.namespace`='default' order by `metadata.creationTimestamp` asc`
## Example Program

**k8m** is a lightweight Kubernetes management tool, implemented as a single file and based on `kom` and `amis`, supporting multiple platform architectures.

1. **Download**: Get the latest version from [https://github.com/weibaohui/k8m](https://github.com/weibaohui/k8m).
2. **Run**: Start with the command `./k8m` and access [http://127.0.0.1:3618](http://127.0.0.1:3618).

## Installation

```go
import (
    "github.com/weibaohui/kom/callbacks"
    "github.com/weibaohui/kom"
)
func main() {
    // Register the callback functions
    callbacks.RegisterInit()
    // Register clusters
    defaultKubeConfig := os.Getenv("KUBECONFIG")
    if defaultKubeConfig == "" {
        defaultKubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
    }
    _, _ = kom.Clusters().RegisterInCluster()
    _, _ = kom.Clusters().RegisterByPathWithID(defaultKubeConfig, "default")
    kom.Clusters().Show()
    // Additional logic
}
```
## Usage Examples

### 1. Multi-Cluster Management

#### Registering Multiple Clusters
```go
// Register the InCluster cluster with the name "InCluster"
kom.Clusters().RegisterInCluster()

// Register two named clusters with IDs "orb" and "docker-desktop"
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/orb", "orb")
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/config", "docker-desktop")

// Register a cluster named "default". kom.DefaultCluster() will return this cluster.
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/config", "default")
```

#### Display Registered Clusters
```go
kom.Clusters().Show()
```

#### Selecting the Default Cluster
```go
// Use the default cluster to query pods in the kube-system namespace
// First, it will try to return the instance with ID "InCluster". If it doesn't exist,
// it will try to return the instance with ID "default".
// If neither of these exist, it will return any instance in the clusters list.
var pods []corev1.Pod
err = kom.DefaultCluster().Resource(&corev1.Pod{}).Namespace("kube-system").List(&pods).Error
```

#### Selecting a Specific Cluster
```go
// Select the "orb" cluster and query pods in the kube-system namespace
var pods []corev1.Pod
err = kom.Cluster("orb").Resource(&corev1.Pod{}).Namespace("kube-system").List(&pods).Error
```

### 2. CRUD and Watch Examples for Built-in Resource Objects

Define a Deployment object and use `kom` for resource operations.
```go
var item v1.Deployment
var items []v1.Deployment
```

#### Create a Resource
```go
item = v1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "nginx",
        Namespace: "default",
    },
    Spec: v1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {Name: "test", Image: "nginx:1.14.2"},
                },
            },
        },
    },
}
err := kom.DefaultCluster().Resource(&item).Create(&item).Error
```

#### Get a Specific Resource
```go
// Retrieve the Deployment named "nginx" in the "default" namespace
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Get(&item).Error
```

#### List Resources
```go
// List  Deployments in the "default" namespace
err := kom.DefaultCluster().Resource(&item).Namespace("default").List(&items).Error
// List  Deployments in all namespace
err := kom.DefaultCluster().Resource(&item).AllNamespace().List(&items).Error
err := kom.DefaultCluster().Resource(&item).Namespace("*").List(&items).Error
```

#### List Resources by Label
```go
// List Deployments in the "default" namespace with the label "app=nginx"
err := kom.DefaultCluster().Resource(&item).Namespace("default").WithLabelSelector("app=nginx").List(&items).Error
```

#### List Resources by Multiple Labels
```go
// List Deployments in the "default" namespace with labels "app=nginx" and "m=n"
err := kom.DefaultCluster().Resource(&item).Namespace("default").WithLabelSelector("app=nginx").WithLabelSelector("m=n").List(&items).Error
```

#### List Resources by Field
```go
// List Deployments in the "default" namespace with the field "metadata.name=test-deploy"
err := kom.DefaultCluster().Resource(&item).Namespace("default").WithFieldSelector("metadata.name=test-deploy").List(&items).Error
```

#### Update a Resource
```go
// Update the Deployment named "nginx" by adding an annotation
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Get(&item).Error
if item.Spec.Template.Annotations == nil {
    item.Spec.Template.Annotations = map[string]string{}
}
item.Spec.Template.Annotations["kom.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
err = kom.DefaultCluster().Resource(&item).Update(&item).Error
```

#### PATCH Update a Resource
```go
// Patch update to add a label and set replicas to 5 for the Deployment "nginx"
patchData := `{
    "spec": {
        "replicas": 5
    },
    "metadata": {
        "labels": {
            "new-label": "new-value"
        }
    }
}`
err := kom.DefaultCluster().Resource(&item).Patch(&item, types.MergePatchType, patchData).Error
```

#### Delete a Resource
```go
// Delete the Deployment named "nginx"
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Delete().Error
```

#### Retrieve Generic Resource (for both built-in and CRD types)
```go
// Specify GVK to retrieve resources
var list []corev1.Event
err := kom.DefaultCluster().GVK("events.k8s.io", "v1", "Event").Namespace("default").List(&list).Error
```

#### Watch Resource Changes
```go
// Watch changes to Pod resources in the "default" namespace
var watcher watch.Interface
var pod corev1.Pod
err := kom.DefaultCluster().Resource(&pod).Namespace("default").Watch(&watcher).Error
if err != nil {
    fmt.Printf("Create Watcher Error %v", err)
    return err
}
go func() {
    defer watcher.Stop()

    for event := range watcher.ResultChan() {
        err := kom.DefaultCluster().Tools().ConvertRuntimeObjectToTypedObject(event.Object, &pod)
        if err != nil {
            fmt.Printf("Failed to convert object to *v1.Pod type: %v", err)
            return
        }
        // Handle events
        switch event.Type {
        case watch.Added:
            fmt.Printf("Added Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
        case watch.Modified:
            fmt.Printf("Modified Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
        case watch.Deleted:
            fmt.Printf("Deleted Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
        }
    }
}()
```
#### Describe a resource
```go
// Describe a Deployment named nginx in default namespace
var describeResult []byte
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Describe(&describeResult).Error
fmt.Printf("describeResult: %s", describeResult)
```

### 3. YAML Create, Update, Delete
```go
yaml := `apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config
  namespace: default
data:
  key: value
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
        - name: example-container
          image: nginx
`

// Initial Apply creates the resources and returns results for each resource
results := kom.DefaultCluster().Applier().Apply(yaml)

// Subsequent Apply updates the resources and returns results for each resource
results = kom.DefaultCluster().Applier().Apply(yaml)

// Delete removes the resources and returns results for each resource
results = kom.DefaultCluster().Applier().Delete(yaml)
```

### 4. Pod Operations

#### Retrieve Logs
```go
// Retrieve Pod logs
var stream io.ReadCloser
err := kom.DefaultCluster().Namespace("default").Name("random-char-pod").Ctl().Pod().ContainerName("container").GetLogs(&stream, &corev1.PodLogOptions{}).Error
reader := bufio.NewReader(stream)
line, _ := reader.ReadString('\n')
fmt.Println(line)
```

#### Execute a Command
To execute a command inside a Pod, specify the container name. This triggers `Exec()` type callbacks.
```go
// Execute the "ps -ef" command inside the Pod
var execResult string
err := kom.DefaultCluster().Namespace("default").Name("random-char-pod").ContainerName("container").Command("ps", "-ef").ExecuteCommand(&execResult).Error
fmt.Printf("execResult: %s", execResult)
```

#### List Files
```go
// List files in the /etc directory within the Pod
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").ListFiles("/etc")
```

#### Download a File
```go
// Download the /etc/hosts file from inside the Pod
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").DownloadFile("/etc/hosts")
```

#### Upload a File
```go
// Upload text content to /etc/demo.txt inside the Pod
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").SaveFile("/etc/demo.txt", "txt-context")

// Directly upload a os.File to /etc/ inside the Pod
file, _ := os.Open(tempFilePath)
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").UploadFile("/etc/", file)
```

#### Delete a File
```go
// Delete the /etc/xyz file inside the Pod
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").DeleteFile("/etc/xyz")
```

### 5. Custom Resource Definition (CRD) Create, Update, Delete, and Watch Operations

Without defining a CR, you can still perform CRUD operations similar to built-in Kubernetes resources. To work with a CRD, define the object as `unstructured.Unstructured`, and specify the Group, Version, and Kind. For convenience, `kom.DefaultCluster().CRD(group, version, kind)` can be used to simplify the process. Below is an example of working with CRDs:

First, define a generic object to handle CRD responses.
```go
var item unstructured.Unstructured
```

#### Create a CRD
```go
yaml := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crontabs.stable.example.com
spec:
  group: stable.example.com
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                cronSpec:
                  type: string
                image:
                  type: string
                replicas:
                  type: integer
  scope: Namespaced
  names:
    plural: crontabs
    singular: crontab
    kind: CronTab
    shortNames:
    - ct`
result := kom.DefaultCluster().Applier().Apply(yaml)
```

#### Create a CR Object for the CRD
```go
item = unstructured.Unstructured{
    Object: map[string]interface{}{
        "apiVersion": "stable.example.com/v1",
        "kind":       "CronTab",
        "metadata": map[string]interface{}{
            "name":      "test-crontab",
            "namespace": "default",
        },
        "spec": map[string]interface{}{
            "cronSpec": "* * * * */8",
            "image":    "test-crontab-image",
        },
    },
}
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace(item.GetNamespace()).Name(item.GetName()).Create(&item).Error
```

#### Get a Single CR Object
```go
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(item.GetName()).Namespace(item.GetNamespace()).Get(&item).Error
```

#### List CR Objects
```go
var crontabList []unstructured.Unstructured
// list in default namespace
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace(crontab.GetNamespace()).List(&crontabList).Error
// list in all namespace
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").AllNamespace().List(&crontabList).Error
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace("*").List(&crontabList).Error
```

#### Update a CR Object
```go
patchData := `{
    "spec": {
        "image": "patch-image"
    },
    "metadata": {
        "labels": {
            "new-label": "new-value"
        }
    }
}`
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).Patch(&crontab, types.MergePatchType, patchData).Error
```

#### Delete a CR Object
```go
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).Delete().Error
```

#### Watch CR Object
```go
var watcher watch.Interface

err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace("default").Watch(&watcher).Error
if err != nil {
    fmt.Printf("Create Watcher Error %v", err)
}
go func() {
    defer watcher.Stop()

    for event := range watcher.ResultChan() {
        var item *unstructured.Unstructured

        item, err := kom.DefaultCluster().Tools().ConvertRuntimeObjectToUnstructuredObject(event.Object)
        if err != nil {
            fmt.Printf("Unable to convert object to Unstructured type: %v", err)
            return
        }
        // Handle events
        switch event.Type {
        case watch.Added:
            fmt.Printf("Added Unstructured [ %s/%s ]\n", item.GetNamespace(), item.GetName())
        case watch.Modified:
            fmt.Printf("Modified Unstructured [ %s/%s ]\n", item.GetNamespace(), item.GetName())
        case watch.Deleted:
            fmt.Printf("Deleted Unstructured [ %s/%s ]\n", item.GetNamespace(), item.GetName())
        }
    }
}()
```
#### Describe CR Object
```go
// Describe  a Deployment named nginx in the default namespace
var describeResult []byte
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace("default").Name(item.GetName()).Describe(&describeResult).Error
fmt.Printf("describeResult: %s", describeResult)
```

### 6. Cluster Parameter Information

Retrieve various types of information about the cluster:

```go
// Cluster documentation
kom.DefaultCluster().Status().Docs()

// Cluster resource information
kom.DefaultCluster().Status().APIResources()

// List of registered CRDs in the cluster
kom.DefaultCluster().Status().CRDList()

// Cluster version information
kom.DefaultCluster().Status().ServerVersion()
```

### 7. Callback Mechanism

`kom` has a built-in callback mechanism, allowing for custom callback functions that execute after specific operations. If a callback function returns `true`, the subsequent operation continues; otherwise, it terminates.

#### Supported Callbacks
- Available callbacks include: `get`, `list`, `create`, `update`, `patch`, `delete`, `exec`, `logs`, and `watch`.
- Default callback names: `"kom:get"`, `"kom:list"`, `"kom:create"`, `"kom:update"`, `"kom:patch"`, `"kom:watch"`, `"kom:delete"`, `"kom:pod:exec"`, `"kom:pod:logs"`.

#### Callback Function Management
- **Ordering**: Callbacks execute in the order of registration by default. Set execution order using `.After("kom:get")` or `.Before("kom:get")`.
- **Deletion**: Remove a callback with `.Delete("kom:get")`.
- **Replacement**: Replace a callback with `.Replace("kom:get", cb)`.

#### Callback Registration Examples

```go
// Register callback for Get operation
kom.DefaultCluster().Callback().Get().Register("get", cb)
// Register callback for List operation
kom.DefaultCluster().Callback().List().Register("list", cb)
// Register callback for Create operation
kom.DefaultCluster().Callback().Create().Register("create", cb)
// Register callback for Update operation
kom.DefaultCluster().Callback().Update().Register("update", cb)
// Register callback for Patch operation
kom.DefaultCluster().Callback().Patch().Register("patch", cb)
// Register callback for Delete operation
kom.DefaultCluster().Callback().Delete().Register("delete", cb)
// Register callback for Watch operation
kom.DefaultCluster().Callback().Watch().Register("watch", cb)
// Register callback for Pod Exec operation
kom.DefaultCluster().Callback().Exec().Register("exec", cb)
// Register callback for Log retrieval
kom.DefaultCluster().Callback().Logs().Register("logs", cb)

// Delete callback for Get operation
kom.DefaultCluster().Callback().Get().Delete("get")
// Replace callback for Get operation
kom.DefaultCluster().Callback().Get().Replace("get", cb)

// Specify callback execution order
kom.DefaultCluster().Callback().After("kom:get").Register("get", cb)
kom.DefaultCluster().Callback().Before("kom:create").Register("create", cb)

// Example Scenarios
// 1. Perform permission check before Create operation. If unauthorized, return an error, halting further execution.
// 2. After List operation, filter results, removing resources that do not meet specific criteria.

```

#### Custom Callback Function

Define a custom callback function to include specific operations:

```go
func cb(k *kom.Kubectl) error {
    stmt := k.Statement
    gvr := stmt.GVR
    ns := stmt.Namespace
    name := stmt.Name
    // Print information
    fmt.Printf("Get %s/%s(%s)\n", ns, name, gvr)
    fmt.Printf("Command %s/%s(%s %s)\n", ns, name, stmt.Command, stmt.Args)
    return nil
    // return fmt.Errorf("error") // Return error to stop further callback execution
}
```


### 8. SQL Queries for k8s Resources
* Query k8s resources through the SQL() method, which is simple and efficient.
* The table names support the full names and abbreviations of all resources registered within the cluster, including CRD resources. As long as they are registered on the cluster, they can be queried.
* Typical table names include: pod, deployment, service, ingress, pvc, pv, node, namespace, secret, configmap, serviceaccount, role, rolebinding, clusterrole, clusterrolebinding, crd, cr, hpa, daemonset, statefulset, job, cronjob, limitrange, horizontalpodautoscaler, poddisruptionbudget, networkpolicy, endpoints, ingressclass, mutatingwebhookconfiguration, validatingwebhookconfiguration, customresourcedefinition, storageclass, persistentvolumeclaim, persistentvolume, horizontalpodautoscaler, podsecurity. All of them can be queried.
* The query fields currently only support "*"
* The query conditions currently support =,!=, >=, <=, <>, like, in, not in, and, or, between.
* The sorting fields currently support sorting on a single field. By default, they are sorted in descending order according to the creation time.
#### Query k8s Built-in Resources
```go
    sql := "select * from deploy where metadata.namespace='kube-system' or metadata.namespace='default' order by  metadata.creationTimestamp asc   "

	var list []v1.Deployment
	err := kom.DefaultCluster().Sql(sql).List(&list).Error
	for _, d := range list {
        fmt.Printf("List Items foreach %s,%s at %s \n", d.GetNamespace(), d.GetName(), d.GetCreationTimestamp())
    }
```
#### Query CRD Resources
```go
    // vm is the CRD of Kubevirt
    sql := "select * from vm where (metadata.namespace='kube-system' or metadata.namespace='default' )  "
	var list []unstructured.Unstructured
	err := kom.DefaultCluster().Sql(sql).List(&list).Error
	for _, d := range list {
        fmt.Printf("List Items foreach %s,%s\n", d.GetNamespace(), d.GetName())
    }
``` 
#### Chained Query with SQL
```go
// Query the pod list
err := kom.DefaultCluster().From("pod").
		Where("metadata.namespace =?  or metadata.namespace=? ", "kube-system", "default").
		Order("metadata.creationTimestamp desc").
		List(&list).Error
``` 


### 9. Other Operations
#### Restart Deployment
```go
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Restart()
```
#### Scale Deployment
```go
// Set the replica count of the nginx deployment to 3
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Scale(3)
```
#### Update Deployment Tag
```go
// Upgrade the container image tag of the nginx deployment to alpine
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Deployment().ReplaceImageTag("main","20241124")
```
#### Deployment Rollout History
```go
// Query the upgrade history of the nginx deployment
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().History()
```
#### Deployment Rollout Undo
```go
// Rollback the nginx deployment
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Undo()
// Rollback the nginx deployment to a specific version (query the history)
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Undo("6")
```
#### Deployment Rollout Pause
```go
// Pause the upgrade process
err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Pause()
```
#### Deployment Rollout Resume
```go
// Resume the upgrade process
err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Resume()
```
#### Deployment Rollout Status
```go
// Check the status of the nginx deployment rollout
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Status()
```
#### Taint Node
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().Taint("dedicated=special-user:NoSchedule")
```
#### Remove Taint from Node
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().UnTaint("dedicated=special-user:NoSchedule")
```
#### Cordon Node
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().Cordon()
```
#### UnCordon Node
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().UnCordon()
```
#### Drain Node
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().Drain()
```
#### Label Resource
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Label("name=zhangsan")
```
#### Remove Label from Resource
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Label("name-")
```
#### Annotate Resource
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Annotate("name=zhangsan")
```
#### Remove Annotation from Resource
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Annotate("name-")
```

# KOM MCP Server

## Connecting with Cursor

To connect Cursor to the KOM MCP server:

1. Ensure the MCP server is running in your Minikube cluster:
```bash
kubectl get pods -l app=kom-mcp-server
```

2. Set up port forwarding:
```bash
kubectl port-forward svc/kom-mcp-server 9096:9096
```

3. Create a `.cursor/mcp.json` file in your project root with the following content:
```json
{
  "mcpServers": {
    "kom": {
      "url": "http://localhost:9096/sse",
      "env": {
        "API_KEY": "value"
      }
    }
  }
}
```

4. Restart Cursor to apply the configuration.

## Troubleshooting

If you can't connect to the MCP server:

1. Check if the pod is running:
```bash
kubectl get pods -l app=kom-mcp-server
```

2. Check pod logs:
```bash
kubectl logs -l app=kom-mcp-server
```

3. Verify port forwarding is working, getting a 404 error means this IS working:
```bash
curl -v localhost:9096/health
```
