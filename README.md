# Kom - Kubernetes Operations Manager

[English](README_en.md) | [中文](README.md)
[![kom](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](https://github.com/weibaohui/kom/blob/master/LICENSE)


## 简介

`kom` 是一个用于 Kubernetes 操作的工具，相当于SDK级的kubectl、client-go的使用封装。
它提供了一系列功能来管理 Kubernetes 资源，包括创建、更新、删除和获取资源。这个项目支持多种 Kubernetes 资源类型的操作，并能够处理自定义资源定义（CRD）。
通过使用 `kom`，你可以轻松地进行资源的增删改查和日志获取以及操作POD内文件等动作，甚至可以使用SQL语句来查询、管理k8s资源。

## **特点**
1. 简单易用：kom 提供了丰富的功能，包括创建、更新、删除、获取、列表等，包括对内置资源以及CRD资源的操作。
2. 多集群支持：通过RegisterCluster，你可以轻松地管理多个 Kubernetes 集群。
3. 链式调用：kom 提供了链式调用，使得操作资源更加简单和直观。
4. 支持自定义资源定义（CRD）：kom 支持自定义资源定义（CRD），你可以轻松地定义和操作自定义资源。
5. 支持回调机制，轻松拓展业务逻辑，而不必跟k8s操作强耦合。
6. 支持POD内文件操作，轻松上传、下载、删除文件。
7. 支持高频操作封装，如deployment的restart重启、scale扩缩容等。
8. 支持SQL查询k8s资源。select * from pod where `metadata.namespace`='kube-system' or `metadata.namespace`='default' order by  `metadata.creationTimestamp` desc 

## 示例程序
**k8m** 是一个轻量级的 Kubernetes 管理工具，它基于kom、amis实现，单文件，支持多平台架构。
1. **下载**：从 [https://github.com/weibaohui/k8m](https://github.com/weibaohui/k8m) 下载最新版本。
2. **运行**：使用 `./k8m` 命令启动,访问[http://127.0.0.1:3618](http://127.0.0.1:3618)。




## 安装

```bash
import (
    "github.com/weibaohui/kom"
    "github.com/weibaohui/kom/callbacks"
)
func main() {
    // 注册回调，务必先注册
    callbacks.RegisterInit()
    // 注册集群
	defaultKubeConfig := os.Getenv("KUBECONFIG")
	if defaultKubeConfig == "" {
		defaultKubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	_, _ = kom.Clusters().RegisterInCluster()
	_, _ = kom.Clusters().RegisterByPathWithID(defaultKubeConfig, "default")
	kom.Clusters().Show()
	// 其他逻辑
}
```

## 使用示例

### 1. 多集群管理
#### 注册多集群
```go
// 注册InCluster集群，名称为InCluster
kom.Clusters().RegisterInCluster()
// 注册两个带名称的集群,分别名为orb和docker-desktop
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/orb", "orb")
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/config", "docker-desktop")
// 注册一个名为default的集群，那么kom.DefaultCluster()则会返回该集群。
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/config", "default")
```
#### 显示已注册集群
```go
kom.Clusters().Show()
```
#### 选择默认集群
```go
// 使用默认集群,查询集群内kube-system命名空间下的pod
// 首先尝试返回 ID 为 "InCluster" 的实例，如果不存在，
// 则尝试返回 ID 为 "default" 的实例。
// 如果上述两个名称的实例都不存在，则返回 clusters 列表中的任意一个实例。
var pods []corev1.Pod
err = kom.DefaultCluster().Resource(&corev1.Pod{}).Namespace("kube-system").List(&pods).Error
```
#### 选择指定集群
```go
// 选择orb集群,查询集群内kube-system命名空间下的pod
var pods []corev1.Pod
err = kom.Cluster("orb").Resource(&corev1.Pod{}).Namespace("kube-system").List(&pods).Error
```

### 2. 内置资源对象的增删改查以及Watch示例
定义一个 Deployment 对象，并通过 kom 进行资源操作。
```go
var item v1.Deployment
var items []v1.Deployment
```
#### 创建某个资源
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
#### Get查询某个资源
```go
// 查询 default 命名空间下名为 nginx 的 Deployment
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Get(&item).Error
```
#### List查询资源列表
```go
// 查询 default 命名空间下的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default").List(&items).Error
// 查询 所有 命名空间下的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("*").List(&items).Error
err := kom.DefaultCluster().Resource(&item).AllNamespace().List(&items).Error
```
#### 通过Label查询资源列表
```go
// 查询 default 命名空间下 标签为 app:nginx 的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default").WithLabelSelector("app=nginx").List(&items).Error
```
#### 通过多个Label查询资源列表
```go
// 查询 default 命名空间下 标签为 app:nginx m:n 的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default").WithLabelSelector("app=nginx").WithLabelSelector("m=n").List(&items).Error
```
#### 通过Field查询资源列表
```go
// 查询 default 命名空间下 标签为 metadata.name=test-deploy 的 Deployment 列表
// filedSelector 一般支持原生的字段定义。如metadata.name,metadata.namespace,metadata.labels,metadata.annotations,metadata.creationTimestamp,spec.nodeName,spec.serviceAccountName,spec.schedulerName,status.phase,status.hostIP,status.podIP,status.qosClass,spec.containers.name等字段
err := kom.DefaultCluster().Resource(&item).Namespace("default").WithFieldSelector("metadata.name=test-deploy").List(&items).Error
```
#### 更新资源内容
```go
// 更新名为nginx 的 Deployment，增加一个注解
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Get(&item).Error
if item.Spec.Template.Annotations == nil {
	item.Spec.Template.Annotations = map[string]string{}
}
item.Spec.Template.Annotations["kom.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
err = kom.DefaultCluster().Resource(&item).Update(&item).Error
```
#### PATCH 更新资源
```go
// 使用 Patch 更新资源,为名为 nginx 的 Deployment 增加一个标签，并设置副本数为5
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
#### 删除资源
```go
// 删除名为 nginx 的 Deployment
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Delete().Error
```
#### 通用类型资源的获取（适用于k8s内置类型以及CRD）
```go
// 指定GVK获取资源
var list []corev1.Event
err := kom.DefaultCluster().GVK("events.k8s.io", "v1", "Event").Namespace("default").List(&list).Error
```
#### Watch资源变更
```go
// watch default 命名空间下 Pod资源 的变更
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
			fmt.Printf("无法将对象转换为 *v1.Pod 类型: %v", err)
			return
		}
		// 处理事件
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
#### Describe查询某个资源
```go
// Describe default 命名空间下名为 nginx 的 Deployment
var describeResult []byte
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Describe(&item).Error
fmt.Printf("describeResult: %s", describeResult)
```

### 3. YAML 创建、更新、删除
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
// 第一次执行Apply为创建，返回每一条资源的执行结果 
results := kom.DefaultCluster().Applier().Apply(yaml)
// 第二次执行Apply为更新，返回每一条资源的执行结果
results = kom.DefaultCluster().Applier().Apply(yaml)
// 删除，返回每一条资源的执行结果
results = kom.DefaultCluster().Applier().Delete(yaml)
```

### 4. Pod 操作
#### 获取日志
```go
// 获取Pod日志
var stream io.ReadCloser
err := kom.DefaultCluster().Namespace("default").Name("random-char-pod").Ctl().Pod().ContainerName("container").GetLogs(&stream, &corev1.PodLogOptions{}).Error
reader := bufio.NewReader(stream)
line, _ := reader.ReadString('\n')
fmt.Println(line)
```
#### 执行命令
在Pod内执行命令，需要指定容器名称，并且会触发Exec()类型的callbacks。
```go
// 在Pod内执行ps -ef命令
var execResult string
err := kom.DefaultCluster().Namespace("default").Name("random-char-pod").Ctl().Pod().ContainerName("container").Command("ps", "-ef").ExecuteCommand(&execResult).Error
fmt.Printf("execResult: %s", execResult)
```
#### 文件列表
```go
// 获取Pod内/etc文件夹列表
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").ListFiles("/etc")
```
#### 文件下载
```go
// 下载Pod内/etc/hosts文件
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").DownloadFile("/etc/hosts")
```
#### 文件上传
```go
// 上传文件内容到Pod内/etc/demo.txt文件
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").SaveFile("/etc/demo.txt", "txt-context")
// os.File 类型文件直接上传到Pod内/etc/目录下
file, _ := os.Open(tempFilePath)
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").UploadFile("/etc/", file)
```
#### 文件删除
```go
// 删除Pod内/etc/xyz文件
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").DeleteFile("/etc/xyz")
```

### 5. 自定义资源定义（CRD）增删改查及Watch操作
在没有CR定义的情况下，如何进行增删改查操作。操作方式同k8s内置资源。
将对象定义为unstructured.Unstructured，并且需要指定Group、Version、Kind。
因此可以通过kom.DefaultCluster().GVK(group, version, kind)来替代kom.DefaultCluster().Resource(interface{})
为方便记忆及使用，kom提供了kom.DefaultCluster().CRD(group, version, kind)来简化操作。
下面给出操作CRD的示例：
首先定义一个通用的处理对象，用来接收CRD的返回结果。
```go
var item unstructured.Unstructured
```
#### 创建CRD
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
#### 创建CRD的CR对象
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
#### Get获取单个CR对象
```go
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(item.GetName()).Namespace(item.GetNamespace()).Get(&item).Error
```
#### List获取CR对象的列表
```go
var crontabList []unstructured.Unstructured
// 查询default命名空间下的CronTab
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace(crontab.GetNamespace()).List(&crontabList).Error
// 查询所有命名空间下的CronTab
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").AllNamespace().List(&crontabList).Error
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace("*").List(&crontabList).Error
```
#### 更新CR对象
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
#### 删除CR对象
```go
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).Delete().Error
```
#### Watch CR对象
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
        fmt.Printf("无法将对象转换为 Unstructured 类型: %v", err)
        return
    }
    // 处理事件
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
#### Describe查询某个CRD资源
```go
// Describe default 命名空间下名为 nginx 的 Deployment
var describeResult []byte
err := kom.DefaultCluster()..CRD("stable.example.com", "v1", "CronTab").Namespace("default").Name(item.GetName()).Describe(&item).Error
fmt.Printf("describeResult: %s", describeResult)
```

### 6. 集群参数信息
```go
// 集群文档
kom.DefaultCluster().Status().Docs()
// 集群资源信息
kom.DefaultCluster().Status().APIResources()
// 集群已注册CRD列表
kom.DefaultCluster().Status().CRDList()
// 集群版本信息
kom.DefaultCluster().Status().ServerVersion()
```

### 7. callback机制
* 内置了callback机制，可以自定义回调函数，当执行完某项操作后，会调用对应的回调函数。
* 如果回调函数返回true，则继续执行后续操作，否则终止后续操作。
* 当前支持的callback有：get,list,create,update,patch,delete,exec,logs,watch.
* 内置的callback名称有："kom:get","kom:list","kom:create","kom:update","kom:patch","kom:watch","kom:delete","kom:pod:exec","kom:pod:logs"
* 支持回调函数排序，默认按注册顺序执行，可以通过kom.DefaultCluster().Callback().After("kom:get")或者.Before("kom:get")设置顺序。
* 支持删除回调函数，通过kom.DefaultCluster().Callback().Delete("kom:get")
* 支持替换回调函数，通过kom.DefaultCluster().Callback().Replace("kom:get",cb)
```go
// 为Get获取资源注册回调函数
kom.DefaultCluster().Callback().Get().Register("get", cb)
// 为List获取资源注册回调函数
kom.DefaultCluster().Callback().List().Register("list", cb)
// 为Create创建资源注册回调函数
kom.DefaultCluster().Callback().Create().Register("create", cb)
// 为Update更新资源注册回调函数
kom.DefaultCluster().Callback().Update().Register("update", cb)
// 为Patch更新资源注册回调函数
kom.DefaultCluster().Callback().Patch().Register("patch", cb)
// 为Delete删除资源注册回调函数
kom.DefaultCluster().Callback().Delete().Register("delete", cb)
// 为Watch资源注册回调函数
kom.DefaultCluster().Callback().Watch().Register("watch",cb)
// 为Exec Pod内执行命令注册回调函数
kom.DefaultCluster().Callback().Exec().Register("exec", cb)
// 为Logs获取日志注册回调函数
kom.DefaultCluster().Callback().Logs().Register("logs", cb)
// 删除回调函数
kom.DefaultCluster().Callback().Get().Delete("get")
// 替换回调函数
kom.DefaultCluster().Callback().Get().Replace("get", cb)
// 指定回调函数执行顺序，在内置的回调函数执行完之后再执行
kom.DefaultCluster().Callback().After("kom:get").Register("get", cb)
// 指定回调函数执行顺序，在内置的回调函数执行之前先执行
// 案例1.在Create创建资源前，进行权限检查，没有权限则返回error，后续创建动作将不再执行
// 案例2.在List获取资源列表后，进行特定的资源筛选，从列表(Statement.Dest)中删除不符合要求的资源，然后返回给用户
kom.DefaultCluster().Callback().Before("kom:create").Register("create", cb)

// 自定义回调函数
func cb(k *kom.Kubectl) error {
    stmt := k.Statement
    gvr := stmt.GVR
    ns := stmt.Namespace
    name := stmt.Name
    // 打印信息
    fmt.Printf("Get %s/%s(%s)\n", ns, name, gvr)
    fmt.Printf("Command %s/%s(%s %s)\n", ns, name, stmt.Command, stmt.Args)
    return nil
	// return fmt.Errorf("error") 返回error将阻止后续cb的执行
}
```

### 8. SQL查询k8s资源
* 通过SQL()方法查询k8s资源，简单高效。
* Table 名称支持集群内注册的所有资源的全称及简写，包括CRD资源。只要是注册到集群上了，就可以查。
* 典型的Table 名称有：pod,deployment,service,ingress,pvc,pv,node,namespace,secret,configmap,serviceaccount,role,rolebinding,clusterrole,clusterrolebinding,crd,cr,hpa,daemonset,statefulset,job,cronjob,limitrange,horizontalpodautoscaler,poddisruptionbudget,networkpolicy,endpoints,ingressclass,mutatingwebhookconfiguration,validatingwebhookconfiguration,customresourcedefinition,storageclass,persistentvolumeclaim,persistentvolume,horizontalpodautoscaler,podsecurity。统统都可以查。
* 查询字段目前仅支持*。也就是select *
* 查询条件目前支持 =，!=,>=,<=,<>,like,in,not in,and,or,between
* 排序字段目前支持对单一字段进行排序。默认按创建时间倒序排列
* 
#### 查询k8s内置资源
```go
    sql := "select * from deploy where metadata.namespace='kube-system' or metadata.namespace='default' order by  metadata.creationTimestamp asc   "

	var list []v1.Deployment
	err := kom.DefaultCluster().Sql(sql).List(&list).Error
	for _, d := range list {
		fmt.Printf("List Items foreach %s,%s at %s \n", d.GetNamespace(), d.GetName(), d.GetCreationTimestamp())
	}
```
#### 查询CRD资源
```go
    // vm 为kubevirt 的CRD
    sql := "select * from vm where (metadata.namespace='kube-system' or metadata.namespace='default' )  "
	var list []unstructured.Unstructured
	err := kom.DefaultCluster().Sql(sql).List(&list).Error
	for _, d := range list {
		fmt.Printf("List Items foreach %s,%s\n", d.GetNamespace(), d.GetName())
	}
```
#### 链式调研查询SQL
```go
// 查询pod 列表
err := kom.DefaultCluster().From("pod").
		Where("metadata.namespace = ?  or metadata.namespace= ? ", "kube-system", "default").
		Order("metadata.creationTimestamp desc").
		List(&list).Error
```
### 9. 其他操作
#### Deployment重启
```go
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Restart()
```
#### Deployment扩缩容
```go
// 将名称为nginx的deployment的副本数设置为3
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Scale(3)
```
#### Deployment更新Tag
```go
// 将名称为nginx的deployment的中的容器镜像tag升级为alpine
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Deployment().ReplaceImageTag("main","20241124")
```
#### Deployment Rollout History
```go
// 查询名称为nginx的deployment的升级历史
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().History()
```
#### Deployment Rollout Undo
```go
// 将名称为nginx的deployment进行回滚
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Undo()
// 将名称为nginx的deployment进行回滚到指定版本(history 查询)
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Undo("6")
```
#### Deployment Rollout Pause
```go
// 暂停升级过程
err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Pause()
```
#### Deployment Rollout Resume 
```go
// 恢复升级过程
err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Resume()
```
#### Deployment Rollout Status 
```go
// 将名称为nginx的deployment的中的容器镜像tag升级为alpine
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Status()
```
#### 节点打污点
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().Taint("dedicated=special-user:NoSchedule")
```
#### 节点去除污点
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().UnTaint("dedicated=special-user:NoSchedule")
```
#### 节点Cordon
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().Cordon()
```
#### 节点UnCordon
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().UnCordon()
```
#### 节点Drain
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().Drain()
```

#### 给资源增加标签
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Label("name=zhangsan")
```
#### 给资源删除标签
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Label("name-")
```
#### 给资源增加注解
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Annotate("name=zhangsan")
```
#### 给资源删除注解
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Annotate("name-")
```


## 联系我
微信备注kom：<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAL0AAAC2CAYAAACMEIzzAAAKpWlDQ1BJQ0MgUHJvZmlsZQAASImVlwdUk9kSgO//p4eEFrqU0JsgnQBSQg9FerURkpCEEmJIULEhsriCa0FEBCyAiwIKrkqRtSCi2BbFAlgXZFFQ1sWCDZX3A4fg7jvvvfPmnJv5Mv/cuTP3/JMzAYAszxQKU2F5ANIEYlG4nyc1Ni6eihsGBKAKiAANMExWhpAeGhoEEJnVf5f3PQCa0ncspmL9+/P/KgpsTgYLACgU4UR2BisN4VPIGmUJRWIAUAcRu/5KsXCKOxBWEiEJItw3xdwZHp3ixGlGg2mfyHAvhJUAwJOYTBEXABIVsVMzWVwkDskDYSsBmy9AWIiwW1paOhvh4wibID6IjTQVn5b4XRzu32ImSmMymVwpz9QyLXhvfoYwlbn6/7yO/y1pqZLZM4yQReKJ/MOnakburC8lPVDKgsRFIbPMZ0/7TzNP4h81y6wMr/hZZjO9A6V7UxcFzXIS35chjSNmRM4yJ8MnYpZF6eHSs5JEXvRZZormzpWkREntPA5DGj+LFxkzy5n86EWznJESETjn4yW1iyTh0vw5Aj/PuXN9pbWnZXxXL58h3SvmRfpLa2fO5c8R0OdiZsRKc2NzvH3mfKKk/kKxp/QsYWqo1J+T6ie1Z2RGSPeKkRdybm+o9A6TmQGhswyCAR9Qkc90IEBIjGiRmLNKPFWIV7pwtYjP5YmpdKTDOFSGgGU5n2pjZWMHwFS/zrwOb/um+xBSwc/ZshUAcJmqeXTOFnkTgPp3SOs9mLPpvgdAGemvNh5LIsqcsU33Egb5FZBDMlQH2kAfmAALYAMcgAvwAD4gAISASBAHlgEW4IE0IAIrwVqwEeSBArAD7Aal4ACoAkfAMXACNIMz4AK4DK6DW+AeeAj6wRB4CcbAezABQRAOIkMUSB3SgQwhc8gGokFukA8UBIVDcVACxIUEkARaC22CCqBCqBSqgGqgX6DT0AXoKtQN3YcGoBHoDfQZRsEkWAnWgo3gBTANpsOBcCS8FObCK+AsOBfeBpfAlfBRuAm+AF+H78H98Et4HAVQMigVlC7KAkVDeaFCUPGoJJQItR6VjypGVaLqUa2oTtQdVD9qFPUJjUVT0FS0BdoF7Y+OQrPQK9Dr0VvRpegj6CZ0B/oOegA9hv6GIWM0MeYYZwwDE4vhYlZi8jDFmGpMI+YS5h5mCPMei8WqYI2xjlh/bBw2GbsGuxW7D9uAbcN2Ywex4zgcTh1njnPFheCYODEuD7cXdxR3HncbN4T7iJfB6+Bt8L74eLwAn4Mvxtfiz+Fv45/jJwjyBEOCMyGEwCasJmwnHCK0Em4ShggTRAWiMdGVGElMJm4klhDriZeIj4hvZWRk9GScZMJk+DLZMiUyx2WuyAzIfCIpksxIXqQlJAlpG+kwqY10n/SWTCYbkT3I8WQxeRu5hnyR/IT8UZYiaynLkGXLbpAtk22SvS37So4gZyhHl1smlyVXLHdS7qbcqDxB3kjeS54pv16+TP60fK/8uAJFwVohRCFNYatCrcJVhWFFnKKRoo8iWzFXsUrxouIgBUXRp3hRWJRNlEOUS5QhJaySsRJDKVmpQOmYUpfSmLKisp1ytPIq5TLls8r9KigVIxWGSqrKdpUTKj0qn1W1VOmqHNUtqvWqt1U/qM1T81DjqOWrNajdU/usTlX3UU9R36nerP5YA61hphGmsVJjv8YljdF5SvNc5rHm5c87Me+BJqxpphmuuUazSvOG5riWtpafllBrr9ZFrVFtFW0P7WTtIu1z2iM6FB03Hb5Okc55nRdUZSqdmkotoXZQx3Q1df11JboVul26E3rGelF6OXoNeo/1ifo0/ST9Iv12/TEDHYNgg7UGdQYPDAmGNEOe4R7DTsMPRsZGMUabjZqNho3VjBnGWcZ1xo9MyCbuJitMKk3ummJNaaYppvtMb5nBZvZmPLMys5vmsLmDOd98n3n3fMx8p/mC+ZXzey1IFnSLTIs6iwFLFcsgyxzLZstXCwwWxC/YuaBzwTcre6tUq0NWD60VrQOsc6xbrd/YmNmwbMps7tqSbX1tN9i22L62M7fj2O2367On2Afbb7Zvt//q4Oggcqh3GHE0cExwLHfspSnRQmlbaVecME6eThuczjh9cnZwFjufcP7LxcIlxaXWZXih8ULOwkMLB131XJmuFa79blS3BLeDbv3uuu5M90r3px76HmyPao/ndFN6Mv0o/ZWnlafIs9Hzg5ez1zqvNm+Ut593vneXj6JPlE+pzxNfPV+ub53vmJ+93xq/Nn+Mf6D/Tv9ehhaDxahhjAU4BqwL6AgkBUYElgY+DTILEgW1BsPBAcG7gh8tMlwkWNQcAkIYIbtCHocah64I/TUMGxYaVhb2LNw6fG14ZwQlYnlEbcT7SM/I7ZEPo0yiJFHt0XLRS6Jroj/EeMcUxvTHLohdF3s9TiOOH9cSj4uPjq+OH1/ss3j34qEl9kvylvQsNV66aunVZRrLUpedXS63nLn8ZAImISahNuELM4RZyRxPZCSWJ46xvFh7WC/ZHuwi9gjHlVPIeZ7kmlSYNMx15e7ijvDcecW8Ub4Xv5T/Otk/+UDyh5SQlMMpk6kxqQ1p+LSEtNMCRUGKoCNdO31VerfQXJgn7F/hvGL3ijFRoKg6A8pYmtEiVkIGoxsSE8kPkoFMt8yyzI8ro1eeXKWwSrDqxmqz1VtWP8/yzfp5DXoNa037Wt21G9cOrKOvq1gPrU9c375Bf0PuhqFsv+wjG4kbUzb+lmOVU5jzblPMptZcrdzs3MEf/H6oy5PNE+X1bnbZfOBH9I/8H7u22G7Zu+VbPjv/WoFVQXHBl62srdd+sv6p5KfJbUnburY7bN+/A7tDsKNnp/vOI4UKhVmFg7uCdzUVUYvyi97tXr77arFd8YE9xD2SPf0lQSUtew327tj7pZRXeq/Ms6yhXLN8S/mHfex9t/d77K8/oHWg4MDng/yDfRV+FU2VRpXFVdiqzKpnh6IPdf5M+7mmWqO6oPrrYcHh/iPhRzpqHGtqajVrt9fBdZK6kaNLjt465n2spd6ivqJBpaHgODguOf7il4Rfek4Enmg/STtZf8rwVHkjpTG/CWpa3TTWzGvub4lr6T4dcLq91aW18VfLXw+f0T1Tdlb57PZzxHO55ybPZ50fbxO2jV7gXhhsX97+8GLsxbsdYR1dlwIvXbnse/liJ73z/BXXK2euOl89fY12rfm6w/WmG/Y3Gn+z/62xy6Gr6abjzZZbTrdauxd2n7vtfvvCHe87l+8y7l6/t+hed09UT1/vkt7+Pnbf8P3U+68fZD6YeJj9CPMo/7H84+Inmk8qfzf9vaHfof/sgPfAjacRTx8OsgZf/pHxx5eh3GfkZ8XPdZ7XDNsMnxnxHbn1YvGLoZfClxOjeX8q/Fn+yuTVqb88/roxFjs29Fr0evLN1rfqbw+/s3vXPh46/uR92vuJD/kf1T8e+UT71Pk55vPziZVfcF9Kvpp+bf0W+O3RZNrkpJApYk6PAihkwUlJALw5DAA5DgDKLQCIi2fm6WmBZv4DTBP4Tzwzc0+LAwDHPAAIzQaAgegqxKTXhsRFVijyPdIDwLa20jU7+07P6VPCrAHAznuK7rf4ZYN/yMwM/13e/9RAGvVv+l+i8wZHgSONagAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAvaADAAQAAAABAAAAtgAAAABua2P+AABAAElEQVR4Aey9Z5ecV3bvtyt0Vecc0Y1u5BwZwUxwOOTk0Yyke6Xxla8tX72yvZaXv4D9Be6yl9/4+oUte8m6axRGo4kaajgcZnJAIpDIaACdc6iO1VXVFfz776eq0QC6yQbDsnR1D1BdVU89zwn77LPPzif09R/9dwXbQgmFQn5XobCl27dQ4z/tWzTOkP4xbn/R3VAobGGu6b+gsPZb8Z7SiEIbgiiAX/Bk6c6tvJee28q9W7mnYMEcahDr66bTG3d8K5V+Yffk8/cAr4h3n9bAvXhZwteNnotudPE/XzMLg+BrSB0uIrsWvCZFyBK+A6V8Pu9fSoC+910//hPApzsd/hf+aVOkL01cCT73rqTS9f9U3p2y30VV8lBzUULRe+E630TRo3wrgOROkEImdA/u0528inXo8xpF5fOdEtR35/uX/alE2b/sdr6g+u+agwerMyygr5V7doy162abIv26e/5lfCzCKAxVd+SGNOvdFwMQ2AiEoFPxh+DXgPMrPVNii5iJUORfBgz/mYzyU5H+P3UKX5qnsKj4Pf/AeL9SwnjBwheBqJGT8uDpUFjoX1wWXo+uB1e8Bv8JfqjYRqnN//yu3VCvIuwEEMH2Sy7rONNPbynomjr15Xes1Bu1WeKH14Gm9PPd79zwyX2802/dd6c+UWX4dqh8RPy7vwJKXWpAVD+4fjfIvG9UtB4qYvnXXt5OsFgKsEUFGKLSEikti1IbD/bO03eG82CPbnL3Hdith01wcwlWpfd7qyisIasGHzwfXPuCO3lfwzRGEw/SyqaU/q7VR0MFn10QIQcfG85bPp+FtdXoRMG4IRQIc/f26fN+F5AjWViGWIB0nzTT3KWOrDW5xms7JQl+9RUk1qUQ8THpblAdwVXACxBSdawRn2J1If89qLpY01pTwS3FG9WFdSVAdF3MuSggShYOAXbe9c/Hs4YwerDU63WVrP+o54ptCKnUT/+ua6pureOlh9SvYt9KlzZ6L/WB57WYvE7Vp3/8Fs4H9RTY1dTo+hoduYv90KKWBiZXrMDRotThtXbV2aCv6/vr7ORaxWsf1p4KPujZ0vPMHG3lY2GLMYf5+9q559Hi102R/t7bC1QayeUsH0lbIWfWWNli2yqbLcYE5iJ3JuLe5z7vd40jbRmbXpm12cUZBsaElCbonsrzUGhhViGkBZln4qKsSQGPz3wvsFh9QnNREJ2K+UmLKbg/Sr0xrgb33jWr97TzIF/VVQnBOSf9eQsDO01OOAzoeQlhAnF4fa3q8/1FCChNkaZcT4WzXNH49ILmbHHO76+YGgWvgIax0+WE1AXLeoVCeLWsPgsHwuAAVQgJikW/hzRGDRZ8aK6ps/Z4g8WAf9b7FyBpad5KiK427yqChc+Xrm4Mg+D+oL5cPmy5bBrcmLap1AxPbA2dt3YXLYWh8FpheTq1p26P/X7PC/ZkzxErj5YF/fgS/xayOTs3dtN+eOtNu5a4atl8ZmPEZ0UI7haJ2O6GXdZW3qge2/jylC2kl62rpt0ayutsZmXJxpMz1lJea63VLaLzNrQwzrVJSxfSwUgCzLp7VEzsgxXQGeQoj8ass7rTWqqbQI6wjSxN2ejSKHiToTowyDsd1Cxk3qyUhcuss6rVOqqafRFNLU7bUHLMstImlep40C6qMVi7xni97arttspYlSVScza8NA6saq0D4qaFNpqeskFgFMplHcHv9JEx0jZLBLBH7Wj9fvtW9yk73rXXotrRvuSSyaXt7OgV+2Hv6za4PLKl1iJ7//Vj//NW7hQXU4gUrCpXbf/Dwz+wx7oOWIxBijoEvEBAHbQw9E/vn+UliqF/AqPqFVtVgMJ01jVbMwh7I9Fvc5kF6gY5gv98LhbntPJWFaq2Pzv2PfvBgZft2W0nrS5Sb1XhWvv9fS/Yd/c8w0Q280DBTm8/Yd/j+3NdJ3wH6JsfseXMUtB2CYmKVa//qt5B/4q/aJzqiHYMLbrg3ZGQyzl2xx0g0x/t/6b9q90v2rOdJyDMIeud67eVfDJ4zpnzAF4au0h28K/YhOoHFm0Vjfa9Xc/aHx/4qj3acdjKohG7MTtgyUwKBZEmKLjfdzN95LuzpTyu4n3Tu3/Tn6AVzePDbQftT49+y17oPsECbbZsNm9f7XkEmD1tj247YJUs3AvjvTBp7KJicagkqIcxq11e21mQ/+boN+xE+55g8/GO3GktwBM9t/bwWk+2+sHhw82qVc1G2S3ba1qsubLJzk5chSeAIDrLrV/Xtc23UinNXOn7pu9CQthgWJpW665tReCDJ1a9atrHIIxTdbJZ8qLhz/JSHRIqVU+++B4Mz2xHfavtqOu0CNclT9w7pNI2XMbuUxGGy/OValYRL7P68rjFnP2hZnar8ngl73GoWFBLdSRmZfooWSUYEv2gjeLLIazRaqzcEMpBXXlf6wX9CetBqKJYKX9BgXVvORSvkslx5KCOKCrMKPDTd9UTKqxSD+3qFYKF5F2vQoFXXvwQ8gA7RqQMHpaxiWeOcF88VEZdsGQwzhGvRx1kfEWuweeeS96O3rUh8O7FL6oD9JtFSE1Fvt2AQ8RqyiosHikzugAcI1YJfCTka/xh2IqIdn4e5z+vgkXpV2d1u3Wwu3orPOPsTNCEtytk92tqu/jkg76r/6UxlN7LgO2++k7bXtkBbLLMgwiP2ti4bH3/EcB4RWEdShRDqFlatelcylZWM07/xO0JOJ+p0NkIFUTKYlYRjQf1r6uonC1eg8ppS18bfukGGmWs6ou25+aKBvjSnN1Kjtpgatx2r3ZaEztVIp2wqYVpa4nX2hL3rmayNpactfTqKmOTgKv61gGt+FFveZArl1+lfZCwUGZlLK4cbWS5pkfCUF+RuTwsGTc6Pz+XW7aR1KR1ZBqdL56BfUhmVy0lhM6t8kwIJifqi9R3uiLsHP21ILSIomFbXF2xmeSKza+mbbWQssmVBVvm9zRYnsmnqUH/EOhYaKuqm4EIIUQ7Vumv2BS+BQtOI/R2QvQ9b3OrSzaambVUNG8jyE9DyQmrTpZbbUWlaUn2Lc9YmjFmYW8YloXBA9XlC526ROrAjDWwCVYhFmeeXWEZGDtsgUsgFAtzHryIPunJWLjcKmjfaynCSngo3HDWo3htsxa2jvQaD6+ShOx4wPdVtu8PRq7aT26+aTPOGvgyWIcyfudm7Rev3+llyKmj+MQy+4MDT9mTXUesCqrjhaqgecHIhfOq+s6jAJjvQGYlu2I/7X3fzg4NsEgLNpwaAVEWrTcxZJVlcUuuJm0hs2xX53rtlf73fGBz6QW/HhLSQi2Csq5yLgj5yqDQDbEai0fLaScDO7QCsONWFY/Tr5AtpVcsA8JXlZc7dU+xkGaX5+2vrr5iP7/1NhS7YIu0nc6krS3aapXlFSyApC3mFum6qL2K76tQ3Ch8dSWTXGYpCMrC0rz9+OY/2qtD73IfbadBJgS59rJGkIBxgfgpFkaEhVtbVsXCNEtkYKG4u76s2ne6FItkiT5rHjU6vXKMV2zS31z+rVWHq2waAjCUGrOhuTG7VtlnGfo8sjLhu2tTWQMwVJ/T1LMEzNmVqET9dgG1CDK9LbE4PwQ3fjnwPjBZDeYHAqFZ/CzFWV16XF1WY9/f/aQ93HoAOaIkU2q/4R844GixEU0sNrp1pPdR3emqWAkttaV00j6YumyXE9ehsFA7Fe2v/Bd7UxBL8WnFtyJpJbTVs2jyUajKil2abrLjLbvYaivXahBrAjkOqFSparoSAFyNZmEnKuzlnSftZMthf+7MzDX7Ve9bNjw/CkLocdqgniTIN5mf9ntKenpRCvXBsaHYqqvk+Kxt/vGOY/C6T1hbrMkGEXzfHrtgu6q32aNtB6CuWXtn9CMbTU7bEx3HEfg7bWI5YT/rf9M+mLhoUytoGKhaoHyq/YR9Z8ez1lBZg1A7Yz/ve9vOTV0tso05K4OdeLT9uL3U/bi1VtTZzZlh+0X/23Zt7ib1jNO/Aosibg93HLFv734GbUmjDS1P28WZPmutrLeHWvZAXTP23uhlm0jN29d6Hre6eLkNc88/sNDPjl21WBTtEf+0G+yBPfjmjifRytVZ39yovT922fY39Njx5l0wWln7aPK6XZ8fsj/a+xWE3XIbWZignjN21vss1k3Yxi7J2FQne5zNpxbtzMRluzZ7M4Ckxi6kD7OjcmNek8G7dreNi3RsECHd5AuF+1lhYfBrf2O7HWneA9tVQvrgLq9PT5QeW19xsLWxQ33OkoP6ZQBKGEFKFIZu8YrTqHh6VHT8XnDefLOGAooDZ0l/eUWoi4FloVpZGEpt1VstoRxbO7CPVVTZzsbttrOp04E7uboAdaq0SHoOyqAeBkUI6DAtXRDE/L92qzvtFlXTkpqssbrBumo7YI3qYX7DtmOx3XrgZbtqOtj1sratFoSMRbinzTpr2pxKN8SrmShkFOfF2cX411hVb9vpY536xat+vI4JBRZgg5Z/HKTvqGixnppt1lhezU6QserqWossBoiqVS94N8ZrkHO2WUNZLfgQsgXYlLaqJvjrNmCYhULPWG4hZ+21zVYbox/sCA0IfaWFDLVwIbiOhdLT0EFbNVD+HEg95fX21G+zVRB1mp1yOrsMTHucjQlHyq1p6pYVJkUk7sCqBEpHfOY+XcgAc8lpJbiDcrAnKDitjN+dcpYeuu+deaBu4Y9acMiI9xU7SXWiu/cWp5+lpu79sfj9cyO913NX47QIsuaYpCxbn6/s4grbuA8MHKobhZ2JIJQFWhEtUy0e/du8CNb6vQRQUesIwI2xU8igFYVLybPwIkBCqOYYzSftCqVnNq/9zi/qvoaoDSzCZuZwp26nbrSnyXDGi0mMwFPEV9W+KB/tqz3uCXYXrvFfRMv7rRv4HKEBbWBu9JNMQXWOlOq3f0FQRCcfW+XmdQimCUZz7rKAuiDBPZaNWZTx67t2zFi+wsozCPH8Fg3whSfuRlT1JUw73hafveNcdBgxzqhg6nWLh+cHtct7SbAVe7N5ufvHAuxVlh1oFYOF7CbBDiGgbFRoiHuk+gwEf3quMXB7oOzY6JlPv/bFIP1d7dBRrBeHWg5ZDyos57lEKjcpctrNQFluw3oMzY9DGVY2uXOjy6CbEJgJCNgTaVDy8JKL9svhD+06fKhQ8vrsbUtkReUlbvEPDHN2aKMqN7om2NOGKOD5qdvsa+WoT2ttDKPIldleKHK99S6PmsjP5ZkbNpOdtbncgrXPt2IjSNj1mQFHblF7R1T+fDRxA+SBUlfV2eR8wi5P33Ik0u9ya15JpWBBrlsKAtJI/eMLk3ZtacgXsWCmVZNl7Bem+yx04zewQLU2Cetyc37Yaiqq7eDiToTOjF2CGosFXepNWV15PazRrF2e6KUtDRTYgdCrCKjXUQX//PY7CK612Cwm7cr0betbmrTr9cPO3tyaH4RtG7fI5V+x2zF2DIUXx+kzAmXkk7H+LojGkH/21HXb3sYeRhFjoWtRbIwfmtcsclPf4rDdon85hGy6+7nL50R6dfbOy2keov2Bpt32Z4e/bXsau1ml0vZs3k89A+htAKT/3z/+EXw8yACbokllzP76xKfZ6prLW+Hxeqy5Cm3M4oJdme+zhVTSJubnQNasza8krTpSa0cbu6wZtmR0ZdJuLwy6IFZCfnczWN/Q+j6Dia4CA8kGlwbhnQf9zoDoIigvDNmHkxe5poe4lx/GMRypBPfwiyOZ/8yfMAiEgWpgSrcEhRsDZBY0NO6sDSz32cDt/gAW1BtFoN3VuMt2VHWxkxp9GUUgn7cZjG9Z5KlZxrySWgUJMzYTX0Bjk/HvI0sYsebfoN5gUMFOJzaJK0yOZnCEBTPS94b3RerRcvj8ehb2dGYOjVXOllNpm1qctV8svkbftOhgwuizGNq8PhdfXoH+MHliP3IR/abv2nEjtru+x36w/6t2uHUnlzB0OVyCfvlj2tnWivZncAOj2J9f/YWdHb9g5RGUGtyiXfdTizd8/12fE+nvr1BbfXNNvVXHKkB4ppGVfGdI99+vGdXvDQhQdfClEUzXgGmjG++7prq13T2782H7/q7nrRmtSpKJ/n/P/9Ke6DpqRxHmRMU+QMjumxyxpzqPOp96A+r4N7detfPT1y0DsnxaD0sN+3Zf+sJ7sJjXjy7o9/r7Nl/w3LuFYYok6L/kgG2Y9r/f9aQ9uf0k30L2u6HLNro8aad3PGQtFU02wkK6giArnv5Y424WQs7eHj1v/8/ln9psdpH+llYenVe1UOhSX/Ve+pwDW5uwHv/hwRftRN0ukBYNHQLp/3ZuEPimfdwb8fHrQHP/R9rTTleDNqqJ3U0LH2GGUfDDuiKV8V0FxK1CFmnBMBcSI4827vOWLxzp1SGPJAIZH6iILEi3fA8QNqvDBWcBTgtL26tgIUOCBOA46jEEYvHIBXjRCLytG1/8ftiCMvyFpJrcHCM3a/b/v+uMbxWeJEO/EfUDXYZ8nriWh8XQ0LSrZEVhNS7Hb1gy7knDP0sdKrZJJUBuwQyUc7gF14Mvqoelho+NiC7gc6FxldulbaEa983R9QcqPC/qJvatpB0LiLVYznXFO3Tnuy96uAcZI9XVL6J8OUjPKKQyEi5+Omwc6o6gOQk30LBPeyqYU4wTmM8FfPmKjCyOoQtuQHeNDn4RVWR1wqbL56F2eZvEz2YiPYMBagZVYJlNYphaghWQoUVsgWYjIt5S/UYY1U4jxHA343spTxHqQa+5Xz3WM74tSyBlkd2jptU9kgfcK1Wj43c5b9HcBiW4qvrFZug5l1v4nkHvnUjO2zSGIu3cM8gU0zhaTScTFgUQ0xirpvhNdoZWtDo5dOMzaGFkpRSj4BhHxaLkqlP4JQSU81uElwvPus54ZOAbW5mylqUans9hzEMLhACaA15ZJzSMgfdgAam3WytiqdR3FS0AldJ3/3IXULgBuErgFRHTe4AbD0pRvea1P58d6dc6xwcQRNSBufUiM7ULlnxzYH8i6gcVhVklegkQQhKZxl3wW+tq6YOus6LwDdEWK7vghclbGEtQyyFgjSzO2222+NssgleHPob4h2x8ZdhSOCZNZRLWXtlmQ0sTNogMURmtsCZUdVqcMrSAYhjCKvFpiaF+TNtKMunGN3W/tPWXeuEYKzzi2SpUkuVFA9pKOo3BCGFcuw7Fx8KCqohXYUAq54rsA7onw2/BGLwuv9t/1uh9bPFwBTrxCl9EMkLNLy/aL/reQjd+BXjnbQL9/jLGqMsSXjE+LaIXT6CWlVPdjcUhV6FKiF4BgUtChRBbtLUCY1pVrNIRPEV/ZE3XFHphTNPpWXuFtm7U9CEbwPMvjloeP43meJtFyyK2upqlbQgH2hjof/HBjd8EO71YJtypsZVIoeZZz6x//u7PYuOCZ0B3uXfq5eozwbfU4Y3b3ezqZ0f6zWr8Eq9r+LKIyhViVewQRe4IOxs67Ynu49ZR02RzC4v2U1lbE32Yzge132OECdv+lr12queY686vwdNrnz7atBOr3kGQPG6/7nvPFnJJO9V13LZVt9rk0qz97Orrro0RtfFSmg9Ikz7KFaEOV4aX9+CU1X4Yiliw1wfO29tDH9iyeGj+iVK3VrXYV3c8ZUfb9vvkv9V/3n4zcCawAAvTigvEJ5+6XYcN5T2BwevFnlMs1EY0QEP2y5tv24252zYMAqrIR0abyvWZqw4HKcm213ba4z3H7XDLbuQV9EyxuL05mMC5DcR3JMlbPfzx6Z6n7Gmc31K5FTsz+rG92vc7W8QTVcipBVWDRffhjoN2GAPQKt+vTlZbU6Lavr//JatDO9SPpu3Xt96zj5CLxF39cyr/zJA+2NxcRVmEMsoZ21u1zR6p241qr9ay5QU7M3jB+sPDCNLVsE0gPTi7J9ZtJ6v34ZbbZNFkyBYbUtZTux3PwA4MV+W2E42I/HQOVnYjENdbU02lvYfAFZ6DZxbSlxDe2wWdQVY5hdUhZB2o2m4HKrrc/WCgatTOY/lcwiDmuwP3NcZqbT/17o914WcTtoGaMauIxXBjWAxoVYnE0oaakVYjTp+6MDDtrelyF1/5rrTgpz6wjM5a3WF7EhzELsjGIT5exrwWLMWP1Oyzg1U9znat1mAVtY9shagEVS4WqzZKn6l3D7veCh6aUxVTVoNL8UJaHqbszSye2midPdRw0I4iyGrhxtOwFKmMHWna4WsnDsW9WNVo5yehxSw8sXilYRSn5ot7uwv2n7/aLwfpnVkDcipi2B6k02x9YglE7Rz6QS13/qraYtWqW9Xj/sWkc5HPKUdFHtUXkLgQEg/KzhDSL9TJ5SxYkyluyyXhLge7tBLCoCaqyyRm4R+z7mrFd5lzvPqgcakTVZEQL89OkpcZGJKLiKGrbAHqLuPgo5AyCxa5jrloMQ8pSIOXxiiEEX9V4m/h8VgsPKdnQLYgAkmgoPcS6MQaOHvAYtS9XPd2qACRiDGq38EYxLfLWSzn99E32pHGRqyonOY0zkBkYQzSgTL2gBVRuyvAjIUiKq4FRUfTtCGpSwZAqXBzsJTqk9gWLtIvH7AGf6dwKRgcHzRI39UE0wco1OsERID5AsqXg/QARtMdlM/QUR51DcL6au4bLPWCmFFev5v4yKaXElaNu/AilOva4i0MXilHNmbPefZzM7csce3n1hSpwTtxBoeqUeuf7rf3Ri4gEJdZP74tmXDKbs8NQ+XQ9+OodTMxLCzSlHqR9iLnCAOCyJsL9dlkZtL+FvXn2/QBlxICGYYx2c8gIGdAJJAJhBtcnLC/vvy6vVFxxcEyRLDDHAY0MAAcKC5xGhF7hIrd71niwxn8YxIry1aLzCAnsN75AVgWFi8LVLpyCaBhFoBcQVzYBRYKpPjL67+wbQP4xNPNQeSXFdg2R166LASdWEnYj3vfwC/mCkibtWF8eaZzs8gOWrIgNe8z6UX7++tv2QdD10H+Vb9nAl3+/AfYPPAsnWdXuLkwzM4V9F9/Ve6bbeAldknFtUoC6GcpXxDCq+kvCek/y6iKzwA1jc/HKGxbB0UHL0AU1dMkY1gnZC5sZYVyayMqqqO20SYSsxiLJu1I+0Hb0dAFXuZwhrtpo6lpgkeacA7rgGUJOeLXY6HcjjxQni/DgzFhFxcH3ZEKozsTpf5AxeiIhD9fhCDbnrouO9l2CN+bZhtMT+KIdQ0PxlrbWbsNKhrG8rto23j4oa79to3ghnG0J72JQSyd1bajsY06Czafn7FwUojLWDQ+8KASK+8x5AvZFrL0+aPpm/bR1HUXICM4huVhXWIg26n2h+1Y0z5HdOnkJ1IJO9m6G76/AT39LAt10BrDlRiBWi1JxTNYosPJIkoKroIpSF2No9uu5h40MllbmcoQFYUXJXuEcFeCYzq9ahcIyrgQuUIftV9FbHt5k3URT9HMs1pMg4tTsFr43JexuljcmisR+/VFYNQu4sMsLlAH7fqbHuCzyxxe28YP+Y6w8U9rV//JIX0JaC7rS0pz7F/rL3MCyISR/NfWLg/Ao0z6S0T5tDHxC+2Y3HHTfaz9kJ1oFXLkMNY04Dk4ZE9sk3GKsD387GNIX/uxDh7nniq0KvLMHMayOZ6bpOJAjan4UyGlqJ/ajOADsr9ZbSFcEiwxkJogIqsMC2m7PdKxH8pasAo0G1L1Pd1xwvYhMwzPjxEm2IiWqBmD0U4r4/c44xpYGLVMBvWiEIo2arF+PrrtsH2l6xGEdJyxsCXcwmqcwJ1Y41Q/GhnHY9sO2XPbHvLxN2GMG8Sj8/kdJ62T/gxhqNqGkaoVp7gTWjwwI7VoaAZYzBkXUn19udfm8wjsTxFVllxNoXmKWz8uBpNLY4wVfoa2sBFys9gpjR0FAnzQzoZu+4P9z6FIMDxDpyxJnaM8I98m7bqBSvHOXPknBx1/KFKF+tz5t8/6h859zvLlIL3EefFhjFXsLqgjOG5Y9Ivib91DEyRzHOdd6kp5aq5/UJuvX1DFvGQp1KTEy2NoKRRFFMbXGj92NBbSWojJlrBXhnouTlCKXGgVdSU/9TiqwDgqS666MxNV8E71TJ73QYYGb81nzSmYDGLleCnGZEmkK3HqqwZhysuoA6/IEBRNaskqhELV6wvXf6+0KjREMe5X2wo88fFSifqvccXApHKPTkJfDkWMh8oZj1w4oJIiz4y3jPYVtlfG9xysVVk8YtVZorJ4TvCK8nu8vMwq4xVBG8C9HHi4XpDdQ3ujBA7CcHDL4B7tlbSlcMEIi9F7ouHqNvWLD5J5tNPJXynKQixnHPqtHGE4Rt2yJUhxLADJSCiCf1fRXBVdERRmGIIdEvuv3U39oirKvQ/pWnCVpeRtS1OZxRqcB7IBLqmTwSe/mT9UF5Ri3Zr7teL4EnzTk19oUUTNrZFBIpdmPFon6Jr+bvyS5jYPUOSANYwaTGxLBOAqyFjmm3XdLvbzzpUQW2sE1iYxt4SxSYIXPi/zsDcYqxIEfxt8tRjtyeVZG52b9kAMVbiMkWd8dgqj1kIguALKWQw8Se4XpRNySGQLI+xpwYJ9DlF5B87DTiRR82kXmCMYY4B6F5bQeiDQae7n4cHHYbFW0KtryAoWkR/OJNFaabxOUwiHiULClkNLtMECY7415hQOZsuoWrMIyRleC/RxBQV5QZoZ1LSyYaSSPEv9GepdzWYtsbiITw3ObejqvS00PFMzCZucpS12i2VgMjM3h8AJ6ogQ0ZgW4jIszXSSOGOESkV9JVcWLbXEdxZ8rkxt6iNwoN08C1WVcxv3LGPgS4B8IYxkKZvGz0nPBBjM28a467gptkN2kOH5Gbs8NeRsls+vFtQmL7GWWcZxc3bMbgxLGwfbqflQi5u1xW/SkIVYWeF1ryDLgh4CBlvNWqwB5dF67K/cbf/T0/8ONVoNnS0gQC7an1//ObrgM97BgKcKQreEtp6iQr3cpLjqDTcBCUsZtCfqnNxW0wRNv7TjafvjfS9bO9H/KnNMzl9c+ZW9Pvoehhe5nEbQiqCSJGyvKoKBCA++TIYII5uFCvEdCq9glyTRTCm8N6NQ6Chbj6yKxBhZtdV6FI4WwjyC3Ar3iLeOQInTaHfSGKhUPOYSKqZdS4HIvjuAqMp0QOAelCug3pp/pSuR8FoXJrqKPq1Qj3T2MShmbbQKMEYI+1uErZjlPpBR6hNgq2AbeSCyBzgrs8pzCiqPESgSB2Gz9CVJ3TWhGqtHnSjX3CUWyXJ+ya3MZQU8FkHijPzXodDaUVRxRtQRoVjsm2QDBbokWVTsPVaPE14BOC7B9y8TfliBkC+hfhXjmmARhfBU4PsuoR0I0tec1Ycb2T0iRGmlUMuiLOCftEZRdsBTjcftT499yxpw/JNwPQJB+L96f2rnJjASgsAqesf5GTmKMMsy0XEZq/ynTf4o9DGL5ojgdxGWHCow4PvtPS8S0PJV9/HSgwkMc//+d//RLiQusktWshMqa4MIbamIMaY95u3LYW8YRJJJC0sVESzMUsv3vft4GbUovrQRpUWjG32r3AQgAqHy3OShho/Bn36P6KHttS2kqkjYDy/+CmPRIfjsQ8x7wd4Z+dh+dP1VhNdJ2hDFKVgLQu2/2vuyPUkklEL/foY2Yx7NyOmOkwSCNOLINWU//PhX9mHiGsgUqBaF1DmF5IF8HhHmfWNP0AJB0NQUlkGZTzYdst/fe9r2wAP3JcbtA6KOmmtq7ZG23SypMvsVBOLvb7+KXnyO/jmDB9HNszOyW1GnqJjC/57d/rh9a/dp0n4021U0Tb8dOo/nape90HXMVa6/HTxrf3v9FTRWS/jlSDkpoGiPgmhI3UiRv/3jbQ/bD45815qwKQwsjNlr4x+7O/S3kU3mCTl8e+yiTS5O2td2PYVhrgmntQH7x9vnEdh32WMde10L9t74ZfsYwfrfHv0mgnuVG6f+rvc1e3/8LFT1AdCIceJNz47HYiE+IHB9CBaEd3iDP76ZOKsrJPZRbnCXfuEf8y2DpQjA3YspQHrh1wP09t52pPUWG1CsDEro3Sn+kTr2zvd7n13/XTw0W5wYPR6SRiOMGs0zG2w+Pq9brIemuBHNSAOGqTj8dDsm+LoqvuO1KSMMuEgIXKPFYmhAZI2nz3LNrIhUonoLeF/x0k1oYKJ4AFaiHqxA+9GIYBivqmELoF+lsdBPoZGYhACgxckSgNVvAB6G164oh0IiQ4iHrya0rg1XB0Va1UKlJQ8o5td13QIDVZT0/g4x6pGqMxKBEiOo1iN3xKHaiu1tYuE0wK/HWFjS5DSwm8VhBefwsVGQup4PegTg9IEdM8xOonDLRl5aSNUItm0IzY1Q+TK+x4ksacKYtprHTQKhXDtENW02E9lVx6sSBJfKsZG+1GG0ayWaS6gnzVcd9ZRa5NK6osbFTgULuDhIYKYFTqFf/EoRg/PJRburXjTEf+EbGjXq8SwRvN9VtBtrN4CNNAiR7i8VRyUa0/uDI70/zZNCVP4JYZrL6kAqbcHQGP1erLzU4Ce9CwnloBXwXDzKlpRXAin4yQqCphWksGmhLfGaq2JXmHj5g+QzcI4YjLIoqZUBQBirzAGycgrZtUw0ibIgiqqKMsqhSr7nWerIQu1XoeZyuMrjHyM9uorGpU/+jbqkN3I3WN+d/Bb/o/tCuCTmqFxB8xnkhEw2RZ+onzZDcqpjYsTWSTsSkjejWBy2NbGCQgItZu1GKfqUpi9yj16lnlXkhAyfpb4sgDjS8hRQ2YZ4FfBHEVvlTmD8po7qTR9W2XHlf6MIJPVnlc9ZWEm5URTg9+WUtyoK7C/6TBtZ5JY87EsOOcFVq8BYBi0Fw8eZI/VFfdrMxVj0QSlEWrB5SPMjuaUAYfM+BR1zeH3aH9/t125i9hhnFOJWXVkNu/YJuEH7JcOjHnfSTB9UHhzpHZCqRZNlROlU2uldj8kGRDDDvEPaEaR0nzez+R9NMzi/VlwnTN+q0F68QPB0PVvypoU2wsSpXsUvpS5ywVV1Y8k5gqeHEPYKOJiJfcgRLI1hCucoCcdio3I8MwtrcX76ivOm4tPPzFz36P0UiN/ELjFDPQOLCNb8EyYKR32ha/KYxIJ0eqLsGrhmRmPgowSvXtSRb+F/3qfADPTZN4g6kgw0lJry9i9OXLdFFkKBdpUGz6kl/VQ8sRfql4PZ1Zl+tD7yJa8nimnULhGpNIPwmvJFnbeL073o/DUuJAE567EYfVrV1yJiCYFvLYzYr4fOor6sxflu0i7OXofVabI0/hkiFFeJkppGQDf4/pY4rg4EbZybuWkzeYRijG8iDoLpCNd/RtaLZsYyjI7+BtckLN5bdEXgaKystRf3PO5asyTGPmlvHgDfvVonosUGtPhUmqMN9mzj4UDeKP523xu3CtFL5c6nB0D6UuMeI1msSQyNFs9O9NR/dvQ7pfq/2Pfi6ixV6iDWCHRdLBXQXQCgA3OTtohmQ5H/i0tz9pvJPvsVmhsBSurBTvxMntx2zNpwuR3ExfgyWQVeG3gHRzPSaVCVeP2OinYbRbhdjiVhGRAS5QnJ81oomi2lF+xEbthPgEZ9rAGNzByISXggIXhBEUSQZ9DqDM2OW2YZz0gimybQ1Z9LfmS/FmXmn8IWuwgaP9x0nMDuWpvCXfjK7C0WqfIxBkW8aRLtyxhapuUIGiGsyDMEidxCxnh14LfUgrAKtWjHCHUIZ7omhNtZvEhvwI+PEeYnzYfmrCCBlwU/NDNhtdQzQ24bRUBdX7xtb916n8ZQW6KdaoeXH5ketsUoLtn0Z3EhYa+prcIbLMYgHqEV35/RiWFLwPZNpefdw1OEoNhjPgYEQt/1OcY4D9XvsEMP7Qhu+YL/ylXD98ZiF4QL/lE7plStm7S3ZUrvlTPxSbQIshhqHekVCCKaSkE4aMXhsK7Fe/C22LN1PVp/bwmGpZ8d6YR2UGjaXUYFqKxk/o/nxDY8tfMQguxpayJVRorv/+u5v8A1YZHtHCTkXvHPx/AW/EM8BDtrWu3qbJ+lbqLFEOvAtq26xNo83nPUvr/nBXzR69AGLNmfX/qJvT2eCKg9gwvjWvtQxwH7vd1fwcW20dWQf3Fp1d4amdWacBjIX/94217ysrxgO8lSIF/3H17/R5sZmmdxUQf3yafnEay6P9j/DYxS1cTRzuA68DN7bWiCrRdWAOZGfP8TnUfsu7ufQ76od6r6N71hD0uUwU1FNoeHWw7Ynxz8Jvx2Hbk4Z+1Hva/aGEYqIaPmJMrCeAjvzv/66PfQzVeRHWHafniD3Dl4bFZWKcMb2i/6LI/Of3vgW3hg1ln/3Lj95ZVXbH6anVJ8N9gURfNztP2A/Y8n/g11h73Pf3XtFXtl4E1fNDlYsWR+mZHJsFfEB+ZOPLjgu1Y01+vm239a9/P9yLH25N0fqEP1il1VEYs6D6HSYs+RESBMkt5AVrr7MX3bMtLLaUkZBkbDM4409R1HHVheJZ32ga3v/LrP9OOTy12/r4eI4ON01od2e3YEP3gClaFeuiokKsA3CxnBAG8jjDwQQiAT2wAo1gDsSZyEcRT54zu/p6b04rJSd3uaPElNfo3pogoXohAsdS1MvSjb4Gb4ge9uaIK18KxoIKquCQ4efLKOZ1OrIoJiXtzQAw8eRRUZwdCkIlWi1J7eXXUblkk+MjI4eef4K8ObjGvqt3TrQVu4WSPUaidTkTAdwubBLbQXjEOLKIY60z0yNQb+iZ2TTODEqNhnj1EoTpSaLUMVrD7zs7Ml+hQgb1Cvxuz9F+yEaMzDIIH4fbBSbWjGXFRV3QHIvX/+x4Fx5+tn/qRtjMql2VHI59Xx2wSrkwgA2VLjEm7QiQ2LQLy1Qh2Co4Sb/+Pi3+Ff3WcHG/e7zlZRLc5nFGvaTLjRzwJcSS25BpBNOufVQYkymVUSno7Yu8R7igUAZYstIcmjLbmBOu/dskvEdTbYYnLJt/ccwpuMIepXnnvk//0hfjJjNQm7NTuEyZ1IIFHMYttClBtz/fDi9R48niAKaZRnFI0vfl6Cdp6doXduEOr/MffUeaTSAICWeswLA5JOeQi14MdkHJhBdzyhoGxeq45BSAj0R+O/lRix3w6cQwtSiQPYrKsStVt4BmL6LeNW38yYnSm/RnrCOhsk29jQ3AQCpFIKAm8V7u+DdXodf3jtBjPsGIOJMRBQe4UorrA2CiKO21vDF7AUV9g4hrvb8OISqKXoEFYrX88o4zg7dgUdewO+S2Pw7Ixdqlp2Vy0Q5RRStuXfUk+d6gF+QwTiSOtWWhyC6Z9f/gky1G3b17gDW0CQllHUfn3ROEtFuPtJRbC/U+580SfhkrJR35jvtzdR566Gk/Q3uHtDl4hiRVs3TvGA6lOlYm9iWMfq8WERf+q5bYoVbvgmLYVWJSRE2hh1VitRHfc+3hnLfY+rvRz6XAU4SAPhVjaQXoDz31hw9aE6nLl6rBZecy6RsNvJPncxaET9qPtmVuA/iWiSF6b81FMsXE9JV0IeWhUiVSM076QeJU6a55m+xQHaCHsQiHaWcdiGEGq+7prtIGs1VtxZFtMwvD/BF+t6Li1JtdLoobbMgLwrtKdkUw3lDbBbCNEgeYgFt52sYrXVeFCSAmRgaYT7w/jD19PnAnUrzSAGIxBMCWkl2CpSqZZ6G6GkGeYgIVkCCrsDe4CSNSmUcJTMy+pMC74+0lVNIgcI3jvxA6ph7HPIPLdJJYK3ET5BLa5xkywgQ1dttN4NUmk0N8vISTXAowEhWpouhWQuYmSrk6YO9WkG67QyPEvrFUyGJlM6ctwccFWorahxVs07E8zyOght/tHxYf3P6wG7/jqftXtJUF+AQCnSTYoKx6177rv369YpvRrR0/yRJVSWuEmSfG6lSD+r57SFBkgv+Nw3vI2r4jbtCNrW5VsiJBan6IXPkFY72bmPtNtKW4dLcHvKfnHrLXjUfXaYtHTSvpxFSPzZjTcwwEzQB9XHAqQ+Ufe1AtE/1rzPXuw6RRBJAwLfPNTxvPWAmA+37Pe95Y2hcyB42h7Cg1Npw6dwIf7l7bdxobhOv+5QM6kYFb+KGRNqmrWd9dvtGzufsSPk9ReL9ObwWbwfl3CAO4YmgtyRLYvkA72E92KtPdl5nCHl7J3Jjz00UEhbirmVjvzlnift8fYjTs3PjV/1PPuncKRrwhYxxiK4hTt0G8TocPNOtv28fYin5BiU/jQOeTXYJkbS03Zu9oa1hRrs6e6jtoJx6hyapl8OvWMTpCkUSOXF2kmAyPf3vgAM9zmIzk1ds7/ENXsaDRSWJeAR7FhCMn9IdwFXsWkKZpmG6q9d9xq29mcdGP0BaM4nFrWv+SxjIXogyyfeHfz4QEi/vj5RPucD11/c7DOYGgLbROmdT13Pbwl/JYULsymicrpU+l7Cb+ff+bIeyLpX5m4ZUbpJSdcE0iexRNZOVmFVJf1eQztYZzaxOofOOKhTKca9aJtYV9R6W3mz7YKKKwKrOlNhlxL1uAe32nbcltW9plry3IOsO+qaONChxeoyUG92hXuLa05oR33OIytUo3ZUhFZ3PW7NwKBlpsHKs+XkigzS+lXhm94PC9TMbrC9vh2ELlgv7IWy8wqJFB4pQU2JUzvQQm0njaBMLxOoiJVufCdW2hbYmziOb8rCvI3osO2EPK6CfKOpepBwmUMq2lF/4hqRinkuG3rvGqRkOf5DK6MWJZIrnGJ+RAxoqwLDXTfj7qHPgBBLNy7KOPGF8AmKiM8tztf6sZfmTOOXc96WCdu6Sopi19qVu9mbtcv3fXiQtj4z0q+xaQGu3teJtQsAQMBwQDDhgR/82q8OOy0IbY0BsouC3L0TqAm/h/c1Kq/PIvtCXrZU/SLKEsnwB4FbnopamCLRvqn4z2ph4yKDlaKm5M2n9oudcH8PfddLARuc10NLuon1xEvcs3YydaVUfAK8UdXDwqReCd+uz+eSjEiKiCqNxfXryAv4iwZVSDqTe7OzX3f6rL0E+uDbuuR2Rknr0s0X7+FNgn0QBaXmgDtCs+QIBZ2oi+qWAmGEt164p6DkWkWSGgwD+YXdPKdn1I52amSkEE5uLhDzYHBfsY7imyit7g8Y17t/82+063DSGHTBbw56r+8ahR/hUxqPbrmnpQdBbh7fsGwZ6bXtCFByBpNzmKx6Gpy88WTOVo71tcKkyconCy2JC0F4XH85GGGVZZvXc9wqihvCow+jnmOPkFrqvhwOY9gEfXKVlYAWaEv6cqm+dE/glgtzDUTIEc/394kwkkeitAZKN3d+9iKC1oSdGbrik3SNCJ8pIoFEfXT8pUoAvHV9hg99FxZjbimJSwJCIXzu9YWbdonUeufHb4BBYbtCMIrCDG/OjGJpbGAHIa0f19LIHFJDaoqUSkPvsl56UAZjHkAF+Pc337GLE33yskW3f4tkqBwBBAvQSerrQQxG12BLWtmpRjFoyQ/+KkL1PMlYBRf1VXrnSXYDsVO900PAMUPk0gD+Qujw8alvI6hlgkCZW2RBqMW1+coE0WPA/zKGpxS89zLjr8FGoaip6zMYy5B3lKYvxTxdTwzYFH1RjnsdpKdsFpPo8v/6yqv2QTVB59zTR2Y3ZZTwBFsiUlp9xaIdV7RHRXMuNbCuiUh4/n4Wg1KtMATIBXw3XpBZyQLsaCGc1xTsv0q7Be5R4SloCwQopQXhWOa4UcaOJ7vDGuIzfXd64Y9u6c+WkV6IKq2otHdNINe3Tp7GUogGA8vlr2+85+c1lVpcBQl2N+y0Z3pOeN7DJYSe1268S5ZegrM7djHwkJ0fvu7S/xP7HoUVIQoJh6cP+z+23U0dtr+tx/3h3+b7BHruJ3dixMFzb4q23rl11kbhUUXlRcm0EMYJ5hgfnIRKswiYDKVy3t1BQHUTghnAbAzrbCYAj8ldYN+4rGJdLrO2lgbYG7JpLYdsMFPl+R+3NXewxiI2vDrlWpXGlkYPRMkupawxrYinR+0QunlN9PmRG56e++S2fdZT1+5pOt5Ds3B24kNSmovaI8yr7yDCSHKIZ/BqF6IxkpHMmH24cJ3vUtSyyNE6aYcsFZ1WXl1fYW1tTSBRxiZD8PBDQ/arxTeol3GxopToFV8uOzd6lQnjWS1y4HFzYIjvQla1FMb9GsvsArkohWfwFN1kJ362+yTZx+ptYHbULvB8HcHxrW2NLAa8aSfR2S/yPLuqqlyPbN5D7QbU21O33U7vPmX1LKoJMkq813eJulvs8e5D7pV5afKmW3wf337MOmDBRucn7d2BC3a4fq8d7tnJHGXtbF8vxrgJe/n4MwjN5CkCN97EuW4Yq7S8UJ1FpjVxDXf3hK9bKFtGeqG8ijQC+znE7Gsdj7mLwBRI2YeabGqIuFCENqFhjo4fAnFf7HjIHZe0MgdGB+2lzsfsIBY6uQ4oxO7DwUv2/DZSblQ023ANz5Mv8Rj50I+07sEVNoafdx7KPWnP1Z4AQM02WzFnk9MkbsIDUj7iQiAVUU/xoZrsHJqOagD+UtPj9lDTQVcntoTqbXB83EbyY6VH/LnSH6ccqxEAv99ebnkCDQqLuWaWMLuM7Sc19aPYJDSuMjRAylnzRPNxqHKjDVdOwuNXIwe0YbQ6iDoRhCCoY5QETE+1HfXQwonqGZvCv34Qb0upSOUuoF1LQSBh/G48QBynLyUyjeKzExN1EWPL2NxfRVQU5PUIMATsZ2uP25MNx32B1axW2u0p4maTMsRhm2CxFKCiBeBbJoQQiHgPUIMRsCD8vCiBTVQWBJZ7cxRefZeQteVR5otc+FE0QOmYHcEr9Hj9TkaetzP5crs0cg33ZtwJtCDVz1JRU7QjA9ZhDGEvdTyK0FxlQ7FJW1pJY8Hebs81nXD35Vpcmpujzfb11lPAucGGKiZtipyj39j9mHulKutDldXYuZGL9s3ux5BlkF1qSdTFIu0nR2lc7tkam5fSe6kjW3vfMtIHrrTKMkX69VWoFaDQ5Gi7kv/2HRAEnKY7P2E0khFDV8ggzTyKpeFZ3hWxo41LEyoKVsa70u85eyAKxW/ehn4XA8s9kSxbIZRKT2rxMfPFoXPFKR1A4H+EnUYBy1ibfDeIOj/Ndioejac3KgU886S7lpdsiHNJC7BtWSYxD8ulfmuEBT5rN1GgjJ//5IRbfKx4YNUrbp/+OTwCmcKVaCBWPop+3XkA+gxVdC5LWbtguTAFOXsgmxdEn6J79EGTqj47pHyh5Nn2ZYyTkctz7attEN1T+1GvjGwSNPW7EFw8siOokJ/6AzYR8k5/NadCYKUV11GZLsRyv3qkEEDZQMLAEKjgr4RWRpukxnwH9Fwo9lVj4l9cdXnb9I+KwsBc8+hxErQhhwaQwApxDG2MI8K0RFl84lZ9rfNeBoOnE070XXMfYXwOTYHiCyhbRnpta0K6PMg0ScCF1GMrUPZp/NcTnPYnH+aCeB9ZCxnB/CJGIiKGsjg0Sac7j35YcaF1ZPqSBXGI1HrTmUVSR8+5bngEf5gxUu81raCHTuLSipwwkcJPBH31FG0otnQGfWwiizlfVB0K4IuBSVMwh+iZPErzQFpJjqaIBJqqIv0dv48oagmhUMi5UfHrTFCCI3gmiLISnk0R5zoHxR5dKsNQ0+RalEmMP3LDmKCvQoxRdjmdOhJnUbQttbD7FODJoez0e4bf5HcuB7h5dgf3hoTvF2XVs4oFEMVSiKFjEd+FsOspqFg4uCAwNTC2p9NZDFBzrlqUMDqJ+lCpvEMEVkgQF5K7263joZBdDQkywIiPAfLzzj9d9YtcX8VsP59dgKUgBSALfwq+f4J5aVhMWFvZLMiedVYlzzzSAlVKeKYWNEtel4gJbeqcrUkyow3qOE50/OPpceZvFg0XqcaB2RJOdsPIJVJ16ySTQkUGPILS5+ec1a1nd0giCw0hXyjbwhA7uvo8CYxnFOml3Q+cE0HV7s4F2t14TvXrZuWBjFMCYo5VF0njM47RJIrwl8OAslxYInaz2v0/mDdHcFnwuiqUSKkC99glG8gNYqGrIYC5HViD0AxkAT23gi50DEwGy6d8YSpYEGXyNweKywR4S4hS4HYYQ0+ehENKZ5EtY6IBhpvaAb/SN9dgNFG6Oc/3uDoJUIg5pU+iaDpykqXHR2FBUBzh1gFM23ME5JGfvRywFCqiM2VrMNi0oZ4UpZ/gHKZFkjiV4Y0o407g+kved/z5m2PNvpjHEQ7nMORURjkrin6vMi5FMylmt4JxiQLLqKM0HjKGKTpJMpBOEJEhy3l4JleTGUcdWc3CibGoUvi1zAPHOhJYtcMeCO1EDJboZw0+8nHoo6K9lqhHEV8u7wivmbQYQTJ1nNOkgw3kqrzMGVf6rARPWlXLq8vAKOlwlGJgFeSU0acRT8xm4gqk+pxCv59E+G6IAQvFJpAtQcmhgBJ91Q7DMmKHiRcqCERp8yN6ViBqAyC+sqV1kg4wCcLO4Ti3hGdoe7yVdIhVBMCkIXajrDyC44F9DtK/QH+iVWHbGekGF8r8bK0RvD1j9K0x0gD+YJgj2kvna91hdUoz++nvW6f01CWdqYI7dCz5EmdChbCSahsqB6gvwH/9wf7TTErMfnjp1zYFMF7a8TjHwDTaAuFnf/7h39tD24/bY9s5/xSE/WDokv31tV9YAu0D8KPA08IWKGQw5RoepylOnRck6Ys74Z/YHObbX3IoknD6OGczfXfv82Qs67DJxIz9n1d/ipB2kwWyAJKJnWKLpN96jqY3LNo+5a6wFOLAM6ivqGwjFOo7O07bsz0naTZkv7z1Dg5Wb3nOGvG22mGalD57x/N4cJ4gWilk/9j/Pl6Qb6J5gTKhH5dPSg+6/+/set4ew8lMp2j/khSCo3gpfm3HKesGPr1kPfi7G6+RLQxtk/Z5iH8M2eQxznb99q7n/NDna5O37c3+c6Tt6LQndygd36q9PXgeqjhuX9/1uO0E0ZTG4yc337B3cZMQXy8vRLgxO1G/235w8BuOwDdJCfgGOT67sPx+a88TtkiswesIiT+59bqfMKjDkWUTaIs12nf3vUD2hSPEY2TtDJqtj0mH+F8e+RaBJDXe5x/f+C3C+SUIsHYksTERO9Kxx/744NewGjfYaGLUfn7tHdvX1m0v7j0Fwi+Rfe5jd1P+zsHT1s2hHf1orf7vj/7ObmVHjPMLfUcUsThS3m3/7SM/8MXZB1fwysCHthOL9rf3PemE7W+v/cZeHTvDQgcxhBcPULaM9MIzFQmLQg6pKp23Y6CxykrbhqZCpvc41L+ruQuEKbDiW1w1JqNRfW2dHWnp4iQ84jBBxD3kXCmTsWNZiYjohrYR57lxoEIQ1mICVb1NCWM6t1TbtyhysKWqR1AXqFA7xqJ2BN0qFl9PS6fVDBJgMC8M53n6KTcJzzCm3WETrNetrpkQd0a7qjeK81JnfTOUXvnUiRYi736UrAusS6fIOahAIzuMzpdSpNEqASydWEaV0FU+IeLvZd2sr6gIDFpYXBksGq0aq8LDcWdDCycQlnMu7zZrwefFdx9al/AWJzKqHV+iNlwrqlDd1lZWWXtbGzEDzdbIDpYHvrswUlWxU8ptupp+NFcVrIFFJP7Z40lpLcwuqms6YFgsIydYoR3rxDRVYRXEQuCwgpsFUV2c07WAWjMKi0oICcqAMtvLnLbjY5THfNDd2AG1n8dY1SqwkzeUvtS1WGEcGYf5lJpVrik7OGO4C+NYJYibZrxtHa3MTzOuIuyOWAh31bWieUpz5muTR3OlsPxW036EVCLKrCCcKoObkIW5HQWHYIGd0A60dlkrMbyKCVgmXraJfkRn4h5QL+WAEkmJaVMJ/vrHDf+IvdxScXYChIEYs3VSLUipoGw1EcqgDuQs1tJW474tOFuJZGkNitXXMZPzYjNAApV5/EiCTF18AcEgrwHi84wf0whvKV5UKC6BS2yVsgLoszQdvkbUGQCeIceiUswJXVM4p4m3eMAt/AAAQABJREFUFuesnUm7gbZeRSitR/j7+Hs9zgPeO41PdaPRyeHmICrmhc+ujtewqVs69BS/pRmgi3HsQnI4E0ujdqWnV9ur7EYrsGlqQkWpsrWtKwpJJatoL2AoKu8QJQ5AeuwMbI8f7kxbYjHkL+QnOOo+blzBBrIAHDOCH/fI70VsSdASBAEgiZtX9JMs12KZlPh2AT/9FTeQqHUEVeqSTl+wFmyFQLovKb058NX4pVVxJlHzRBFcMsgYnhvICQntgLFJjV2joA7tluklotAYa/CdaDDv8wqIT58pkkVkm1GfncjRnhsKffenHgidDt9L4lekoP/gGWBK29ISqq0gn47/tKU/W6b0G9UGnCmcjwRf/+H4FefNxbv+Dl+UZfjCPB3TNreC6fr20rC9cj1vQ22Y1wHIpcn+gEXiszBYcHPkAlkgzmCCgA0A1AgTUHAVH5+5X9oewMV1WmchXOQcqHj0HadY45jm+0l/p2BtUVnVLRogaqBn/Tn+3l8kEAuEuhtko/plhLt3R855qhBt4RemrpJ9QDuTOiuaHUKQn4NdOI+tAAGbf+cJoE7grKYJz9E3PrjX568JBh8k85pYgPOTVzkqZwGkzeIn04TdYYpjJwd8nO4ZSuWC6RXOi1JyWR1zOYg++yOOJpLRKMEu4me/AkOlHycTCrsRwiLC543ZQR+i4CbCIAWD0u/9ov9dj7mVQHlh6iYGuEoWLNnNQL4rRFLNIny7fVcgYmxKofIabI+MZVqkvQSnDMBK/aj3tzihVdnEAoEo+DRpYTm5Ab5pFtIlguB/Dq9eCQs0Sdr0j9DLz6MgYD0ik6zYddpWRFrkdsx3RQW8jKRJMuVygXzgtcDxZl24bT+6+Vtknhg592cJrieYHTlA3qpJCNtZjnFdgQiIOGUFZxaTjF7SrGnyNAyGv2H5XEivGkUxla7iwsxF8sR/zAWQAXZFarNr01eBHx1g9YrtOUBmYSVG0j/Fv4YdIbmDOhQ3Ws42fnrnU0T87/RwwbfR1co49RzBFFo8ieVlAi3ex1V2iFqhFCwEZsSiGJXi5QiK8MF1IEklAtq3cco6gMOZVHQXE7fsPTIiKGvAemPPeoiIPz/UtBfe/GE//GwKxH2z/4y9Nf6hvTF6BghCywFuBIcxBVtrTxB1UphmOUJ2JUKnxlqDHvkgTlonCbgQ2yLL8BUOeqvmeiUsi8YcR1VXTQqMWtixStiPGhI0NdbW2sF2pQ4/7PaO8xCOd3F4+/jSdYex2EkJzDs5YrMa1ekqC7+sPIL+f9iuXr4WIA0ESItbDmq+fH2Rm5+vdTNxO7gGm1iHLr4bJ7Vy+qNY3Sq0JoebDtgpzqTVKeuj2EZeHXgP1+d37TUOX/PjUYG1qrt5qY96pBQIY7irsxe7n+K54yBg0i5wXuwE2poo9VbB6tSAB2T6IRvFB/ZW/wfgJ3w/GNeD7UPH8EjALsfzVUHuqtxjhJkHsYXyO6oCZkDMknyuZiHFUJ5ImEXv6QL+oWYywuFoWIEgv4DmSUcqzaLdUiyuc8rC/A3KAyD9nXXjn6B8cCAiZACAbgoZaMzDCYXM/pIbgQqpMTDPfwUh9nDzfoASBYnryZX4sVM01ZfFtaEa/vGR5gN2rOWgB3TMQtHqoTQPtx32NBgJtAW98zfd5df3ZNoVssrg9WLXo+50NVfggDUWyuNtx+xY2x5nL8oA7sWxG7aQR7gUJVCP6JizO0y6b9FQh27qeRzPR2VUmIa/vQnV7McKKNM3wwmeYUSi4iqiojo9/FTHYQ5VxvMRFkvBHlJjPsJp1rvx0BxdaPZYzjb46qMY3jT2NI/r4Oavd2PEwTVAJ6RIh98Kwj2LsU4aK8WufsSOsIKgL2qnTauR3x8hcuvZzpPOfsTQ2Vye6oWlW3aNklR5QSyDBudd9H4LiSIQHRXZGeqRHwSfp0h/sgzLWaNDlUGaU/IexRrdz9xcSfTiHoExT7yPWA4RGeqMxPhT4Jxg6tTJ3YexvD+F8U4hiVGIjnLd6MDnGuSaIYTmiywEnY4SKYdMMAgtyO0Qg69zUHOdvD6h9B+zg05m8ICVKpqaZatR1NnXUIRo9x1jIaVJwb4t3AwOPQSbh+szmpsF5vopiFs1GSKUuOvt4Ut+wmLMOYRgngMo3P33AZD+zoNCAKktnQ2hbmkCI3ROaj0tsQL6W+bIeX93VvLJQDsDUMQ8qOikPKkcSwgkn4+KHL7jnMEqYUYFjx5WNlRVB2yxwKQ5khFLumjpqb3wrhQZMpDpnkoMXFG2uRCVuBqNa+JGNPv6VypqQzJIYKxR3eSEp4IyTTLPSMMk6yhMFmwKF6hf7atoZ9K4dDmOmjOGelaURbxlDEqrCKzAOc3xxMcBB6IG9dfVqexNrvtXlzSmSlpSm95FrgmWjIi/ILI/BTLwQURFu6J+l5Yc2sMd+scHTUzxboczX0tV6rLS7kXoazmWwjg7jYrGr3kMA0PZEqjIa5CBqZRR2WEnuDNGp6DQYvH+2u8ctnqG38R1Kk9nKWpL7wpXlGFOrheajzDtxuHHBDM1pH5Ka+dzDEw1jXrJ7SPooBYy88puEGfxeVE/aF9Jo9w2xO+qSRyGF75r3JuVz4T0qnAFiiHhSJY85ZRZRagtIOyoKH22LHEkvXCE0SG+ypx1ZXrAo/A1Cjk5KWNXGlWIDtQFgu5LciU5YNHFKtfN98NDjmGoGGsmUh8T/TiGngltXwzaZ1tvDF6psHsJwthWt2oJAjJGoLS3SJldQ5oI3Xybc0hX5EzFxJWKJloaB09bIkRn+x7GaHIDJ65WVKxj8LhyJ9AxOUpnLaDK4UnAVioP5VcR6o1jN7jCGbIVCxwIwZVeUmVP43PSjWajcrXM+tHtX8HHpgVVqLQlsksoVfg4gn4DfGsbGYDH5slsDFVsw5PxBvy3+tSXHHO5KCtBWnCFcCRgIW4Ck6Zl9PRQxVsEg0jn7ggrbA0AUxqirgQvruv8qRy7qVSzM/lpu566ZZ2pJrK/pd2XSUajIVyUNdQhxq20f8EaErIKEyWjQO8lPPqiQ21dliLx6xTwHYfyLtlVYDdFwtntsJ91GZzngOcUWekKjEFpBt1CTIbj0fSQnZsjfkDuJ8zVNHp5SJTjgNrS+biy49xmDjmAyPrww9HJMnPYDPakt7kAP0B68QV2uBtNGDxXa7GhkHyXGAaFckp8+6SydeOUIFgsokHbKjvgLeNY2cifyPasVG8dRNTHoBgjBCOsIMi2oL6qDlW5cWqcUDPpzCvhH2Xyl+ZByrEmBlIbrXFtzjTI0Ig/RhsGLDzJrX9lEC/CJVY0R8mzyLQ9rqDukhem8FfsidSRKE3xl8cVid0jx46RZttT6j1RUE2UNBoKiXMJjb9CfmkI6oh+asLLEYkDCyQsBoLYNlRxtajGZGWWgCXq20qKai0yWY+FbA2wJJWwTEuwBtpW62HL2jFgSVeiiZ7j2RgGsyjPKGBaEy61Xoz+iWCsknUqhhGnA4NNbXk5YYUsZiyZtfgjdcY6vJ4xFssCkUqyA1QgdM6TCnCWNB061aSzqg0akCOSa5znmGj9066pradEDYtzpd1MiWYF58pwFfOVxLJNgAdIVs53IZkSBmbD5LOhz5gGXaZIQSQ8sBrMD4lC0G/NbSMGtXp46Dk0QBLGlQCqq2I7iJq2IZA5wwLt5pQTzfsc/e8jY3IFyo2WalSfLJZZ/IRSzEUXvvptYQ5zJk1LX3rU1ZjqeqBh01jAFWZRh2/L72ilMM9n5Db8/MPsEknmWG4Oeyp30RaEIzeNJXjQd2mlnwwIQBEI97w9MKXPAqSTDQftv3/k+0jxNQB+yn6MsNOKx+XLPY9hZayyn1z/LVbavJ3mTCOljc5Cwf6XD/7SLs5fxwCi/CfaOg3hab/96ZHv4HDW4hTjF7jfHuespBM4nFWg1/2HwbftH4be9sn2BElAXgirrd2LEJ9pUqjfMovIZ8bvYScC2ZJMuG/8esaRIXhOqr32ilb73s7TbnypQPj9ee+7bqh5uv0oQRiNHu3/DwPv237y0SvFN43aK7ffI2Qubc9sP4HmoQVKNG5vDpy3XZzz9Oi2A1DkHIL2OfvF4JveZ/fIKPZH2K68kiryeX+c1Hy/v+8rICOHw0Hp3yCaqp30Ii92Pe5hgu9jqNLBCE9jHOpEp32NXfKdsY/R92/zVCaSART3+qObZC1YnUdmAlHUScbsO5oIAm1JznoUGem/wGBUiyPdEATp7269hmHrHCRH6kUHIsimxLbShqiHAZwFLbGfWszymTlJ/vx/R3rAChQOQ+S9eZMDLaohHN/pfsap7rvTl3Aem7Kv7XySmF0JxGP288H3UEz02Mtdj7j89gHem5eQlb679xnmHac9dor/8NGPOOn8ZsC6MU+QCtoldQl4Q9PFIpkANa52bP4psOaxtpP2g91fR8CHcKBt+vdn/5JQ0X5sDRAX8WKCxwZl60gv2FCPLHY1aF4qkb7LkaZrV2ugRnUetSSEV+oK5XKppL06KGIcKsX6RErHYISGwxlhRpIneWe8sgIhpBpjDhFIhToEUfK4o3qTcFcB5a4mz4oyBYCzDDMYgG+59wxEk6x/dxVd04ObFEUKVVUxDoQgpdeWAa2cftYTNaXUftUAvYbP9SxaxXsKFyplXEKVWAu1q2Kh1HO9BaFPiKvTQmQA02EHigNQkdaq1APvI/3RIhQ/XIkmpg4jTg3Utb6GPPI1GGmotxq4asyCqahwLSb8SihrJcap6lpS65GysA44S1XYRJsyOM2nJEeBoLxK7XkH+BOGnaoAASsgLEpNXgdxqEQAFbJLA7P+fu8jfVMp2Vz8G4OXYqKc5xtgSWSeqSvPE9/bSkLZKGOvlLsVxjHqZfdT5FkN12rxQK0HPs20V4VbBvudz3cLlm7hiN+Tg55XIvHgyg0DEIzDe0Cbxb4Uv/qbrgWEjLqoW6kFq5hDEWPNC6flMWaSw6rjvoDXPx183jLSO0XQCnRenfRy8IKrIGYaKVoOXkk6nILfw4vGDR8rrDQ5Z1URg+kSENueBBkdYeOd0QLCoKU0czr1QvlqZFKW74yn1kN4jaxkeDRY6hpoUO5Db7/8ab8XH157CyMj5FcIUMEwFEbQWoGqpNniM/DzWVx28/SDjvEZ44o+q9s4XBmUflXXMCBlUjn8P3QbBindK3dn3KPDUPKNigvOjEMCcRRYAUQ/DFk+NwU+5wl1lBFI7YmiBXDgM4iaS9N2CikEYqF7xOOn0RbJ2OdjFzXQqwQmEX367GkJ0QSp73nYFOnSxS44FHU/n5gWR/7g0eBviWAEwj51qT4IgfIM5ahY/vwRCABMj/cng19UhhziKxj0BI+KMHjBWb2hFQxtMVhM2pdcIaNShnFmwI0Uxial7wgzB+pnDn7Fu7QGvKAvwVdhsYYX/NOnCDBZBb8yEJs0i01mOYS8YMHq9vWP+9PBny3z9E4ZAJb8wOtI9vnQ7mNQwHpbITf7Wc4uUqLR/a07QKCQXZ2Q34vZkcY9VgNlWllJ4aPxMTzqnHtqBpbCHJSN49zxeW+DNZpYTtg1Uke0Yohpa2yB0hKjOtdL4qLxYk81OSwatCwlKqQfNC5NoDw3FVwgA1fJ6zJ48P6/el5RVDpKpwN2QXGovdMcAc9E7q7pcFP3PPzqbQQzxbfurucYH6j0DfTtsh3uaNhBpBLJlciOcGO2H2rfat1NXSB7cM9t+FhpcoJpKvbPdfyBhMVUu/vCgaY91lhDprS5GbtFpFQDlGsvfjLq/23Smiyhou3G10bJUnXEZ+/STRy3Gm0fPu6gFGdmjRMpNYwAiKpUqbohSg4fTb6wlCKvzkae2d8I7wtvnUCw7CVySunC5d+jkkXmESopi4NgKQWFgrs1AB8DQNYiaECeOIitpbquhvSCpD0klUgFVHYvdWfTCPEL/QSar9hhskVXc8rFIkal88BsGzaWo9hAEgjSIzMjqDDn7GATWeLYOeW5KpVlGjkosEgHUFN70vBJXpEspz7KPlLKbar8nNUoBk41n7Awdprl+UX7cPS6JSO4Urg3MP1HztmobJnSiy0peTVG8IWuxtuvPs92BuVRXnjpeQf6R2kDpGPyu6u3u8diDcgbibM1wnsnuY7WTBDkNoWJhfHAIyMvQmENxgfsVjgj9dqVueuOzOLDmxEiH8eZS6koZNk7i25/Ag2GFg56IgCSJQi6m8ATUtuxrc5xwPCHOnjNLYwbDTmYQDljXccKepVzm1ScslHlJEikRaHvHqJXmILf7HP5kN5jcOr0jMRRdMxx2AwJmcooXAObIS9DwULOaatQYg1Vwps8QGX42olDnLp9CR34PIayKoSySrbiOp5XpuAKniWvsTTiwKKcRdVvA8NDa/KIpwOBJZPAGUVXXh4n3DDUYgeRg5qAk9KGXEG3P7Q44lotIbJUnAlg8e7INN/YFfhejcB8HP32QXLTKK+PxjdHRrdjDQdYYAiFhFx+iMVzSoevCXFwAxACzSB0vjmJkWkigI/mJw9rdYvn9bvUxh2MsRqbQE2hChhknP1SFugaCGUITd0845vnuSrGSpIUW8HwFAOZ09qWxMLyRq0E6bTYU9sfsyoduoHd5eziNQ7Ja7ST2/a6DeP9mY/gLlIYwWCBIWD5GP73MSgt/7XQqWLTsmWkF78lX2ulkT7Svtv+5MCLzrMvI0CE2W5/M/o7BDVRN7Y4tpqHevbbH+34CvwnfBZjmcRsPTe3wPPUoZ7Bm+2v5wyjfS+grcETEyFv8SNyLaLJ0Kp2pKO2A5zx9J2dT3uE/zyCXQqNwBSqKV/51JRn6z6F9fPbu553SiQBb5HFMTs2D19HO59QNGmBgAthLO7xMrRtVERt0C7bE63H7duk2quHJx1NT3nu+91VnZ46T9Fcrw3VWuL2tI1CBMRfyITfgZPVN/c9accbMJbR5q9uV3sU0Us7n4AnrsT/fsmTFTXCu5/e8TALuWC/QbAdRW05BVIq6EbaJrk4f63rWc6zOuosyhmEwhEiik7vIEKNVOBTxDj8CLZoYHkI+MrvB8CzyES9lbZFEyFVfCvGqa+jdHii/ZgLzToFRqePP9f5kDXorFqSPcmQNzM2HYDCETKox7OpCT+LRS7K2lVUr1S6D2M8+5M9LyGHVPi5VNWDFZ74SQoCJbD6iCRYlxP99nv7nvd8+WNQ+mlcRy6QckW2D/U1xqI5Qvjgf4XwLYezKXCifgoX70KDPYeH6TLGKR0flMKl/VvdzyMfBcYpnSlwcewS7gjAi/445Sx1dN27ftpS8Wh6UUAZOBQGVQSENL8p+HRgS3d5iULyHXSkYRl23JbnQNEjEoalXhNPWrSFaC6gktTg0gdftBOo8CZcpCr9Z6mwpMhRGKS51g38zu1pgIW04fXodOuM5kH6LO+R37blP0L+9S9/kKqlSpMjVoZFJTcnVS1uWodCwPV533R9BcKQZty+iAQv9Rx2xVku3524lx1Oghdd9Xqykg8iymIcLFJRaC0yjV1IIJjykZrkmSPOle0ePjYrWwn98qfoj2DgfUNRrXgGIb2eVdHzeumvLsmDw7/xLrCn4bt1dq6KxioK731X/9cVn18qCOoN5AGN0oN6uFepHzW96pPGhxOnpcuQQdQNxp9nrLKQFJDfdE07f9Z3RbXJJV4S9DPuX6Vn+K45gei6tZ9LihFYiZFsCsou/AoKC1v5K0Emx7NPUNZvmdJLzegBzFCqMVblLbbQRowCCc4dSuAEpXRzUslpXpljG4NPVSR+ExRoBco9D39XkPejAAarE4KCJTjGUpH54jOnMdbMsqIFKcHHC3XNEMUzgPFHARsLbN8KRC/lfdd9GuAE2QZ68TefRdOywLlIqlf6Yh2zow6Jmsv8z9taEVLKCCTHLSGZrHmKynco+11CNNwBJETqHtqJYUmcwpemj6xmS5wRNYSZfmCZNHrwwHFYFBluhvAhX0kn2YIR2hiLfNMXkA/6SYGnZEvyO5oiokypwC/Fm+DjlQEhYcPTJF2tztgAdWvHlHleSgLlgc86T0hcNm4ZtwjabiITg7bwPrKZyd13kDbTlSl2himoZgKvRQRc+Geg4zyw8rhrB1YRXstlQP3pQKOjoP1+jEtKLdhXPubUeJDA+2kc6LSo5aukf0ItLVyNS7CT/OQ+VrwLOKAHC5DIMXbh2xzrKU3UKH7wY3NTFoO6XcsN4eC2gmvHJKc4znP0J7acSs6gwq6RxGlPlEO6ea3IAoqRecZ/ixSKOmxCJyjeYqwLkaQ1zzRC6ZWFeQw8MOurHEeDo90y4Spe7wh1+FrlbaOyZUHWR12sQQHAbRhxZJyS1mWKtM7lcaJaMEAI6SdBuigr7RC8Zi0C3yLb0ccEGyQxNAVmNyED0USoJpUPU7kaZ0GCG3N9TAIuyUxMqcja2AIyy6dcrrhz5Kosx1ekphwekU7Nwl5FoWp7SW3XzNY8npjn3NY+rMRlLieIdMzDPiywqEAFYCoUZ9LonzL0VgEwaZXmCEVUajtNrJCFO5wvlsGoDnWhcGYWr0EFhSiTWiVnSqmvCXTknRhadtd084zyYeJrohA5/MErcUSTt+T8Cn4pbP216jMInGCCxG3srd3laUtGEQpvMvYIvLNUphr+AuPSYqtFBao0eor+WmDsihvYq8xtEJBbWG9lCJTKVHLRIpqn5dQy7IFSLtJnxiVvSUV7ier65PCmUo0s0kJyK+XEmVU4JZblBuBcCf+ts6SUoU1Eo0TRATYqzzLkpmYWOE5gaGFkLMvgGxRS8K0KABC708LYdQrLMqzWPDJWO37/B1FYrNCPQZB3mOiyRvnRQwS0AJWCPBMmiRR9kfuG4o2x4fu8x+mPssolMIB2VG3Dd2sPIjxsEnlJ55aXrJVMGnHmUsbQCc719R026M2GfzWeLVP6Ug16SBWPILnLrMzMYKGst5c5FO1rHM6gY2d+2vuWLdK1p1qOMCGNbJ+r9h8+JEcMJ3nLSUuTIal8T3Wn/f7+57BmNkFZ5uw/kh76MucaydgUAFvV54inJX2g73toInDF/cO9L9pjeCNK+/DT6+/QUgZh95BbThPdy/bjK7/GWHQQPvsACyhkMvT89fVX0Bpw2C8AVf8VpPB7u08TdXXIvSRfGfzAfn7zdRbRtLctS28d1to/2v9NNwaJaPySXPY/vf0aTmhQINwvJJ904Dz11e2nOOPqgFP890ZJr4fg+EwX58jWdXuc8N/2vuqemtPJUSgtIIf1eg6nrO/t/gr5LUmVgeD49zdI5Td+Ab484carKjQlp+G7v7n7KV8YVxG63x69RBRWOw5uB30M7xF99qObr6JFGXCE0VLdiyHoD0ndfaKFM3TZpd7EWPZXvf8Ij47sVZpE3hXqN8tCZ7Dad3njtPGETjwUWSrujPymoh1PMoHO0vpvjv+hn507jDv0j2n7rdEzILrfRl2wM1DpIZLNigXTDruN6KiXGMcp4KFcpGeRQ3545ZfsbAjozs9I3mD/YxcVW+PPITcqj9DAHIoR14JBZLHufwMcO8Wh1El2m4pCpf165R3rl0u2xsBuE8F79ZNKCac+E9KrkTKQW1RdSBwWZYFyKJZVETNNER3HQ/IlkLkJbYCfw4Rg495J7EnuCMYAKzFgtKLZ0EsupLL03VdgPrWSQwBGLgBKli2zvwwRcViSap6JQAl1GngTFscaLLnyfGzCPF4vIwyEWxoJ+fAHWwhIzz+dK1uH+0MNKkm5NDdjEo+TPluT7jIH4JfQ1oBWqY6xqej8KJ3lWoYqTf904p3iXtsw/shDMgN704B6LgkPIWOMzniqw7gkg51M+M7aoaKSr38T17tYeNLIiHDU8K5JkWZGx2rGMJ5JCdBIH6u4pjjTBrw/dVKIjDvC4CaMY+UsfPVTCbeczaLdBsEBbVAezZDyz3hiLR/BnT9S/8k9QSgt1k1kSA5+KhBbTbEuetFHcfgVwLmV3VnX64GLzq/SF/2uottDwEe7mmiUxqPoOJ0b0MqhEWl2o1Z2uxj8Os2zG2Cs5GH5MDn7DLHThgQ9UU3MO+7asDxaAsp/IwtuE24qRCfD0lQRE41rB7s8jFZxDNqlNy/qj7Nmm99y9y+lbUMPerIdRugwoYOSsNe8HnlM/HeK7VhpMgKhFmFFkggUVryl82581iR5oQ4JZQFr4cNmEFocUpNCZ7SjcL+8GyXU5OWgJqhS9EwKVkmuBVQNdUOQlWrURSnACXDFkhTkqcnsCuHVb/Hzur9k9SOOyZFRsy33Ww9N5BGxRDLI+eSoDy5wa6IYC9u6opYyEs54LgrvrPTWaZDf2Qm1Q11KdisYCbkQZpxYpGFdUhoHFWd4Ri9BVOP2PtCk/GtkF1ERcUnTZwmuSmRLcyAHPDaw9lFxrxMTCEPW66VtqBKnRoH8xTEEVXl93pYYfB8bP3jfHCB85ItgpTfu9pcaZL50MqB+VnpyJaoVEIP76JPfKMjwom/OCTDurGDGPOhnCc1AgP4rxE/1qDpp9IrzzLvuExsYzDBtaDHAKq2ElvmFrzSknsiRTYtVs+G106dAF6Brdxea4RZq5v8D8fSO8FrFAKDUJa+a71qBndXooaEeQwhjWSZrN8fyVKJ+SoIU12bx+4Z/8zAfIS69q4Ri9cS2sztE2X4X0fUPYZAhwh1rqQYWIueKtl0JwBqcoqdi7BDbYsRnwluKLx9CbZiHfeqsaONc1Co/jaIfNkZnrTZADYR9k/C9E/hrC/i+6KhNft2dZa04UHFGLYtzEiOYzl0qGWU0LnmGantuJ+haI9aZSxMIskrZx4zyH6rP5G3nnmZeKjKm6XDmLsYu24PUpyPYFWZIIiuIC/jaxpUWo6eyE7M8sakYZvoR1ufgeRULrLqlQZErQmcdshPBFAkES3mP1hGHLMc+ockkAvEQ/LHYQZ7w2hWk34WxrJUdZxU2QcLtKB6JsuhqkdC0J9sS6+GOZOr0vUVIpQg24C6VqxC7jF2wHq/VHfDVcR/Xso3gGaoAeFegOFYxNuCtFH4+YxABUfoWdO7bY8TGQpimmIshEsYqZoBBONEJYcdxlZsWNVRKTmaeI8mRWYsZ9opdoSvebq0kolWo4QAGwEm8QRVqqbWrgUWIafaFR+taSuuLvmkOVR4I6UXhPRyLBxUBJOqvf3JGOgGP/cxOgiIAzm+I9v/d7CWorlwQQDMA5+d6avW6zzONI5S2o99+uedpggra8Mabsnf7z9repm0IwFgP2Urf7P8AhJmz0zsfw6rXxiRP2evDZ6yVNHE6J1ZelG+hzz4/iZAMxfM1z8SWk53reZygDnXsp3chO0sI3wejHwRuuABV/a6EZXgGJ6jj7RzYxeKRw9cZIr9SuLmi6C1SLxknWCD8LpBJVaZ3/egqTKdetMCkMEv8AjSgtEIsnYYiOhQc/VikoMWdK6CMWtA8424Z3Kmq9VItoqr6wkX/yB9lcJOuXqpjZxO5z7Oj+SNCYJ7UI9499Uf9oi4+7iad4leJcGpCTggOVvgQa3NfUVulCu4UETTJUbWcCvgMxqHjnGel1CJnkYteH+J8LjDVpQDg4mps9c3xQBiswO8d9hIBIo3IeRPM12t9Z4mU6rBndz1ii7glfDR6zW6Sb/N098O2o7yVKLgJwhLPYJnfa8dIhSht1e/ImKBU5YF6U7BgEBACZYSTXlPwUcDJ3qrd9vX9z6GMqMYONGk/7H3F5jCokTCOudKc3SmqhZ76hc/A0+s5Pewz5BMhp63jSNWPYnypAlkXUNldnxu26RzHofjcgURYWwUcReLoSVHJzpomO4V3otL6daKVSOK6exiEP0KEUQVBA1NQtzqo28MIZfp9CpP9XHaGnJjb7WT9PteOTJPU6QaCXAp3WffbBy4xtC0nqPdRvAIFrzKwsHf2mi0tBtujhKUG+MzDmM8fatjtFtV5+nwDoWiE3C4RsSAsUJ9Mhqot2EFWukYbfkVUEGwTcvEIYxJYlXiJZ4QDui4xURf8AR6ELIky+bYsDOU+VgjIXYJL0G7xB9e3C2B+7CYT6XVBgYO+qRZ1kHq1I+pNdZbqFYVm8e1A2H4aT9Ea+P3RlVYbTI2gDrxJX+5GjKCTPA4bU49F+DEMTccb9znBglTY74gXXkFDpKRUYsOcwqtzapJn5Dq9H0H6KQhJLfLUIFb0IXIb7a/bZQ+Tmz9ZlbIy1JdyJVfUVjvuFc1ElPUSb/DMzpO4OOyANYVtoeJ3pkljogYYi+ClAJjAN0cEFC0RsmN3W5c91LIHL8tamyWl+ttjF2wWlah7DsCRbFa2jPQe/gdQRW1URBFKJSA+rEStel5KTvT/MfeeQXpm153f7ZxzA41GbOQ4wGAyOJHDNMMhKYmUrEBlu3a9tS67ylVbrvInV/mbXevwwSXb2l17TWm1EsUgUhTDkDMckhOBQRjk0EAHdDfQOefg3+887ws0MD1DSLtl6wG6+w1PuOHcc0/8H3eFmIRgQTTayeXwsxgwFBCVQa0CTmAxs298i9xCESTIRe7pYsku5QXvWQzah5XF3aJNLTN/1IkP2Rs5O1SbeL7tQSTyjkyUzhp9DZrENIVaqAzxONrMHcMCIpEGiXJdPNc2xQc2gpM9guPmLExeH0SdOz9OyF3DHy/PDjk15/BBbBzuA14Xfc7OiPe8lGZzn8DZ5Kyclv8o9zfG/067kNm9V1zImfFQxkrC5PoSHug/x1WlT8X23sO75z5Tz/I+3AOUkdi17He220EuzgFnR7KGfYEiAxPTeWagnM/w2jJGik9FzFdJnOP4OB/e01HmecyHc2AmlT4Sww/CoaaCLbHLOLinKBh8wzPto3PIszhFBGuWNd+4+Qkhlu3i2U5pb+8ejo+il8cDE/3dy+99ZUeEeDDT/xZOq7qlMrgqzhuSFYpoXN6Em79Kzi8X1QvZC2pVF6u8iC3rKo6eqyQ5a3Voxt5fhcWkG2BYod0GG8YjiKiPRJQbQ9R25frNJY1pAo3eOq0mJUvwoQfAcadnZsIevBXnizvKjZHboGQhZ8I5TDAwCnIMu30vO8sAbS4ltKGdzKtxbNVyFEdRru0QqUgDQOLQx70Qdvgy35t/ZH9pt813UWY7AcowSvUgjkHL+4iq0Am+TC/jcR9JcL7jQ5+xrNhX81AvjXUTltCITrVAdlpn2PLD8wpBrjBnhpm7CGRU2FFQUBPmyluRWF5DdfNeynd23gI6nAIUw40zaRQzqbDlV4lvuoFTz/nqJPyhB7/GududgeczszKT2gdu0Ad2Pm4d6y92Q/Y+n+ms8OEyelXvMM6vzaBKCDuImGQWloX2tGkEI3KFrXH8PWR6THgR257dyNV6hzOxko1GbAYVi8fhTZskqGgCgiEmMbb8u0+2ExJVITgvhcjVjVxTzcqeKJgOR08VRF+N7FlGiGo/0NjTEOIKhRYWiwjvxSAcCo5cndkNro1t1nh4FcNydoBZiNuiv4sSOQqei8E2iFhgrH4Fyq6ezVE8mZUiFGD2E+BoHC/hBLZs1DDaBzeTsPnrVlyLKCTLU5k08hENL+4bXDs4/d3+Za9yneSNY5SN/dorJT7NhvTOeK6+W3DH1R/kXsez5Xz5g5c+yyTyesy1tZgWVcqH6eckDiuxiTIOyAUQD6w7167sBhXMRQMeYh1hU+CFjpCRVUVQYWMhCjHtG8URN45oaV8Eyy20ojLrvwYzqtlcK4glg8AwzkGMjfg3AmaQuKiBxYFUS1CZ5uRZto4JlHaz5poLakkUwjkF7QwusAAIQV7SlInc7k7IOgqxJhgZn7l/C19YF6BYxYi+1NfF040W7xpAV5QdwelhbHJ4Gat0lj8cmzy9/r05/ar7xD01pQkhYW3VV1BgtEl/v+PN9O+v/QBvIPIfYoSc0enxWonJhy9iMTlMMNk/f+jLWEjWsfKBpLvyJigCwGAAZ1eBFeI7V36evnv9jTSwQrhoKHzcIduh3M/YRYgvZwCf3fxE+o09n8WJ08gkT6Y/OfkX6STVwDVvyvpEVn4UR9lv7/1sQP+d7m/HWXQqsrSE2tOb+b0bP09/ffVVKgaSuQ+LwY5A9k9d+v2DXwSh4CjbdCGQeT9P37jyExbMyD1bZ35g83/tp/3NH/m+59/n/97zOdQUIgtf5icn/zd//sf9lThUQHcg2/7egS8BJYIHFDOf2DV/cf7viG6kfA7UlMnJELxEEZuWSBUl6THk8D/c90WsUA2EhnSkH3e/Ry2BA+kTGCgkZOtv/W9n/hIzKPkCWtN4nnDlX975GWC2n410zzf7zkedgV/b8zS+l1oC1/rTt678DD1tNw6qo3iVZ6it9UE6Qx3fX3vok1h0gCIk3OBfnvpaurZI5hRNKjB8AmZjvgNkz/Sx78C4mrBG/c6Bz5GNdyzSNL959Y30NzdeSzPFSBRQfTGilGmoxegMgeB8h1A+PGouirUPZ2TVjzK6zXBigsvQEDdCtxtDO4Whkyt51KIwFWt2UtG7s0cxUF7PL8/zzKqQwhh5CFlHSThyiqrpKp/BcEtxrBQgwmTIBrJU754dPleFSlneyn3G1vgAw1oFRY19Pp7N+ci1OnhKcNYYWquduxpYPWFJFPNsWTV5sWZ5OdByRC0T6gzC7kX9Vj5XUdM97nPCvk977M/9R9bT3G/bzEkSSfwLMUyu61XxZTYmcCwtRY5FyLW0Pf7ef/P73ucXBq2NURWJwTZ7lBFjX40b36jDqH0L+/Se9tH/+cO+xPjxgd3RDFvLPGjhMklb/cedoITxyA7aDT3oFCxlZ/YiwwfMpGtgx68QaYH36kkV1VShgcHBQ6AbzoEpCiZbDt69hx7ycEoGYUC4XOXchZrmjuTBziqQU11hJklwo4BXNKYoxFE6Y7dUYPnPUx3E7NK1fvOEBzvyxL7W2cIwWKxgmlKMBkJpvdHGu8CAFfC5iAUOikqwdZ108DhJxkuYsaSndR5xZBqxwwToSUSaJbOy2AqXSaJeIctGi4vyeQGEK2WshK2clzxjEjOXCSA1DM44TjEDv3hgjmM6UYgvgKmOg05WD/ebxl8wTYbTbLk7hQ4gKpETDKULPUKfUQBd0ooHE1hzjA+RSmyv2UMfM55BzBJ0AQrhMvfTaRZqGwsxEiBojkFohkT7LFP2xJsMxVyK48j9yd484G8XlA3TWTaK2dV4J2N3xkFp0wMsTKBOIB1qerclHMfRww1xnjaZwO5czHD+JPeYIWYpsrgQHaYUM9HRlhmbfOaVwXiGCws5aFmiGXJsJ/FHTHCtTNJgtvlp5rhsMhCpp7D8jBJEN0nfZ/AbzBN96ZgK587w0hDbB8PxNYfjFyZgjB6WVR3hvsJ9i543R5yU3zmOsUNCYyrFD3I8MNF7szxXiRvnZl539jQE8+7gGSZ8Ho5RAuTaRZIkxnATo13rYGLLYQ3EQAO6DZHppSNKEWXmuyABWxnvFjLaGeDl+pEBu8lIUgY/3nshJrCI1wL4qE6adS9BFxuywD2tf3WWGk5ynnVELA6hTwgdItOUeQTXZjDbieF+FdSuVpAEblCI+SoZPUOYOy0SLDbL2ZErICAMEl7hcsTGDj1Yt/WNrjNgws9ClEXA4V0meVxYvxiBNX/5TB8uYRSzQNcRntCCj8EwhXq4pc+aJgZlCHm0FzFJZ43wdKZR5vHevfE9Y73mk+5+mD9X5fQ2yGw/AmbwOmLDEgR5HsjDaf4Ww5XVw8Iz7K6SO9xz9Vgb0fjtjp8R+kASCQBbFxFxeoiHugKiskjS7URyCulSUkA6kAUaIPIRHFNvktQzjtFiBhq4RKCfBgL9azoH+3l9CrCnPkTGCQh4Gsej5wg/OHcDuZ6gOGE7ejFGBKUHtWdtcxzdmQtgCIYiDMEQX+09ifxvTsViOjF0NaC/S5Dzxc2BDUIfJo3H6sl3b82/D0b0DNZanD7bohVx+EcG+hI5p67wBRSPzdXb0iOkCzbgVZyCUAx8sgbTPlL0vNdF5Lr3iZT728s/5j0rVmWYTKKi8hYCsM23zHJqW1F0H8ER5cIQmP/9ITJokDv3kTJnxs2pwfNk119Pf3ulI7ir23gRUOKazRS/NI9KFCq9gqTOU2SJj4IzvXfrFEkg77E47EMh9a62paPI+DpWjIT8gPaZNLQEQGyguMHphPwOmskt+mxUfeNkuShR0OE4W8iwegTYuSdBDdM5pK6z+hJ3EVMATw0C39d3JnwE03DCAFmiLXfO9bZ33mRP+8jfnMcwUimGNpC/K9dWQtiGR/kowXd1yNm36NdJa8uCVBAm4mg1ocx40W8MdDBWGceUYfXCHN5DtjekQDNxG/VwX1j3SBTW0NP7ztDFdBJY8BPdp2kiixwC3UL6ZQFeUgvBLbnLu2ciriyS36suUMoYDhAd+r2x69m4w6GNCtWTH9skkgAkhA2/Ob3Q9ngqX6rCoz4OY7skYm1aJhlbfKUSuP+28tb0JJCP9SzoMZT1H/SAaIF0IESI4uRHHR9L9BLLHWLnJrEj5u4UU8wgGz8jfuSL244ErmCk/l0DHBSl4pNA1IkR43W9cO/Pb38q8ju15bYBTnoN0+ZCGWkoDKohys2gA7yy51lg/faEctXUWxcmyc9tPoZHdh0gSXjbGPwdeG2PcI7BbUWYqG5i1hT+2cV391ABYhxj4ktQmveCn/8p4MPXk45IuMNVQnV1RLEY/Kc4dnT9HiIfnyNFkeLJcMwivMsHSKl7kqhGj+/eKE89V27jCKMe1apnxQBD7Cpdwks/v+nR9Dze3k0Qm5w9T7X5aXA+VK5bUTqtU3us5Ug6NXAFDv1uujB0LRYqm5lrKH8pL+497nD33KT4XrGplVDbL+1/Dsfbdoh+Ob0NsO4NKsB8GWWzmoXXB2edgyvfZAy09Hh4reMoFmi+Wz5aK3gigE+moOy9Fwi/33rolbhG6MTpSwBksQOUl2eIdqXoRE9seCR9dd9nQ06/yW5T2l2Y9jdvZ7EchZfNAfS7Ll0DxGkB4FlxKbm1LbAR8VcxTRjGPU1702/v/DybJuEWRK3W3gZmnNTI51GIFTdrb5KDyw7261tfAA2hjHRHgJ+Gu9OJ2dOBMgf74X5rb8kfS/Q255cdjrkKimpLfsQ0L5UiwiDF5yaOqEa4q9GYecIsQQlhScLhkcm8klON0AticgvhfRk7RilQfxEjwvtCM7aAAEJmQoTgHObEsA0Kk3Bkqzvss77NH3wP69WTkf3lbSncuoQfn5s/1DNcjHfg+OiTIlUOjTxOc+LXHkY2VyZNkKjPbzlGiPXTcHaiEe3Eqqf4PKVou+dS8PC9CdLPE3pbhwn1L9nlLoAB44NCVuVcd6xfdgRz4sywZNBOL3Fe3A0dR//5sJgpF2Lc8r72reZqucGxX3Eq3zGluescSg0COY7K5DkdyuN+ri3fi9xdK4Q9BKoxdxNbGPPsfGefreqZz6RpRdysNKw3vOe182/sT86rF88pnyciFkNGlrfNRdy5JKdAB2lwr2xBrbp/7uU/iOhtr0E/DoeDPY0Cc7wfmR7Hj1vc8aELyL4L2MMpF0OCwzzKxmW2tCXAnK4jQmiEvYqsaYy3pkHlZytnmJz8zs2L1KKaSZUQ+wmgqfuojVqjZQfMl8HxKZCRMYvNdYCeMIg1pZhzLlLNZDyT9R0gJ4fJDs4Hl3CmXfXtA93pF2Vn4a5wGuT7LgLD3KXyi9AFZ6WNUpK9DYceRu48NUYZeuT8IcrIeN9TtGcCXST239VjGZNSzM72SHoODl+JXOyETqPo9ZEbbGLFBiqYGwqdo6Wo26qzTROhHN8d7BAhGGNLT6Xxa+ME7YGuhsIZJJe/aPUz13itUqz48pObJ1MPuadCs5wcvEQSCxXTb4hPo3iDnD7YSxcgWNpIt+JwvKwO6YexgPyesQuzb3YGGVboYOQU1IBCcIt+XRruCKuK9YFdyMKWnEXv+Vu+r2AcR0joOU74Rz9OKtHjzIO+hrlYx1ehTA9GKJF7yNEDqp02zRAbehlowB9cPx5jMIgo9e7ohbSO8O9ZoUUQb94GJXse3UgfT3FNcdQ4ayfPgWZHmdCw+XOvtY6Pdk7lLrAhyulWhwgGQufk5LuIpyilc1pCOoCvqCWb6QCxN557EdFlmgEXmrsKW/c0gUYXxs+n+sI6Asq2MJiFiDadODNG0mYg/Ez3EpG4L+rDwp2Q0XQ0zGANMqKxAjNkIc+PNEHkxPUohtsagOUgIeMqsBPTZF21CJwE2ND4xHTqIarQxO2o7sEgWPBtlIm3rIvyo9gvDlgEg+VGZZlZq8GBsoMMrHqiM4XruzHWgW4xR8ZQeTapXBfhtDkWInFILEYi7qppS79HiZujBLlJSaIgnEFU+fqVH0ctpf9kz2fSjsZNwQlnZ6bSd268mV7tOZ6ewQ7+xV0kmlONwwWoEvlXF3+UXr/5bug5EWIR26AU4qRkXG2tycw+Y/wwF2eiA9g49pN/ZfRdXSMsZYiSWaLO3btUGd+PHlKOiDLCeAoxXsvit4KJ+kekOGJsMEegkPnQQrYAkdfS7k2VyPEwtm5wPI2e3L1uO6JUNYttBlS7i2G6rsC0qU9HvJw5rDp2w25paHB3WF/SklqoKIPUnnooPzoKbVSxc2oOX4R5WKx7Q1lr2kk1G+GFzmP0sHhfFdlX3CCscDPQYoRF5Bbh3d7d++qXcnqJ2J9Aw2IBqOnv3bQr/ZP9Xwl76yBpXK92vhehxU9veSgSDb5f8laaJtTtuebDeGnrmbyV9KenltMndhxNB2o24eihwAFQH2bRfHH78xFwNoIF4P85/0M4+Vnk+7EYEZVLxgPlBPMUHkBNcEKBfH7X0+lRoAVN/vhe+8/cOIDeOxChuFOYQP/q/Ks4uPaBtUKwFMT8PtjsXycrq3daaAxp0h2Bv6vGQvOY0Hwvb3k26sgOkwZottVxrBOzWKIyEW71FXcvViza07KdQDhCq/0Xk5ntMbOYRssoZUPOckxwXAWXs9TlLGY40/V0DuUYXiTAiAlzgTzTXnYjNVNFNsk9Atpkqb75yMN8AkCjIJK7B5AkpGra5/zhGHjEgsDHcpSx+srOTwcOzzVk4zeB7DvYsgsEhx2ctYzR4WL6V2e/hcmQ0O8IOluJdMPPb3oufXrLE3iqKcwxBk7RyFD63GaUXRDcesC5n786FZAnAZjFnbKxz/oQfYKmmihm90+PfIWUy43kIlhw4UL60w++niaLKW7B+lBnasE59YXdzwAxfpAc2bn0oxtl6YdE3Aql4u4kH2J9eLL/P/b4aKJnUPLX2lAFAS0XevTqCqtDQRWD3HIzxdVkMZEFpWyqsrQON/QkRC8Gfb2ZNjS6FJy/pmrCDjBTlSCT18834HzCrUxYqJlNWh0oScSgsA+4peQO5VmfL0dgz0F0oYABFqFm4kjK4GiC+sPWGORGCIbsJ7bGYrANq4j9rqmuxVKzkuonAXWWEjkk3vzBLYNTZ+8NpShHia0lDr8WuzAyoveGQItQcnXqfJRFQKK1xH1tOGpoJ49Q+j+EhelfPPHVUGZdrDyNPtBPnvOFHcdQyg7zTDggjOQ9LCrXSNZWbOsnFsWKf4ts4wsMjLJzWHW8+AGOIKw7s5f10c/WPKAWBcBidmqLNdSTrdUAlzarTCuW0agsTZRFoAR1FLJLiHhgqR22/4BAbGD+TKCpmyWnCaWymbgboRqngBEpBP7PoV897ve0Q2plF3eXruNnDuZTMwHYKzj4Wn1geyGGluAgs6h1A1a8suKZVEZ8llSuA1cRJOvegw3QRxN9rmUxgLkBkwD1lM1j+luCmOYdMMSPJcyLC9jC/cw6s8Zfz+D0CaRgv0fhkIgWp9i8KrSn8hmoZ4YQZJlQQDrwWqBXucqah4Mj3SJ2CgVnPhIGRDgmDg2UPx0lCxQWFl7POH63UZUFM7oEc43Ej7hB7u62nZ87xABN6JQBe4ATuD8WjhWSQQrYS20RZzP8axNOmal9iEQuwjiFdkpIVl+xIFn+sjCNcgvmMtAC5qZX0usd51I7pW1aCJveQdz5Uy17YBxVkXpoqZxzmBHfA0Gul8QLM8bWbkGuT7k/IRuv+uhOH1d9dvcld7S9iCvOR8j2UOksRoY5Y6eYPxe741sQ3zPuyuPOBQQqKrPjqFhjPSxL+lg5fRE6WJzDtwIeS47f3PPI/JtgJozWIk5FnVOJZ64g82d7m6YQLEtMSWRqcX9Z+vI88zbF68z5k7/VA/+9V6a3I7nDgXIA/BtlZ/IuYb4XjOdw/b60DgfEEEFG53AUmF2/n+z+YgLnL4y1kxg+m/YSx20uq+izl29dB3Cziop1iDfc98pITxpZGU3bgcxrKmvGRDicrhMXb2WMbJR8PhOCdSfe67iAWrAZYObbkNqqN9KuZa7pwVu4QOWKVgCPGkn+Bi5iQqi9xrS1YmMM2g3gSnqAmtChkidczMEc7lySp+NtNlNj2kkydy1K2CjK67XJ6yQ+kFDNHHvQ7BzHol0xk9mAKQ//3oHPp09vO0bfsAB5co5R+PKeQ4UN5nBt7HZ6vft9FkYJCeCPku1EBheH4y1hx521W7K7SvyW6vwJyRZmKunX8Ixs9/FsD69zN6RPXOMrb2I4RRb+6xlcpyNHq4dd4L36lUFegtHuamqj7xUBu9eBfmWAXhvyOjwOxR+0takbBOchqhlWIBemfa3VG0jm2A7Rg5QGONU0SAr7Gtpi1xykrZcwWMwTOZnlC9A2B9Gd0O2Qwyk2uGMLiHhba7eF3nZ96HLqWwDFLnrF7kL7dDiIhrAPvWMZMJ3zo9cxeIh051204rAGqUErCoemY99/lJjzSzm9t7z/mIOTvnX7Hbghg0hjyuFy+1FeHtm8j0kEMbZjMb3ddxany8lA8pU+ytg+H9m4l8SOh+ACcIC+yvQWHtKTIFJFiXk4hJGaX9z7Qjq8ARs8aWavXjtBUYbh9GmyblpJIukH0+V7V36BPfY6RNPOpBsspVu9kEIIpPFprcGCVI7f4Nl1j6a9G3Yxs4gHt6g23jGANQluHVaLlVTFYnxh2+Pp4dbd7NLF6Scd70LoM+kIMOEtQGFY/GG+R8jA1nRs08HQQ964/n56C6+gLnm33fwxw3UDwJ7MIteK97Lqq/wpd/5KhBfAA/oZgVf717Wl5yn3o6LhBLm2pAkPeRwd4hXpkQTk/e6Bl7GTbyOQ71UWdWcWtn23CZwHt6X/60mQfx4Z+wj9moUQ3+k8hyVqJL2w8/FwFIoscaLnYqQTPrfzMfQKZPGe8+m1zndAWzgZi8wd3WTufRu3ksUG5AbDWtmPHX+sIv1q2/NYnGrBoelLb/aeJQK1Nj3Joo1U0Jsl6e2uU+lnZKk5Bho8mjBkfJL2fBJ9bpLwFL3sPwbynMAF+pgR6zwyypWJa0A64qNAlChFXN2FaPi7B19Caa5ANwBJrutE3OvJnQcIQVhKo10TqZnCD19EvzOitB/l+9+e+XZAAKqDxspy4a9x/IOIXo4i6uIKBnKrdNQvN6RHqg+kx+uJWMTyMjc5n87izZxVAcZDisSYKuYr06ebH2M32AkBgTLQVJrO4i6fK5qBTi2ksBLF1B7Gi3ukFkLkX3fDUFgQDtXsYOKb0ljJunSh5jphBNfZAPDwQfByDmN6CiJ8gJZBJ/WErD7ZuC8drT0go0zlDRXpYncHJrMeFkgIaVFM7UDd9nSkZk+EPIxUjaepmuX0BPgsG5jwwfLBNDE2jiUCvPyG3UY+pJ76fpxI54kNMgOLZ3kvni/EYDvhDL0M/G6JPr6NPx/6NYjH8BzAsDtwIh1jca/AAGYNdeZanTKCuKoYG0t0GxAkE9xbyD8WaeLhDXsx4VLC/tIISdJg2cjRXBerDvNqH1m3Nx1inI2nmV0vJEcPc7OXe2COLapHDFkC/aEqPYTX3JipsZrxdLrsfNSuKsCfsYTRoB7cyGMNj6SHuY9Kfk+IScgAAEAASURBVMMs9n6AUZ9tPRors5F0wlniasTffLZ2L/ecSyXET51mMcsLA1oPwq8G1v2hxt3pYOmONFeG5WZ6Pr1ZciYicEsx0youK7vrEwk5hj86zTZVtabHgByBF6b1WP3GN02m1oKm9Il6Fg9MZmhwIk3VzaXD1WROoYONlI+n1zCSaFhxsQUj/Qiid0k80HHv2DLZeCoyGRV5j9cLCl65GTB1LBRP4qOXtTXz/WIx5j4APVdMx2GpZdAOjIrbEbKwcqLbss+JbYmWKUcrU8ZexalyQgO53B6XUO7smEFoAeHHRZr3zLJZgJDmNXUhYrN2OMcECeTzkGl8Au3KPUuO5I+J0MaYLCvGcf8Frp8jHiaEH/vFZ8qV8g4vWX1I/JdIfDgLTqOc0+99ylrH5VtXuc9S2rNpXxC57vp3us6l//6dr6WvYS1SzKNb2OkHqLL+t+l/fvcv0mnigtRdNBcexSx8FLRjd7f7Cd5tIp9Mk/UrYwj2wfHWymVIgQFoQZk2MG6inmWr5bzZrrNMuMYKP6ZPkRcE7J+VAnPn0z5HweBBZXfHhgZx84wulpxjP/JenBdzFtsYp/JZIbFYMYacFviVXGf7Yk41wXi5EpwDAa34B5UhGIz3lUoWnSc+FH8/nu81hF/YDMfhw4PjldmxJqcPuSt/hn/tj5QYd8x9wSRI2KZGaUftwB19rpJ6rnD6C9jORaVaJvljBU5uBFIR558CDnuZ1LAiBtrar3IHx8shL8bmrsXiMkhfpoFVEB57HWfIbcSGqyPdmKaIkpyeIl6FMjkoSgHJDQkKs605NRrJ3dyFxCy/Sk2rShaTc3oVn0DUtUVRWkExVcQYR968NtYVXF6b9hW264nCKSwW2OgBqLpFpn0vbnTbGtg63LcTOddSPo6HR36cjBshlpDk8ncDre0YVUYibJlz4lSHjnaIxz5K0BWBtVS/rlKq4SaFmAmB0ia3YBNV1UvgtC6sSqxgbaK2LVlZW+Q4xpKF2lLVzC6xNZ3ASTdHf1ZPiQ8Zw1F4CeebOQGaby+wK46QIHMen4Gh2yKitROCMIU+1uq4sEg7xtGLMP3pnLK9MN5w8Z8b7AivqvFEl7m+hwDBE7cuIXoSjsHYiLDWiCNOICqjLS+TaTUPIepB9R60mMhLco8nutL6kcaItLwmUgYLXQuddGqBPRmNfg1RL6JDMEhh+s7gLNTq1odz6sZQN9GaY2njCHWJMVbcBGR2AjH7g6rLkY87QnDbLRAvtO4Fe1Qp/ohjTUXWyZR7efhXBITIkV09wsE1c3elrU1kLm1q2MJEAssxSGEzuHorilAp3GkB7tw925VqC+rTFpRQTWA3kdNGgWHL8HIcIYYIetI0V0iwk77LuWWKBCPXlhl2QBvU1g3Vrcck1oj5ijMDEWyUAcq8eQ4aJMN/oawLFXnoi06tUmo7bQCWo5pY/1HCXwdBFV5f1ADBqUAWMoi3yPLpR1Fjc2QhaR1a0pIA9ysj/kRqmA+jcRZeca/5krbLZZnpPSRSf2nXc+kxEicsfOzk55ek+J4/Q6RrwJz6zJYjobdkI8hCpdESQShivNJhww1dWTkCcd/zjIL0k5730789963AyMzk1+wutikCxuCmmTeXNrMY5f4lGBj0jxgePIv84aI0Fl6uuQTzKcMcuQGnXxm61Bg+ExPuG4sbmcN1tI10PJLJ+2Z7mReBq0rDMreAhcuKkb6XxEQy8FnR4axJ8VtGVIYTSURpq5dnDk+oQBZOHw1mawEmpBZbvDvaMIXzhpjTUucQItayNIs3V7FYi5i2jamCGUzepWl/JYGHMEhDoa+B8CDMSQF9tMJ87DCr2pF/uSanz395/987nG0V8bvhm3b3/NancBo9S45rWfr6xdcifv3lLY+CGbMOjX4h/a+n/3365IYj6VGi4kQse6P3PDB+30kj5KcGtg1czFAAB0+To5WhhRmxWNmCYofmUCarlhDdL+18MT2PbFnOxH3z6mvph2Q9jRLSIGGGRcNrXFpyZeV9FoxOlt/e+/m0E0/uKWAo3rh5Ih1YvyM9hpJaw6D94Po7ZOK8Cj4NYLROhvdyIdKOWRafcqfv47P7BsYt2oK9cu720XZqUxWxKJGZCStwV88fTpox9Q4f/+Nw+/fH5G3om/ZmC9+NVeqxKQxN7BQyhNhw/SD7H/eINnFxtJHPlxDVhMXj0uxZXCThzrszMDYuKFHUFoC6tiV6cB9qPpB+d98XCMijvtXt9vSDjvfS/k17CYbbB5jVMjVaz6R/d/ZvMOkKMMV4KN4GH/JJ7n4sTscALh8MINfB+MNzZ7DZuyBlVCaEKKUGd4J+6ijK9gf7Xgbm+0AsyJ+2v5f+9ZW/yypPuvvQkcjH4BnuKK6rEkTnJxsOpd8/8gphETWAxfan/+n9vyI3+hqLQ4K392sf9xJ9fiZoWv4IQnc2vAd/sk44TU4W/2hIGU6KdQCNVsMhy3HUbCCepA5FpxZ0XavBVaLMbKACXAt2aN3GRh62NlK7SHkQZ4eKrtu3OoIjaTbMCp2Sy4aMytIOCGi+K2XANuOYqoWLGrvfQPW94FiLU1lAlDuCciMDI9fUJ2DVEVMJ64EqiZqtFThQGhvDyRWVpnRKsZhcvKDo0zOGJaiL9viXIwjV5tk2tVr6nT8cXxhUfMY6BOSJGlvsTteGbqZ3xUvnFtsbrHiClYmIwSKURLl2NorZXeIx3DP/WTzVXxKzp/BLcVJL0RgRoNrQTc7xyDOjOJFmmRYYr+PrbPIN8cXWSNthAnFPcsAwBSIdUzSDquWIdWINSTCVvNYsvBlLUB1imCvXYm9FBu3hjWV0g6jMPQ4hhVvbco8lP0O0coc1XNwpdU59psOWKd/EbdEX51wbv/B8Gxs3ZFwcRtXSsI6FCdMTBt3m88uuOO5qWUvQjYysqZF6YLSvihj/9TjSWhG1Opx35tuQcn5x1YePe4k+972TlOtDfOKg2kCfnHWNzvGqUM5BI2Zx4gxTYtHPPMbHx9MtINgmwDFpoN6EEH+34Q7DIyRsV7QyWSVpaHRU3waHMtjd9lnG0bpKDo5oCsZGL8JBwgPI+UuYHS1ioBOkggGexL0+i5JKA7lG6mNSaYnsQRQ1cdA1g5n4HTEfPGuW2I+h0bE0XUpcSq6vIyRQm9zhfeylYknsQLx2whxuqTfDn+Gljc4d+bHxQy0Hc8jJ3wfw6gpBU50THXEfkcyEuZgiLOEorvTDs/vSduR2x/ruvWQl9x2ewPNzQ0sO6u3IHxBF7P558uJYkLJR+8H7iBrl4iXGxbWQYbz7FHYD3jteRYzPDFYcdQR7r6nzFgFrG+YA3yWrTCfjEDpOiHeMa4E48BJlnM/YBDX7SAk+W84xfvEeCvFxHEFCLAgoJyNmFwX3WkJsHcZS1obFRsfWMLTh/Zl6mieREF3rPZwHPpcdKkqOc40hLv6z1OoQ8VWi0gVjot0fddwr068+y97nDuXGIgiP9sYRxXNR/gy1RSDEyrEIdFsz3FT5eJFgJbKBGPjNyIOWYJ/GY9sxfSPh5E5N9UDksT1bOn1wkYwZOixHZ2mHkumgGQbsAKpYaY5kJ4WaGCzeyxmakT03k2yiBaOH+JRhAqHQh7iHiiuDxSo3W8saSnI45eUKwh02AL1nNb9xApWGMAfW4Reow46saDIATIaQ1Vo28oSik4PuhJ5CI1lKDDhjEfgrsThiOO75ZZtUxNVF5giYc8vRYhHQhFwjA2kgZOLXd30qvUwlEvNKf+kRc0FKJovyeyymr4OCPIvC6uRnlJ+nCRarHzF2Ilf4rPBMMzZhE3cOaZfweoxwiFouYkVHk1w2hUxfHjCDfeg85rtuBPbQRMN+4vAH5m7FfbSQcaNgKnLxYowTMgbt7X5ujq0IbxoNFhFNza67czAGWs4cY40E0rS748aKTcj09fQINAvm4vY8SS5L6ETShcouoi2if9BD3Is+6bnOchZKiNcS9uU2IhHVyemP/2j2mseanH7NM/MfcsOtjRvT7+59iTgNih6z1X7rxjuEq94AWg2IZkQRrStHKeXy0pZjeEkzj+xfXZlPD2MDF7LPQflg4Bqy+E/TJN46j2UcDtWIBa/seApZeHckIv+450xYCT6LbtCMg+Y2C+VnOCk21G5IR7H9qnz9lPfv3iZ3km0fmgxuUIGy+qv7PpUeBnVNxeYE2Umv3Xg3XeGZsTbcEJi4MezjK+DBSL/qE6FExxkSp0jCNTjGniR4DecU0/0GIbs/R4mcNPmE92sdXpdVI1Qf8UE8EWKTP8QBcVgU+Tvtv2BXKEufxml0P1pz7KzBxu8+YYwiEG/g+PkR/dAiEonpsk4n19+81FnYysL+FKgUD2HWFNj2OEkkA+y8z4Eg1kR/TOZ5j4SVVixCz1AP15zTM/1Xo1TOJ1oPhXPrBvP49q1zOIg2Yh7dzSJZATrxElDlFGqQKUC0Pq+OQLGntzyZPgVahEGBxwevkoKJM7HtaDjCeknb/Hr7jzASwNwYV/tl3sWB5t3pi20E9iESd+NB/9al75Po0pUKxzpZAYwbdLCvYQ/oB7+CkpzwSA/z7DfBxrmIQkziSXTYIhWT6RwMTMhDFaqIHEBkY+PiM09a+/hooudhHpni5rrhpqy4hcJpgDRb0gFg2gw4G8OMtL7qCqHC15H5QLNCk58FfGkraWOHId5GvGVYfIM4DxOJeAg7s5YDieGNLtO7TJEjMhCOUAXI515g9vbTYZEHrrKVl/LdbtzTxpyvB1+ld/Q2MSqbQ0k0c0qQ/9PDJRA9hOB2C3WVUZBhL+cIEchHmOUW0rtFp9IofcpXp3MSpNuQ36OnDqWTCYFyDzlihZlTKMBHmCSPG5g1y+jflAO6innFl6t+aem6cwRhxp2zj7w/rwYXBtNfdnw/4vVf2fZ02oizKn845ncObgWtkD98JX23+3UsKLdiHQXV+Zj8qZ7EjzVmDwBXKGThAgQqTk9ZwS0YDk4crDW15JSOov+sp6iEHl6zkCYIuSjDDn+oaQdz2RQOp6GJCfAld8A4diKeWL5zNv2w+x04M7PJc/wx2K+Nwmr7gFmcwOJjrkEfMvYBMqz0kgZsO3Dt3XDgfJfcYVrY6Y6QpWbGnYC0r2OJKyrkHBiPu5SoCy3I6EdxVEp5TegT77OgPgCfH0WNTvupXddKBS1JT7nD4Y7hz49L/otVf++everDtV4qYwe74q4xbUw8TBQxAnMUi0GxJGq+0jtpAj00RBdlE18Hjrg2VAU8Biz/YMNl7YCTpxMrZGffcF6J8pRbKa9duVpmTGHT9xEr2RWNoKpHtwgxK1PSgJNlyxT/RNnO7d52hzbPc7lz7oin5t9kxM5HAoOGxefOubbFj5RhFW94nZ/BO1c/yIu7T/aVtxhHj/jbjjcoxnAu7aO0/EHifrawiwmTMY2O1DXcB6b7Ynp4y176qC+D9rtQuTgmdtVj7Zk/2r9jbB0vxqAIP0oWtSqndZ7cX3JQJk4UEonOomJFMN97HeNQSv9d10roStEyAiI8OF1bPuPJe1Qv5oKT+e98FCNWlvhCxVNRnPsWAkNyzxENR2SFqch0PEJkZJ5jB0GPC4Jxl+QU1bNoI31WFCvAjBmmSJ4rylpG/t4ld9wd5vwnH/qbp70PffGhDxiRFVaUSk+n8Rsj7VhDcNcjG18haQMvNaGwNNgBogPtOKJO111CLGnG7juTOqmC91bPZZ1mbD9L6WT/FcxPyHvx3n6WhLJp4sUSE1OOXf0cTpYBYrLba3rJpqfOELLeSRKKB0RPI8bHSiTXCCabQDGdzVGBxDSKDff4OM4v3dwMy0m4xJjxMnISB8+Jyh359wWhW7CT8RVaQUy8MH8nBhGJeJYc5uJUF23GenH38vxt/kF/GSmIA1QBnD5v41h5b+RMiE3Om5MZkPoQ+lv9JyOYbg6i0uyZtf/eRth9eof/YTy9TYbXGI1cILHjBCLd6NwwWWFXCJuuTzcIvHsfhAJRySrGmsIE+MFUJ3I0iiuJ+yMkhuhMOjlFsN0Y2hAukmVk8/PM9yzK8zx8ayUKaywTMEiwIWOyfrgZ59YUz7qchrHvtzQ0R5zMTRxhvcyfc5I/zFbrmxxMJ0gCsazSTaAbbxMbtKDIxMJWKdYX04PY8/7QFUTEIsI7blMl5aYcCa8Dcw+j1Hek0SCU1lXzmX/Ox/39aEU2d1V+qy2WOjE/6kGKhOnqHciI9djHx4iYvBEBXG0oqcCr4kntRQ5DuZQz8i+0TOimtaA1THcqgiKa3Z4Hp5xBsKNaFeSmrSRTbyI6spTEhqvYvPsRA1YWVco0c9EolVTI0jI2cnHFmVosI9sxsTUSsz5KjaV2IgIbiZhsLd/I+WC0szCsR2RsengCVw1SnuhVkA0W21K7GfgKMrBIcVMWpYh9RHTqJ70x08l9hnjFbXOL7OMG94G+YwK9lXH/3pcGxntfxr5Efy08YbKJnBeqz8bUPvDf4Y3Dt4y3jprW0g3I9uvCy9kF7v2tZWoxQbjByt2pER1V6ndUbYkx75qxdlUf30vRzIfcGk5ri1zgYWbkQaZBGs26jr/9pB1249CrBBRqe8U25P7F1DHfAW4m4yO3ph9B7fYnGmor7Zv0gNxN+/PWsWKgGdtKt6bWhvVYiCi7M9BLlKVGDvtK2ITSQq7N24myRGtMncPWExiAdlgoq+YzG4yP/120+zef+O8+7pQ80SMMw+WxqOBpPdK8L/3XT/xOKEf7qQU7hfXmET77bRwMnwS2oRQLSTupdkYdImcgq+lVLU7/+WO/nr6y7/n0LMkTTTX1xIqDXeIigkcJ+LMea88fH/pi+jI1ap/fcpTrijHR3SLOHawZt3V+JFqrWARxoCAaz/3CtifTHx7+Yvps2zGy8felfgohvEL03W/sfzE9s/kgHtyqCEsYx9Rpf+70aVXHrUzyImgN/9n+X0kvtz0BXPguuOBS+tT2x9LvPPSZ9OzWw+7a6To6hEV7w2O66vp/8EsJlyNb0NlrZjoj5lhYthcxIdfuPI3HmaveyP30bu8gVPuPjnwx/ca+F9MTWw4S3VpGLivgq/wL5Zd7babQw+/u/nz66kOfTY8QjVmBbH0DpXPSsGXHOZ4vUTrmEi3iI/P3ZOvB9N8++ofpqU1HInxYQjzYsiP9k4e/lB6jFlQ1u7NV34F/4jrnKetGrldZk2OFKybS8dzrZiSG/+qxr6Zf3f0C+QQHSAaqTm9T38DQcfOu9dtsQNf4/QNfItr0pYB9NHb/GtGmZonlH7RqOFY/8kOvH1i8kdE76Sbw1uB4cQ36U4zbuhyFtoRECt3cPthKgOLO6GDyHPumw0SEAN3KflqBI0jurstYAraGqcV3leGDDvi4FLObXj6PvLs96CA+kSkxeHDIShS0MnYh71eC5l/KJJoTuhImSx1m5FqaTR+ty1183x8nSGhCIxyjxiuvy7lHGWEU0Qu+L/S1smY0ImvXfbf5/++tNMTWB2tiHugrHTJJvxyl3rinDGGYHZPmC6UecB+01v5YP0wIEDl6LL5VvZCL2l2nwchXIyCdnxKcjFVFAO26OzBCzrd5sCJVYFz9uJFmLjMR03u7mF1QxezW3lfAq1Je64yMZ/O9ZkstYcZAecQipL0xF/HJ3+/XAxO9srGKhg8ah9MNU/tVy8gY8tg8JSEnwfgYxnypnD5AUnXUUKIXdgQap+HLhM+OpBE8qHYySlyyNWlTZx1x3jJDtZCmCFcV9XiWzo9PkcfEolC+zW9hbo95cot7MxsGqg1xbwz7yHy42ydAYqiapk2gbdHIETi/CdG2Pc/lJXLv6TbrXxeQiGaDIuFyjNKvafpoQrnlPuG1KJeEFNOef5Q0b38Y5zl2Ph1tIwAkLcIQJqfNApNYMoKhs8jmc4iWwJQjohgVOj4F9B96UYwH8+AIZ+OUjQ3Tw6gyP0Dy9aFXCTPSj54wsDSWarGlD0yPcc9ZwoWF2oNGHMB7jmyc43MGfoWGRsCgohQ77CLPHua+dQTiaXEapm22U2JXZvc6+9XPXAxIY0gQM1So0ZIUdCkBccicHuR4YJleTbkY2XoZcaVipSrtWEeWDVzV2PJ2ohVrCDs4WLc7ApAu9XdiWutG/De1zIoVcn0cEEVkvDc1QzTF6fYIePZUKhH6rpj3EuUcpkVrPDXWN6VyBqaTaiYOgFGQKsfaofV2ukBiABl87fANBLLtbNoKlnpNGsGbdxXRqg6T13qUaMdBZ1k/zpYg/NyUMJ1wPAFIM2KYwatchXy6i0oadQSEjWKya8dubIjFZsoD0YB0k0SVIeK1lVn/o8n0DzJLD3COpKFMX8IOtZ6aUxuo8qJzz5iUAWBMsrBtFT/mEa68CWtRW/OmCAbsIt7+JkgGok4YbLegwsp4KL44Po6V6YKGNu8G4aK+riYNjIE9j3JZTV70Lsy6C/gEjIS0dlcUbs5RvovHHGJrYSl+LUDgMiJ3UEMejFqV0LdWbE27oj3AxdzqxJTbm4kWcBiZjDv95srNaXvLViP/0pmRSyyyIXQKAGqREAxKm8Fs6k6Vl0qCOa0xdr+U09/hsHJcCF6ZehYs+XODZ12MwREUJTaD7NWG9m+e6Dj5r8N472bwjuXjQGDolLbpS103bzKErsplgtE2phc2PUFJymasDiPp570niL7swTrU7Y3jWS3Y5z+78XGgJpqD4/7i1jupexLcFp/N9q2S3IwVYCepiY1k8YwVNGINGAo4Cuu9SgwOfJjtsgfHJFbizXuy+aG0h1I/FvY6ThbQJB7UnTWtUVR5uHgSrnKbtrCAJ2/SYp1XaB8sPm73j++INlGpEI7YhfWlg/BuxRW9odH3aDHhgUiQZpa5kPdA+FPszEs4vqpZ3MKZmzHVNdZPWO9VYM03AZ+4TVUyjBWdWFr2kgpYRXJ/VaqFyEg2wby6r2ozetcceb2zpPBpeNCHm/3Ta3p03UF0vv3AelAxnLyDXoonPEnwoUWf+3j9dt8pdJFN1BXeHBGxszUk0KDM2n4oPxTZcuz/G+rWA7WyEW83BdvIbDOy9/mWw6mMhPKxiZH03c7XCWAEmY5nB4GYZrjG8UuJPn+NBK77WLHb7sgFPLTQWEjhGF7L57fiXSRoSczEjkmUVM2EXOfq83DwNRvKkVbg7lbZtrBuC/HrU6zSHpxRFhI2dlRC9d9eznlu6yORLjiGZWaQxdQHx5U7aLsWtfZIy26KDD8XZjhLWvZosZjp53k+OPdwXsmxPNRL1uMgOdb6EDiZ+4KbFWGrnMcs+QRRhetANhCAdojqKD38uBWHjhJX/+P+FXODAcCehtiW63Oe6+losv7sc5uPkLn1UOQkn4bpaJV6YeOjZK/VphvVQPWtlKdDQIAcbmrDTUJICemaJ4ovpt9CQfaWirDNmD3LGd+Xdj9JAByBYgP16SJ5DLPMZYw6v0SWeG7TUYj8YMRoncHiZ4WTz+9+iuysyvC4TyAevbznBQi6hR1lkbluIl/gdCw2R1uGWV+NdxyjwlMsoCkdmljvFNM0XlRWVARO0VlyawcHEb/YFfQ95Onu/hlbeyncfxbvJXp/BRndw+oUNiAYqDtPVCqnvg4/Re5ecf1d+gszZcB2+z2tMJRYaO+MLHMX8Qdy47HZhUhJfIACg0btYlKzjpIxBnCwIG2cIpWx8HHFfTfL5NTs3vhkMGHG/GUfyFHY1qOWLRcr63tLR844kWh/7sx/7H9id76vwcr7DooORfumbpI/1KusaWtQmodMQaYih3dunBVFHhGiM1nDc7gGD+wcydgecU6Ecvv+7sAr0hjIFgfP1xrk2yxhhPtwLoHksTt4jvcxKUmcoNXJ43L9DL6Ra3g9TWWaaZ6t2J0d3JeFoA4Q5m925I86PprT3213XMutmHgJIUf4GVnxKbB+rLy3e86i6JEqBzc/O3INJwWJ0jPg1DOgbrFFlcjloBpb7JfWcV1JyOw/7HwXhFpFEmREMng8PxJWcgPXPt5O0sTbcF9AVVEkNYkFprmmKoxjOine67scECONyOIjUxPY93sjhl9QVn0AML6QK91pQtRhIQ0jAr11+yRy4XBwtXdvnwOKcBZHDs4VUBVMYrg+3JXmUJq0QTsdJkyUIMqp1Ea4glNEX+adJENGGbOogcpuZhSki0hLhegRGViqS9E2GbmoAsdks2A1wZYQRxT4nprIPBhnoxsXiHiUCA2es2iZr72vtvRI4EBWdm5iXligOtbmzVqjTc5UwJwjBurN9vDvIJz1R50nUu8YiizBgFeB6Z5AN1vGYtJMPMzNsVvpNM6sW/hIDOijqTinroMU15v+6tKP8YHUU99JeJKrEfZgaMY0c3EWVIw5nu34xD5Ne8Zwlv20/UTqR9dynC6jJ93CaaXZV6BcRVGrOn6PKi97mrZD8PPpzO3LoXtkHlfbXJxGJycA9nonXUcXXCDS9gSIyfNKEpNg5pBM0o+TtH2yM0JcoHv6rj/H8f7w8dFE/6FzvZOTzFCG6JE7wYFm5VnsaonoIMUXrSnrkMH3U/9VIKcZlM/jfefTlobW1LauNbadq4M9dLY9fevaj5lM7oHlJTgrRMqavfP0WxRH/saN13Fbo8g6+nx3COi8HY2b8cgBIUjO6TRyvfWMlokLn56TBErSEyAmbwZuhJ2OgeokIK4Dpck82WwoZlmo7978IB0vOM9z8cSiiHlvlWBIBbGNOHPacoSklz1AEXJhutTfQdG2m9wHAqfPjoVK2iFQHtqIPzHS8MpAF1vtRNqzfjOoCiweJv0i+I23gRaMDTLofoXowA2gIeyJLKrBiTGgEK8iug3AIOh7NhCM15a0D52jEUK8DXbkVTLSmgGw2klckWJlO0hkV+mXEYj2yz1XyJXdGw6nrQ0tLLzFdBXlUmz+Bc8JIkD5R1ybo4zOUgXyPIaGOZhT+1BHun6jC8sMiTz0zyIOXXiK3+0n3kXOzOJbQuzsvkZKHqEFzHjoD/pFz4xfivkS6s8xyXQI51DSI4GkhLDzap8DLg5YlNMFFHso5TMiX12glgb68a030s9vAwuOHA7oInMIHXCvmEDmx1go9hWuYbchn3phEk4Pk52H5hZK2R3wEkcqJ9f55Lg2+htNu+fXAxO9g8qtolN2LJtBuwUyGMrKE0ArfxrwUmX6IrLip4oX0gsbHgZZoAnuvphGJqeiYO5DrGZDDk5WXUr/B1afqPFkk5BdzDW9f3W6HS/g+lf8WcZZ1IC89+LWx9KxdYexCJA5xQQozDzT+jCJBA1pbP0sEZtzQPTtpaDz3piwdyrPp9tjg3AV8Vey1e9zlB/JBgiuEsgK7AQL5LBinmEhU2sWr+ZniCL8BDAdcsi/K3iLnN1BnG4oS0yEKAjrIOxPbnksfQL5eJGF+waZ/j3oE89uOpx2g+fei4dycQ4setzx2bghrtGnhwis+/Wdn43IR13uX6fN/X247DHBunuYh3oUVIMvbH824DousJiqij5IO/F6Pw1RO7E/LTuJdeZ2wF5Y7gjjL0UgGtNL4O88zIKS6N8ofT/Omacie4wt/WoEHeyZzUdBNjgcaAh1pEP2YeW5RfVxz9FS4u5jBZg5Avn4yDiRCDi00smSdacYR6zrYcpcEPaQUwqwsoWiL33wX+K3PsELLY8zPocCE2cjlrYrJd3p5c3HQq/qBWa9b/BWuoiJdRoDiXWG4ZwwRncmWoP0YmRnDdao57c+GnrINJalZX4myGh7CSdmPYq1tQuuQE/X2G2MScuu5xZrHA9M9HeujQ5x14x2mCAXI7WECOcVpqKK7baxsJatFwWXNDAdJYYUiC1jhJ8ODONHqogG1DxWxNbI+NJKZTMlbZuU2+L9WBe0W7NmTXYTY/iruY9QIzqThM2oZBHV4yCrKa8M+buCiMEqLBQWa1MMqdDZwQ7ETfjJDrdARY8ibNgR9MbXCgcFEI9x+SvIryuY6ypY0NZbsruGModM6i3ouMRJRSqq51Ui9hgIRXg0/a1FQaskm8dtvxqnWEU45DJOLCO3FVZaqYGwy7Ga1OjgQ3TioSwud05Hgc+xkViBvZTrXeBmNlUD3aEjTvGmBmItYVxMvHEQbWMJslwVMfrlEOAiY1VOPgMd5d52ULFskevJOsJhV4OD0MVbBdamAL1mdqmwe6OwnLAArdvrfWVu3qeY842DtxMujOJQrhgvPs8u5Dw7mDucY5liDWPC3fHaUssKnKRasC5ridKcWKhEBHNcrTvPpTFPvmB848H8oq86pxoxKUtDzqT4oPMEiQgjWclnmqOr+ad0aEaYyAosF34+fHj92keuD9FjXsul7JhZSP4oQ0ajIBy3y4CVcDA4zMIfI3bFwDLCKthaRQKA+yMfK5N661m2tagbxZbpgIaMrKYaN43bxC85rKtWkpELLZtJNZ8lkzjIAC5EVrwih5fOIjdOsSPMwE1Dt+X+c6CwmUyh0hQqtg1wZmhvxG6wnQQUBZ9H3AmfR1e4zyyypzuVN1vKiWDZpkdrmHSdJgtwRNZmLH7rNllPyTHxUKk2hzTkTG8K0TlKtlM0iOgDcqxu9Uz0UlTxKoLo0DF0v3iISDBlYB3bhAk9bv8LKJJY1BkFFwx35TvnQTRh+5f5PnhNW5glPuEvg7SAKADKT9zXtqizGPMCa6b/3if7iTGQ1TpnfOaPpStj/TBWEcrtDaKJvPAzW8ZLLV76VfQViOETlMqJc/RpkgSYDMLR8WGM6YUMxNs4/oFv771y7Y5oWRCnDFnx8JFCSy5OYOaAgoNncv30CtiZ0KaVxZ3Xjzp+qXPqzoU+KXfI2QvpfMRyxGdZQYJmTF9yA8vKzJRMpZbi9XActl0muGeul6p3NZi5GhgsUAywp45TPNmt3obakwxzRtKmx7lDQo5cTihN+dNvmgoaQLptCg1/cK6PaV8iA4rERECeZkhNHCIjqxrOUguQrNNgRtQ4tYpU8jxCxOF53itEDv/i0RCOInsWbWGh4RXjHlUEoAGjzfsRlKUptnInKHJAuYGTXIeJs56kDG8/tDQcsuV68mD1BZgkY6rdDEqX97Cf9skQhwYcaKVweWOURrm3cSTuNnqvhbKrYrzqyd0tYceZBqJjBBSHGnaQBhTtRQh0DCV8dBFPNPfT0qFupcJbI5gsQWFQVnijR5GhFdkiIpFz3TlrgWFswL+hJjPGs8cYI4Fq9YKGLk0jjXnJmE50NOZHx4/jxAAwZvzlsIK3C8J+uSjiObEQ+A7GUIuZUtqYY3FN4smdwbOrYaIMLi0ymnqblWQUa4LAXVhcVzTPeEG8agbu1HUAVdWhK7Lc8aEMYzlaSpsKmlny5YwvsfwEzcloA38UgSEWbbTw3l//QUSv0qrpqpSJeLr1ifTS3mNs9RXpB0Dvvdr3Bkqlk+hIsEXz95Vdnwau+zBkUpDebf8gnSZM+Es7j+FYWocz5Xb66yuvhXfXS/JHG6hVv3Xoc3jj1qc+qk1/r/142tTclp5BUa1gC/tWx+tpEkXsc6AjCJXdAUjSty/9MF2lxIv5ncG5IcaI7+a+TkzGUXlCbiFL+BpfHl1/hMytZ7ETN6Zu5NtvkpZ3GWU7U8wyLhhNYwEakqFxtBqx6mWueRr9RRPoT7uOp1tAfbzQ9hhBWdtTF23+bvtrZIqd98ExsRJMnnvaBKY42sSX0e28EibdyPVQDYNozVT75PZPpKe3Pxr6zVtdJ9P3O3+OMkgAHGJOECu7hFf4n6v5K+H6Ey+j/34cXBoKDjEPpnOoaV/6yp7PkqzTlK7evp5e7X4PWMRt2MX3sXZW0ns3z6VvX3sdcCcHk/aw8OsQtV7c+mT63M5PwAwW6Pvp9KOuX2DBAS0heuJzsh1iBbHLZ0WWk7sHRwB10VDb5njybTAbRdg9NdvTHz/85YBf6cTv8o3219MVwtHFBfICcy0O1u9Pvw30XzVOSWHB/3fgvYexDLlYl3legTv0GsffX6a/7yb0KWLfd5Htv614HZn0uKpxS79FEbOp5UF6gySH1WW5sjw93rov7arcEEQ0sXmaGJ6xtAuvXyv1pCoIBW6uOA9kH+i8QaH0DSLZSGbOtrLN1JxqRoLDYdLUn7YRctBWBSeFA26lynUhdaR2rAcYCc7ZiuL8i1vH07WZruBcIZIxyG4mDmpGDPd1Ir4Ct7K5iaJoAsECOU7EoKl3AiNFiDSEE/m73MvpkcAUYZrIENpXs4Xw49aokLEHL3MdHKwN6JNWvJsmvFgv6xyTsFAwy1bMLiGnpG8htnkz78j/gBKBexo25m9YCr8zbiuefRVQ2m2E9u4sJWyYqo19PLeuqC71kr7IpsT1EqPk5l/eM/nxljf2X+KKH2guEtzJDEcdJ2iwOK2H67dRedG+z9fMEFKyi6IaO9L2yrYIIRnGXFxy821QiBGLbD7PUD7fSij2ZuZ0Fsi+3S0b00/7TIz3QbbD5/iXH4i6yN0onJoSPf1DWvE+tlEObQKPi6SIzmyoXQ/T2BqWqXIiRbbXr0uXe4i8xDPrPQvRWbYQtrAFh5ah5XXoVG0V68AU7WErA+8fenc41joemOhtWAwYzXRRFiBbI1WFoskLNpzca56yDDqBE+pW6KS5ZYpUHPZl/sZEMPFCeLtoPBQ9lDl9iJxXEcpD8Sd/+Dy0ArZkiMHGcEjUcumVjM2F/VvPsQ9RRHHI4xmc767jex505whRJ94x/cjVy3IHJwFOFtcxcnIn5y628+Bycid/uJAme65biXdf5LnzIc9mD3HBGMcffeZeElv2jS2RtDNum72LR2c35SQnN9N1JGOfQc9VsG1fdqVRGLSDX3FviT3b3WSmLp5sYUEkjInPjXb4XC6hNfTRwO4cddghPtdpOOOz7Bdzp4MpW/h+b5ttg7s8upXmWw8HC/s5nCC7Nw9wShyTIHzGRJ+JhGhbQoTjeTHG8XzbwNla8RZoNwvHfyroy4h6+iyc8xgTe8+zjeEx0yZKTclYsfxEv9S9cFrF/ETj7v31wETvDWLA4BhGJC4ia5m+ZgrZNKaja0NdqY24lUpkynN9V0napagAdy9UUcWLuoKo83MSvcU9Vz57G6Tca1TbOEWFi6H6EeoZUU8U1LPwFkphHA5IF46SsyBX9bMr3EIEOos9ewC5Wq6gVeQadnWh8kqxgmxlVxA5rZdwhiAKlUYabtuVQ4PAs1vfOwoSALPRPtiX3qu5TExQU1QB6SYGaBklE9NB1nd3IAkBUUKHicQyQATmaaDulukTwUBRekj4jCZkfJHEukb7IxkliEYOziCGtYj7ZIoaC9zJksOz0CVWCU/iUEcIDKAYj5U0MYk9/XYPMj0ZTyxQofrGGQu9b9zNTsZFGYfnWX4UBCoxyOr5AKJU9pd0hdiwYwtYGnox6X5APV5NsNdwyp3Fbi+D4SwIbiEcRvOUUVqickYR4cQ61SZ5fwV49ZaR5jB9nsShNed40Q4TPwz7znQmOoMoGkqqxMl8uPWIXxlcg3G1n0Qn0h6WFN7WLqrGHAeXvwTLWbcQgiP4MNhBbbe9suTR9ds309n6awDHUiOMPNwuEpN0DjpuzvlHHX8vmd6JcOUqVuwiaVg4jXkUxyukjI0T2qvzSFe2zp11wElsIwjMsvXWeBKKuRb02WYUGO8yRFirqACNdcCEkO0zR61Wa0VlCh8DkGu0HtpF5MWYJqwWFu7dAHpxC+KQA9mHIquVYl1JU9ShnSDwqW/xFknkjdQx4hwOvX692Ne1iORjueOLVb9sdy3x/q1l68OsqtOkh4CzCRRIt2AnRxydekCjNlW2AlFYTgDVRLRZ9IVGlMtCuM4gheEGyOiZwzYf9MtiCG+qOwS9CC4H8Qk9splFWol5cQJnVs8k2Iwo206WzMW/jTiaWgkMq8IkJ15+J97RySnCd7Eq2SSZRwMK4pZKxhlz5hTn9JCxNoxXWa4a4hLPrBGrv5patpg/tZz00MZqxIONxNMorwtz3YsjahEo7aUcNzap26LPTRgI3KkMtZ7AOrK9aluYOsUW7QY1QTSMJfwdMo1CxArFvc31G0JBnyHP9wYEqy6yGWeclq4h4qbGcF5uI9itmgy1KXweV4CHmQl/AHIM81AEp66EaWyp3YL1j0WJot81S8oo/SvKxXw5deK4rYCi4VKHNDCTozm6sKUWmUtwEM+893hgTp9NBEIIT9izYWf6F4/8LhyHtDpizv/Pc99Jby+dwvowG6UzdTE/ScbS7+19mXMoZcME/g/v/ll6mYA06xhpS36/D2g9MOx/c88n8U5SlAHi+pNT30jvU4xXg13WdkkdG3I5MpriEIPRUNic/uDQF6L2kLboP7v2Y/q4kj6DY6wFS4foCv/mwt+h6BJMBhqD9/lF7wfpa2e+SzWPfojFZSu/yK0qx8OXVM14GqfXV/a8iEOtAYyX4fRvTn87ChsLKCQBWXDhmXVH0q/tfjEcKx244l/rPJn2MsnHNh7UWJIsnvBN2oTLiJbDuWmbyqjjn4krPJmd8gkKQv8OWUBCawuG+ueX/i79FFx30bncpithFsfaHqWu7Yuk5zUAytpF0bafpNNzF+CimkCxXcNcHm/Zm75KgTeJ8ybYNJ7zE7Bxsk5l1pwnmg6m//TQr+APqEidVDT8+c2zWNZq0md2PR4hHY7Pn199NQ3CYCSIJUJ3WymA90cP/Wp6Yv1+TKCLeNTPUSj5UvpnR36DjQVmg0PtLzEYvNr1JvPDINN5XQ2Pka31RwdeCYLuwuH17atvg0rRlj619UhYbU6SBG8Vx6/s/xSZck30fSD9Lyf/PF2cuRysxbkoJqxAx9x/8+QfxDz1s+D+1aXvpp9eBjYFWoi5gyuUsDOsIP8LPBDKK4vTHX0FK1DsIPxe68jC8db45q6sm33p7lMIh1mGu9Ys18bK9hvdw2U0UulZR4QtKGS7tKRihkDg5zhNGJEyqhGGJYSOCTlRzgQHcbOQhOATGkSPXmbLlxLhjA6FKw7C0xhexP3N9MnL/Lg2GGydJJ6XiPUn44mdQw4hwS0jWJbgEDOLyHvDd91oYtu9Y67kPKnS2HFjzb2PgKeW7DTVUQeOFguj9yw6IedWnDSwqRIFqpRnQsdwYCae5xpWUQh3iJACB4nPvEUcIbcsY5VgPPjns3RilXPPTE9QBMK/QR8qcU+VIBp6jk4j+66Sb4hEpqAzFmznttNzXASl7ihMfgBS8SxTLvXWOu4RJcsIiBlfzvM9HAObVkKbFd30Pdhn61zpYHNonFuxihhCdpdsfGgN12SxQEVgohsK4S0qsOgUOB62h7GvxrAhVqbbnk+itezs1cw3XJ1zZEA63wTcUnKMRB9g/orZ0ePwOgi5nIUG2+PefCrR8Jlv+M0Jvna+EUPpvg+7M96+ve/I3fm+T1e9zRM/pEiDJBqyilbG0jCiiUF1g9RwtUR5KFwOOM1QhBgj9LSfqtCuSl3X04SbmvWyFdOjHRvkmmm2Lcu1zBCyMMqOIThnQI8iDjlphXi27J8ReRKVA7nIwhsm02ekaiLs12PYfBECERFIAsEcMMwWOA3Wi1XnJrD9et04bbEmVtwsRmlVB6PF2a8p2jRB9o8w4WPcx/ZISmFf55WTYhED4a0Np57CNDdOkssIEAEj/F1kIkbYpg2z1UlkIKPXmiKZhTlzE4bIXWMCMUOxax4vYj/BbeMmx9BE7dVu2PPIxhO0W/+CnuBJ+mlVveBy3IbWBLefIERjhL4ajCZyr2gNmVBrR1V/cdpEO8l0I+Q7EIkRLRNw2IoZjr+74wIipHnKinlK8jrGRLowjFfFdoS/UziEBhmXasZnnH6O85k+C1xmQdQmGY0i9gwSElCNt3UAGhkl9mi8ajzGZIZ5HeaaQcSv27aByVUUE7lNBik9O8g6tozXGqcviinjPNMqJrZM86wec5IKoTnGAH0jSqO6UBkDbupdPvYoeOWb/2U8Kzvr7gUSWxz5F6xkB9GVJMLuw40HCCarjwLIF4DJiNh5lzoP1XZfVVCd9mCnrrfW7OxiOkc9UZzzUcFODX2E1K8ZICSa2eLKCBmYIspvACIoQ+6swakjwx0mGlDZ2sp8lTogmBSJeQNybBvoW47SVeArpojDMLBNpLBZvKFDOL0q2MpFKHNgR1kEkZ7I4lOw8chz+bhJtFmuXY6YUB/gqzN4NYcAW5X4ojEMuOdatqeBqNAKBtg+e47gRduAAV+CK7ej/A7gbKnB1V6J+CVu5jgEngW0OT4+nwp6RCruqmnjeVWICqMo8jfRGfic+zOVGAKQyzmz0WcRLjAB0Q8hi+e9rXzFQZgBY6VoUwFHnyZJw/TGWc7N9BB7uxjnNAOFXQq3tDbvMMTcQCz8fiq8sGeATGEdAIpUoztYpnSK6wVvqmFMGxGDJLYxFt80c2H1xUrCLiyYPcizTN4WCsYddQSl2mCCg1Ru0ak3uDBCqaFrgVzXhCl5jjEdQ4Y3jXFXzTYcVKIqjKRLJHjPSeCycYhe+igDUrBF8zZWGL8bxOFpdctGTKvO6SB6kGJXA+EM5eyUphAO66ij53cP5+zukZ9zduN7v7h7Su5Vbp9wGxIJliZRh2lz+mwugnLYQCFW/WXgOsLExUTIjXe1bkkvEgzUpN0Xzjd4uS9dw+bdM6zzgOmSA0LY43PdMcmavhohuF/d/Qyw0SKcFafXu94jCGosvbDlkSi/Y0jsq52nUe7WYa9vY/Cpg9RTCp7OKez7N3JbNeIFW+U4i6OXdEQHwS3ZvNxsOFb/to/Ze0USCaJjTO8lg8/7rJSllpv8GBVAjHAvuJTbvpYmk1GeaNmfDoMEtmgmUX8Zi64Fn8QBiie0BGbnD6it9T7QgnJpRQUTaHbXbSGY7TEU2hqSYobS+7crsU03pqdBGrBA85vUf/oh1bkNby5U9GGi7VfGg+5OrLvaBDtobGkwJHFi9IK6YLIFVhjob2PT7bx30ae0DgfUsY2H0hPNBIGxqBuGgU6pGUvPtZLNVF2furCSfaf9DYotEB0KfAjUJrtjh2Zc0WNCJELcXMdC+vTWYySkPBqL8r1BKqyTRvgiCSpClfcSXj7Egr5Owbo+UC1czIoyh5r3p5fbnmS+6Tt6yOC1sdQzzrgjMmXzQbw8u8TVYYjY7Y9F0ESR5ld2vpiEHnR3+gmh0ewh6aXNT7Posd5AG39y4RtRgFsxVJOqKZJ3D3uejVsmDN79Zo1XueFTXvcle7Za/4H123G/k0vKoL87chnT1fVsMFitxpZvJrT1MEW6hP6T33ynF2hus5kofyl8W/5QHPW+C8ymgKo7cVbtgIsbsnt5YhOhADh6SCVrFW4OTrR9+DbOqU1gYm4LTnoVz2npCDI75jzvGmuU+7npYGn8mCMbgDsncKHEUqrHL45cv++ckL1Ql1BkCR7K4naH2YWz7MB6iB45fhDzbSmEsY3UtjasV+VYVepv10o3iAk2zMlIpL41pYOtO4Azr4Z7kU8AJ5Xz7azbHPbnTqwc5XB9K68rkztGihIhR+TaZAvtYiFcPuhl1ee5l/yxXywY9CBFNRlXAzvgftIAD1BPyzDwWf71zpSm7VQ0b2ZXVTdYDwO6ys4Q5lmu0XTpEXDdEJNMsIL2biR9b0fDpgDjGiJEuA6LnpB9ijfVlf2pabAxdU530QcXO8ondLSeyisPEwWriFg7Bey6aaKkZGYRlpzH1NgvzdAyCBmMFR/3gQC3Hdpw12uZWJ9qF8tJZ8QqxW5Zww5kJlcPdcUUr2myQ3bPkdMA6N8DHtEQxzxe3KUm4y7KFJxzj/CBvixhciUNJ1v2VLyAIqltVq65ujW8lnu5hS6qrKp8+QxeVuAxLCPb3uu8TJnR+zuA0o/nFEMIbnveJLYvP///8FDBV5TI4EyAOCRNrRRvSYEKKG3WlKeCJZHkPaTar/XUqvR6jko4aPsorVm/7b+5xYwePYkH5P447rzPH7HdcIOPOPLbuV+r+EpKmYNHpZl7qwxCVAW0t3iRKFGcQHcP9qUQNbi/z1n9VYw1bWTci517x5y2F1k5hAJVwqv7mdB95XzvDuE/2+MUE/4eP/bdeTQjykdoLnWkVh9cxWlczS0lLQ/vUY7zCvbJhXzHexy60Ib9y5rjOffdKq711wMTfRCZ98FEdOV2Z/rraz8F/poAL6AYTgG/tmjnnEQOg6iOd15Oi7NUCWcVKgu248RYIToyd0qc53YfViEGz3DjYRS3H3S8n84AqFQFNz0NJN0wAEQrcM9qkifCEdR/LWpatdd1s90Xk+FzFXmdmG+VjQ8dEoS9f7DD9uQJxcHO+pMb6dwt/P7OeSy+Afr/6vXjYNH34ZRdAVipgzZPhHy7HlDWIWzglw1lYADte7QG68T79GOBhWGV9UFEgKskhCjq9NJf4a2vUHZTG7YzqucxyMYtm1dZG7kT/4Mocm2Kczxz9SDT7mivE2gbGJJ+nvfqjRPEFxG0xdyYOTVAMNvA/BBIEMjZQLpcIVnGHSKQ14Jz5e7DQsh2HILrMOv++Nb71B5ATkd3OYeIO0tlyQmqwlQRxjEyPspnnbSINUEfbIdq+hmSev700g8ILy4PlOMO9LI8Dr0CtMSdNx7Qgth5hpDhv3/t3dSFeXceHfHtoQ8I5pvAT8TuXFWcxsbJdKPuWSTvswB0jMrt1zoemOhXQDaTI+g4mCiexpkxkhqxNEyAfBBB/XQqC8xi1SG+3AJd90fdiDNMlKYy41SC2cQQyBhMlKjGxr0NRxKQEsil10gYHkL5KZkH/4Suji0yEUzMd7reiol0sA1VvQ0a1wWSgJ1yCUJt/17SXKurH/+ZFoB6Yne2Nm5layY6kiC2Tmzj41igYhZ4wgrPaiK4aSvbuSlzQyzSDjyBI4gINeg2ytWjKHPCYd/iRyuVNvnMVZ8tStsp7zJMuB+YaSszDuqow/pjptYI4+BzVMh1PG1C3BPxaxDUgBvAdFiXdRtBeIa39I71hmK/HZFIZ5dMowsiMxpUS5tKrKEb9bR5J8jPNTjfxmmrmWTvDZ1KJ0fP8yybjS7CghstVzlEIcU6pn5j5KlBhdm8EpkJA9tFoY06iZUiFx2gJmjFGSZBaA6n4BSF8HpBqui6zoKJvnNzxm0raAm7qQs7iXOql5RD22chuzlMl0OMgbrgLgpvbyb7bAFK7cQz3jPTzY4h55aZEq6NSPPe8Kl0YuwMdzTcA2OJyjLXa6YdVbGF6WgfX7baBHTis9c6HpjocUYyCBLvfHqs+XD6p/u/wEBVkA43nf7sSkpv3Xw/CDtc3HAIV7We0jCfZdR+5/mZs6cg7UdG++ODX8C7Wg+HG0o/uPFeOgyU96F120HOKk2vdW2IKMp+LCR3D7c1frQH/kc8GPd0bPND4KY/g9e4jomZSF87+/309sAk2262YxSz6D/RchSFCuWJcNnueYr6kga5HezIR3GEiffyes/J9J3rv4Cgh9n+JXR586qDsdBze2zzgfTlPZ+CECsDeOnN7pM4oepAfrCg8Tygs+2RdvgYSA8tlDO6DiDuCUI2tgCz8jiKs8v8OLWzVPSfp2ib+sBtOO/ftL+eftQDpLYRphCzYbePNT2Ufn//SyT5kCFFu75543UKU7/FPQR4Ymcm8eUhYLG/uodMLkIwhCT/d5cX08nbZ2k+M8m86/c4Qk2Af/bwb6USxJZbAsV2n8eKVZxe2nOM/ABKgw5eSP/X2e+w8w5lcjXXbCII7LfaXkxPbTsaO/7p3itALPYHGkKTyi54RN8+/9P08r5n0g6C9YyxOdl0Mf3LE/83SUOZXd6l4zjqn0CNClrW0PHs+kfSb+76DFGWFQE+9j/OfY3UzPOxUKW//2CiD2mL+yBeE+2Ii0G4OEyNZWw1uoZjS/U5DpKPo8NrHnxPAhRhFuwAeDu0quhsKl4qRTcwG4rPEFtMkgZTKgtoW/NG//E+zIssilgCS5lwLTozID6IvHA6R9p8EMCcAABAAElEQVR+IZasCF+IWa8EJasY2bUMr6BK4nIFn+MvKMGhY+rhRw24rVbktf9F3EunnX1f4p4r3NPQWWPlywhqquA8M6rMOisoL0vzcNhl/gqlqOWoqJQxJFchlRPPVEY7EZdWOFceI5OC7SHAosAyT0VkkflTskA6p4He9DVkbSarkLEuwolkMQQdhHq69VVktnNuJkfgPJ2PhqCgSIU8P1+pEy1T/o2jL8bfwiDc0/UVnVlV3Nd+cJ9C6KaglCWLF7dEGoIrT3OfBW5bWs642W7aQHksBsk+fPiI+aL/i+TLFlXg3OLcMufC/kp8HKzT/Mvsg1W/H5jTG1BV7PbBGIyz1YxO4+xg6zY3cQwLTl7OzOQ9NiC2nHt6z7swaUo8Tgpe23kh/HAGVWJndZsaZIsfwBHRCtDqPHbwsXGA/iLZWDkydzeutTvZ8/gst7juvLeBEE0MDJzO0+Mczsu30fd5mdG7ZQcVQpDFhxFnrGg4ipgxj0MEr004Z7ymCAKbRoQYdUtFETQBfAhxpn62Kk3B+XTujLFDaHKMZt65c9beDPIO4YYZmcURNjlJ4FwlIg0y9iSyfxXP0NE3wzOHpqfjtdVSRB7TaTbJWE/Arceq2X2giBFwgMaJxRnHclW1RAwPbZgj3iUU0pyFrIAFOxvwhMRGoTEqTkwg4tCiGFMnw3ZZF2yMOTVOfYTzJ7iPQXDZ+LHYERtmyNQawp5vUNfEODafMRxFFbSD+dKRN0oOsmEqjjMzhgqh+RrHIzrCaP0kzjxBwDT5UiEddINKVvU4rxemSRYH4nuifJrwbJxlBOwFHSHDGY3pzGdzlo2qv6WHqQmeOU5fgNAcxkRtVRW4KE/mCi9bS83zuy986597zi89LLiwBAeWA+rOP1i7C/GmNg3jkb1CFZK5ZQgkd+QbGKY9TZ2hrbIiuc6asFoSrFZiafudFTsjSGmAzJ0OkgVEvdpGXLec4bJKIYhZJXCqYpxWlmvXAWMuZwXcSBWH4YLbCRPHOXaY/MFpMu712ZWQQ+oQOFmGNVu6pZiwAZ0ac3yW+RWyRjtRNcRqt1FBpRbP5SiE0TnVTQTiApyEcAHuNINns4KMrG1Vm9I6YAwHEMlukKxiZtBWYlV81g36cJvtPUwKXOV9le3dzcrYEZnFACmCH4dTqxJZegQHzRDxJU3U7dpeuQmCWcbM15NGcNi14NSpxxk0CrH0c98GTIlbCTBbYoF1T/eyUEeJ3d+CSLYOfWg0dQJOZeCacrjMRQjCWkyDbTiDqnA0jcJYuqc6GAN3EXYOTrL+aznjsoMAL/NuBxCB2qcphEydWHeDwP1hoZYjim0v34aDCocajqku5PNq8nx3EcwmpEkH9x3BaRW1XrHHq8fNM1/mOeyt30GSOdGT6GMjKOvbGWezp3ROXQORrZo2bqnaiAayiJ+gi1wMPeyYuxGRjegV2En9bbV1p4Ic7L2gS1PzhMU8nC5MXOF89Dx2HYbwI48H5vSGpqIihAKxncF5Zf+zoSDpqp6lnlQ7Ya55Iso4KvOeLcxYrWXIw1/Z/en0xOZDkOMKEH7nsM5cT5/c9jg4ho1gKQ6mNwzeatmW9q/H8QShld8ox5IwnF7e9SS+gXVgUg4ji76bNjW2ABu9j0krS9+/8iay4iLBYoeJWmwCKmMsfffyz9LhDbvTwxv20GKyfnovpo6h3vRs22GwKreks/03wvHTSfXqiMm2Zwzu9g2b04u0Zz1RiX3Iyr/oLg8H07Mb99Pm4vSDzuNM2Gx6dMN+ZNUmFkVfKu+j2BfY+Ed4nozh7e4z6c2bp+A6onx53+W0iRS3T1NL6zDBYcas/LzjVPphF9UHJ7HqjGLRYldu4H5HtxxKT+F8UVk/AczgawSvXcSpJRPBWAsmTQO1vHYDb0JwGwRwqu8KdbgG07FND5Owsj7dpPDFazj0zt9SGXbOeRo330KCz2d3HyPcGUQxwm9P9JeCKrcufWrHE+w4c+l9imX0jA8hdx8mOrY+dYgbdG0OlLpOeuC+yp3g2tvxO3xx7/P4ZypRmHGokdRRBxLBs9sei3CF07fr4OA4jHhWM4p1N2He3yPgbGvThvQcjsppdpMzvZcxWNxMLxLstolFfpMxHLoyTImeTgwA3fAW0I9ZgIcaDqQ/fuwrMK5ClN++9I3LrzJvlyLN0l3KRb2zaVv6/N5PkNheEx7b/jPDRIxiQKDBiuMfdTw40TOAhPLHqttM4YQDWF3qsCqMkblzkhDUDrh93rSWf1hwOR4vFrm7wyEauRWLhABNDzVOsT1NpcNYHoxfbwRq7tbwYDpMoYB9VVvZFUpSbxW5rnD4rRYZIIa8ChvwDbBstuCg2FJGTVoWRiuKUgG7wm6yiJrgZNaHPVFzkTpVOE1wY1vlb6yGGrNAZm3lfSshxxOVU4QC16ZuOLAOJYfIwLmt7DCHyBbS1b6heDTdLu9NeylJ34YH2tnfCnG3EPJgm+VetcjPUxOG27amh+Gk7goDOGQ+KLkcIoTE6gTVkGG0gwyjnRXsBnDO62SBVRNaMEsYA1sPZwDLgdiyj+cf5lkq6QM1g+HYqsDsa1SjRss6FP795TvTQ+yOWjWWKhYTxt10uLaNNps3XJ6ul15NF/guEje4szveNsbu8drtiEmV5PvWRgwNAd2pjRpfU8TYjNSMh5J7kB2siWrsQKKmM2SmdYrhGUwT0Q4iayXa8xj1wJS719dSSwwRpRwd5AhtHkcUnG8aTwPFI2kPu1U1taQKyXEVo/IAYcQHKreQXUXsUO0MetBKeoRxbuJ+VjD8WcEZzJ6D0AWDjNwvesY6Il13M18u3iqMGlvqW9MHGA0gQ5uEFFOEI29jepiK4RWVlWmEmsU/IkOtlzzZMubaWH3Hda3jPqLPsea1zuQzxRJXkYneOjliMUHAYtI4ufcfEj1TE1uOcqyoAlp3XIPKvTMEiHmlh0FNJg5HJrtbBMRhSR+3YZhMEJ2EJ5EaXY9PJT5TthV7PWRAB4QLwkzIc7zOrgtbZ2Z+7ECOK5f6Ot57Gz/g3LAk+4ZzbOcs9mmxMbMTMp3Advqdg7+AsjmN4raoHZtrCoiC1ASpoVCCzx9OnMMf/eBvjBaTq/jBKojvFuJZnEU73CFCEeWt4+H7kI9py2IJpmOe4f28xkSP2M65TgV4zjHgvjrsnCOLGS8glnofbhOWKLAXaCPQIBzQMp+BzsC9hVuJg91JCG1vZBOjnZwYC8kPGHsdSVYOdueKCUURnWdsAjmCPvmxCq/wgPOOFx/ZaNszw4Aseg7/NV3P0ScV4Uij5OZmvIl1E0eMNedrrUN0ybUw5m7egaVv+fOi2grty86xoWsfxZmnLvty9UTdf7qRgTamCAtBJ8keb906S24iaLGIATeB1Fa0yYYou1KCct61OTthS5jz3um7iO19nm5RUhM7eyflW04A4baBuOtO5NXLyHI21fKYZQzKFZDB+rn/abdKtvYhMuAv8qwhuNMySRoiJF8lwCuKIvP8BuTjcZSpGzi3SkG+snKeI3Bm+EbqGxtSx4lndrBd9pFscafNPFS7cjtOmp91n4ocAAvxGlphrM0EZlm/P9OP7sI/d60NcPpu4kYuD3SmaZRJ9Qbvd5FnabteiEA15olx8F5nbl0m+pMdhwZdJE5IpU4dTeKFftIM7T5Hf6UwUy/PEvciXr5oY+K5e4whs5+jXlQx1g0Dxc6T6SQcYmUf4oryMc/pJTFevSHP6WUA3eMD6Q3EpXrk5luESJwHFW0ATlyO9ch6AheHb4YhwZ1LP4B1okQTVocy40vnrZ6eDuTx71I+tZZxv42if2GwE7GsCvGxjgC1mXSWfo3hrGuo+IDgtXJ8C0M4vtCL6IOVGtXHLuGUu4YuUNNZldahn/ShywwA+LRCBKdL0QlbwTtsUs3Pej4I7j+AWbOLqpKF6G0ZKSM20q52nvda9wmwfUpRjmdSL+JdxB15Eos+dzJv7j2IsvwvsoVx7+cfeueCi7F3xcMFnAazu1Q0XHA+7KMOA6YE8RGaQYKRDeg0cWVrFTJe3UUSBGBj4ZhybjmVi02XeVzPavcc2ZP/nFjT+bZi3z1KANVGYjpuAtn8LtDPgxCDnYaRBAc3EvIYsM4bCFZrH79J+ttFbLtWps5abeLGFpw+R4AMbCJybxCn2DkKhzWSEXV4/T4WTAky9AUm+gpcEYxKWRkExS/aR9vhYN7M8AHhxZ9ANt+MLD+AteP4bRLex3vpNn2L5+myZzzsN3cwHkYl+zCIBIeIh3Hyz4PZOYVV5CCK2gaQ1noIAnu3DzEAW3zmeOJC5gESifFzfqJCun2mdRkDysbV1D7HyvBjVhEtts1851xym6jWx0AVuArtBlaTShxZD4NKtgcdyIV3FUddD0FjjmFtLRDbpHeeZDxEc1PfsR0ucMfbXrlQosoMHzjGBVjoAn4QzJ0NJPI/im63vor7jGfjs7eRUqqg34m/c7b7Qjo+di7mnSh+GCY3hS6y8GzH3EZTwwyx+sUtTwMeRQV6auV+r+vnONUIs2aTcg+KxmRn3/M7tzfc89mab2y4FBeJEQxWGGSwmpilEgTglzGEH74cuo3t2oQGb7JCCK2Z8QiKsQ1nScKQsSNFUIoTyLIODmjCdSQ3ez7cNpsmBtVJ4jyLMhwkk+uVnU9h6aiHUIjbJ61uBNOczoywZ8MZ91Js4TNAPW8h8vHsUEuExRozrzPEHU7Osb+lDQiQY5E51Q+xmgSxm+uMSJTTm8TQSQjwHBYHF6Y7mbProve1ThzjaHaiczwPrJ9/B4igtBKKqX5EyXOPjGtHnxmvwJDhNnpEj6Ewf3LL4yjmwAWie4zABZ/aeABIknVYyLrxYlJgYWgUWofpuOBYaIyg5Maw8kMjnAvnys8ihsmGMV/RVFrgmOs048kxXbHsYlx95Q/iEsTfgB7yKaJbH2kBHpBxPH6rIZ3pq0hf3vsicnwJC2AgzbAoe1kI0pdTF/FHjEGYZp1DGR3PyiQIYqRsrWgZOB9/Y8/z7Khi4d+KMJKXdjydjpCCqpPLcOZTp6/E3NGL2BFdkPSQd7Gk8MIWRzTu/9vemX3XdZ2HfWMiCIAkAJIAOMmcBw3UrCiWUkmOonpKLbtu07VSp2t5ZWX1uX9A3/rS1b509aGrL4kdx66TpllxZMlSZMk2ZckSNZCiSHACQWIiAGImQGIgAfb3+/Y94CUFqJRFtXWiQ17ce889Z5+9v/3tb3/z92UUBI04JY5iIf6AioXtY7huwHKbDUEcXeq4ZaQP9gbkkdwL1FBh2gEpsfyjmF06AhmKL7y7UOa5KSKp/A6LJEHPlJz36B1tOCn8D6SnERFWXbFTVqVJWuBCFT28Jkcaob7EqLKCBWU/TGNn2rrgablTfl72wRR/1S46PtcCFCuMZKyN5gKUxpkabB6OVGz1ekiaHrCIpjLzcKWGONo1PK2g1JzIMBAJeXatRqeAFed5psYu1ZauDvtiHwIxQUYHIlJYb8voLvskG7aKyLJ5jHUaf7w+jEA+HyRSI+Q8+Fyb81zRnwWQOghCCaYxRpHelw8NwNHP+Jg74mkbi6ipWET5eXUgqIisu7JwWMn3Or01mQJVz/r5s89EmzFnzpeLzTFxTcR5uwPm1mO+AV0Y1WqBic/UEGmc8BoMbhJEZdkGFpy/6cwoIXGXqrC2lcATcbhOydDUiKbhEQoqRxqAT0SHZbD6w5KHuS5v6biGBqECChNJgrhjjqVkMElEsAQulSN9Huhiw8XXeBiDEeLOGIen6OMSR+5ZNRmU5rDgmWWXx3BXBpY3OO969I2Mj4ZhpmlNPQmhCGCGv8PZMbQgPsFtf4zAdQOVFzBkTMCjT8KHB/Is9q2KcvJTwRs2EHV/kWsGqX1Vv2IUak++GuZ2Aoo9Z6ROqf/gWhzKLaaccMop1oKxCM0UOmkhMkVm3XF2H3XWLhbvDaMPCCIhYbWwFhDuyOhg2XeTrk7z6oU3Hsdha/dqXAXo8yRao/FLlJcBEQLBfLK3M7Z5nu1L2UkklW20Tf5zSL8LaPu1GHDpJF/LznDSA1sHFHcAdehO3B5mYUv70dQMT07hrEadqYZ6DIkjyBBExmmljgZETubM/76AlyJ6ORslfGSdLk6Q1Qz/ni2rgTNZMi5cmsHPaShtg12ZQ/08MIzakcV0FXdMHdEkbk62zm8htAO3BcaoepoKtPxGnAOs6iAvCaTd0KC+3FHxlVvk6VUmVzLh83SkFsz5ypbfSY/vuB+zdH1sI1VM/OLhyuTfcoeTf8sHAzL//fvopF/sOhgGGImmh0AWiVZiUDHaqpoJML/52FXSRbBYakn4KfXSy7OK69Rza1wyql6np6uFhiDamo/YWv3bVRHKV8+wSKAlUDj8IZgECwRfJrFSljcYH1RMJHSrNUuDh4tCGaYVNWE9+mwDHkZBDnnNwFInnv6LoDXsKu4+Bt1cNWoKQdPUhGZmU+iTYzcxrhW+TflneJ+LQqxyd2Buox0DM2ogkeaHtABCEHl+lyoWCyQs0OUIb2ejhfhwwx/bdXdbDSxW0id3V9PvkZw7tVWsI8RhJQuUyDLkiysoFWKnl6KDjPoV1UC51WRdg02bBpKWYqXD8TzZOl1NWqpa2RENuZwlfHCEYj7Ag4RgEoppnBgvUcTN2mZsCyA4zmw8x9jkWgRks0FoAXaeNq3cjEsDWiMipwbGEb5rIMwsDmO5Q1a5YWT5S56pJX64+ZSyhBCswini2/d+g1ItDxAgUs+5+O+PN9/yEd9t7FaPjOHy4msIJPmbMy+h0RmJReUTZUXmSPE8hDmbNw4WAS7P37rv99PDBLHIq7/DgrFe0ucx/mhxbb/QnZ4/+8vUQT0p41FdoMoTMyDNNJbhium8ZP1r0qMJrIxxOLHFeHkX+Ka9/vquJ9KjrXeFWu1Ab3tU4Hts811pJ0ah8/T1uTOv0QdqdNk7kRV3jkeRE57d/RSRV81YIIdwFPt5OjR8JJHhk6ugV6WFbVRadl+gN560EQ7fVBc+hLD59Z1PElpHAAUaqec730hvoamRBy7XzMVNt/jHR7sLmjvzGl6vLhw1RveSoeAPybywGravC57+x52vpUM8qxK2zD7qmvHA1jvTv9z9DPw6Biy0c39+/GWMXV0BYwm29pI71+5J/2of2RAYew+7yZ8eeZ60jmfZIs1nRDwCRrQH196X/uTBP4AwXIu8N6+SsWEzev1/uv2R0O491/E2hbvH0zcZezPzbVzuf33vf2Dx7YM9dLcDRiLIEsctI72uxVplN1VjDaXArSmTjUGIg8EEuyW0+CxilB+l+Ss/teznrI922uXabCwfWmjvJ1rq3f6W0KxoPvcxcYXPB5mCuAEkEwS1Yllch2V1AZZsPYLgFL7867F6Ghm0vmECH6j6VHkJo0+wC9xLG4rkodEIHlilIIhju/zios/bLAgBUGXz9HtvwDGrDcuiMaBmA16PUUe4tPKcFtR/ajCa2HGywIt8Ebx9RagGN9OvRlSHugqsBZ7BnpR4eLAuxiMUYqHQZhBMeVr/S8XpZyNu2W2E/61jpzP2tIn2nAwrjAugLPfkKeG2jz4EIEewIUFzeYYnAI5hpWuQl3ZQeNnWLJWkNyq0He0Rd6DFIwgX1wIqGwKLOljE6RVzRMNBGEtHwJhdaS2JdbcRU9zADnqFPutpqtGuEqs9zbFL4MZMSsSN9Y3hqatKehMJBTaS0FY3Fd05WqlBVQefvwE468phlcrm6rXpHP9qrtZFct+MHMXTr7/fMtJnoa0SJ6H62JLlZ2Ej8wE0g6N1IkrIWP7AD6+4DNyiG1KA4ozbnwvIewSoxLU4VqA1WEMwSeUYQq4O5fzmdQKzEurJWyBKNb9VYh0FCwOhcFQB+UBiJiaMVCIU17sVS1lsJ1SOtoCmwm0xO1vJR+b2nQwvjWd4v53WQsZEYWcL/5/wm+c5quZMhacgrvlNAx0Pz2NycKwKBS6RXYOQfnwK+l6GiSkemJObsrh4lk910cmvRzZlL3PQnpRFY8fRuONzQrFgO3Q8qyLZWWALAjDc8lFHvh440pXyQ9QPOYa+GAoaMIF1YQD0jOeykAUHJ+gf5xiblHYe5cNCWBGvt6YGTOWCGTJq4d+FlTEalfrAO9l8dy5lmULmsh3hA2s1Z74f4BhpBulLZFlmrquUpyA4GhIFiWN1fpYb7sdAemZc7BCgtCYg/CuAxvEz6aBelBFM9ttz5YADRbi2dDhZxceCstBRqVkc/K5EbgXrfbgSGFiegxj4lVul8E6vr1BxclpkzFTQveEaKUqm04GhE2kY/p5f0qExy9/Af6Ig+BzZx05hjDmL+m8adWg8lXadJG6NdkJjRKNRIABRP6glP3sYx2kZmitIrI7kAtUO3xg5ijuG6UIq0ntjlNGhEHANxrGBK6P4uhD8QWCFfLHOW94k62HVjJ8S/L2J3aiXYIyzBISo/p0PFoZnugN5eYzXcfFs7re/2jSEidFIZwi8ebX/HXYz4mxRs56hJKlCvJW34+CGKhYZjwxZwnbyAaQyptrtPGf84PQ4d0I5P610jv50YrT6256D7GarUg/GohMYBtU0xfpjMetReRwD3iurDhFwgj8MnqP9BNPEQEpP1WLbgc/Rj8+/SzvNaQA26RxZDK6wcKQHarm01ndihPz7XgxP+FcNwLYdwrBpTbFKXJIlGEeB8zgesQ1ralIzDoJGsOmjswKKH8Y8nhPAKj23/O2WkR6Qcx/Ahm8T8HEwKzoR/fj462TWeiksoErPMUPFNbwLwBuPwPK81cfvQjm3GrsyCMfiT9+685n07D54NiJ/8uE0eF1uLzQhfNYaeQ91ocyYNYS77vsUBRvCIapzDh93JnYU66I5ZMjFkOZwVz4PIlbOzKf7KIHzuaZWxMQqrJRd1JGdwyEN/xPzw+Cm207Int6f9ikWMdSlrXE9On/8c/D1MHzxBFFI/TxrJc9ifaUhEKPrcm86SVppDV5ihMmoTNQUyCESg7ynJ86lE+TlNBGp/uOrCKi4F0PUTn14oCyncfZSgN9NoE0rwSU622mwGkPAKw771MHi0QrMFsZpWBB4+4db7047zaIMEE+M9WBJ7aVNCMDN0xANuagrgkW7p3VXOJMN8qwjlLHR1TpmhbmRmJhGpbOP3KEEDw3gDj2Ja7PTJnV1QUt8rG918uApxigmyJ/jKudC5vBauzBFhFVvb2+aXkl+UmQG4yseXL+PlDGbQk46euF0Oo7n5X95/bsZLbivknag5elA55u0KjxJaYgt41wvNplq0sdgO7GGrj7/7BUB9yWHy/M/BtK7FUMOY/bphFSC0V5G1XaeIggVtddChyuQ4pJ4oouE7wVF4TcPgRGUPRBdiyLXaMgQuFxbTUTBpQpynZArcgZEvOGw8XgAlBAkNO3GQ9uJQtr1RDgpzczNpT878ZP0+KZ9INFuttoFLKIngro/jsHIaPoPMMO/SXLQ/Y3bqUu1M8zxL51+M4rFfX7j3XggNqOSG0vfOfp8GurDEFXCFnXBjyDPPIvwpLm+F6vv611HCKFrSw9hydUv6BXM4v+rQ9UewjbapDgCKFJWW4o/CGi4PpuiLlbCArLG2vR7Wx9NXyCdxiwC9UH6Z5E2a722oXU6hTl/pmM+vXehPbwwhaJNVcEjuys6k1L+bYQXfnP3kxABsjPAKhzofT999xhxpcClvMp5zAnP9p8ZyO5q2ZG+teeL1NlaTRgghS6OX6HImpFT7jhQYZD4vrad6d8ROaUmr59w0R+efCX99NxrEYSivt5hRlfUZN0058LB/Vmt0L0bdqU/ue+fRx1ZE+7+8Pir6av7Hk934pRnRZZt57ek0+9+P1XXa+l1MedD4oR1K1gc7SCPETH27bu/gr3AOrKT6T++dTkdIQRyBcJyJKwt3buIfwV+FQ3+uu8iKiBhwKV/fM/lY1iX8GxV8M1ZOCx/Avwgw+FC7oJNkT9DMVfNVibkcukU2AGh6IlljjBcwM9p/IloJZpcQJhSIIzIYN4FWSWGFNV69tLDGF4jn4ShSOc1qg9DhuAqmslX0oxUWZZO3BSxRIBohe8KzxpuFHQz28UvflcYvZWDNvT/RnKjH75Ki8QW2aZR0AUVtimppWrXjzxoT8tJuBtwof2oBKEzIB3lEvc7Lo7ocklIy+NiVmEjvKUQ5nNKRIGW+2OaFoEnbEq0MDfm3yX7GuSOUTHbbgsBfy5F9ViPctjDcdZiSJRYCpqbj0UEFkYgvqyOh5i0AtW1lTHESZUJnsu/RZdjXv3ONH7CgwcE1aaZgoopbNRiYTPz1TU6FkaYQC2fJRRzR+PJ3g8ApOj6kTvvWf8OXQhMi6uW/eOwxkkq1DXSnyZWEsFDirhxqFA3eWMs5ib1O4fzUg/Z0zaQLNQH9I+NYNAiwdACGdewKprVq5ffL6OhWndxdaSgNk3f1CVUlbAIV2OMyhpGOU2GL/n0CkoKYaDRyUm12rpJNBkIlCM4tl3BMBXjvKHXwgkIlYYe1maFW+Di5FxV7w1f2gN/PwtvbFryUajXObQY6rL7xgbCAOd4fIllhZFPRMgvnfXI8EuwuhoOy5ReoL1IFQhsCzSIhcuY4nA6QJxJDF99Y8NpavUshrGhiG5STMUQQdsgkdfAlpwa6WW8VRimMAzpuwRRiHmjndwuyAoRExdE4GL3FidszqwQUuXTZC6w0JuOf+MYDk/hY69VeY5drg95wfhYDXoMNkBZsEg6wHlU0q8x5LTeUQJZsOCatW0CGapC66BaIIdXgvUifpaA/8mRPrrAn0BQppCGrXf05W2PpXvwW6nRuBA9KC7k3VXuQcekm7riHiXdx89hDQwZzFTDwdnz5Y7Mi8pCHII3HsU3vQmd+cTIRVLk9aDLp37raB84fi11AtCLeEvqA9OC4NiHcNXNuR50up3w327dZ6i3egkjiIKViYP03dEb04Sta/EmtStTGKfeIVjmHIKqnoam67uAY9oRFtcvCdp2jxiCrTG3ZAHoD/eehoCVO0s9aj1jYudQyU3S9ovdb6W3h08ADcL6aEfo9GDYakKdaTaEM/jvGO2kX7zzaS1cd1OrCRppNotmxQwNL6Cr39LcivaEBQ8MZhG8m2GRZAk0nqn7NwhF6LrgFC6Pjp9Kw6dHw2A0eXkK2Wg0qgLWYQxy/gwnbB8/l77X/hJCM9krIDRnTNUCf78GXxldSMx7qXPaKhz21Oub2U4rcy0GOD1yXUTKKSfp0/889XNsHDj22Q6ySh+wbiFliog+jJFJteV6VJDuo1eED0ZBd5YGBGQpxwTRV2+PoqwgTYo1uUz13YMQb0lSF51yyiKFKS1wiYSL8ZMjPdBfpMyAR7XdXvjQL2x5FDNzCwglaP/Px461m1I3+eBH+436iUbzTR+63enmJdV0D0T/vn/zDmqMPhCBKJfa5tKL8Od6K+4ueQgeGWtJ/QSoPLBxV9qM4No53JOOYm3cTlDIXnh6PRx/WXko/PsfWLcLjRFuzCySN/BK3Erwwv1k7BL4v4A//ilW4U7ce1kngdhSIN2CB4nY4USmblx7w2GX45Da03uo10M4kn1p2+fDiDPIAv0JGRTepmr5IOF+UnL13I+23Zee2vIwi25VOjfWl1qGG/Eo3YDX5145rnR4kMgp0oc8RiTVet10WRgnESQttnxn8zYQHM1VQwfC9XB6mrTfpuoeYBG93P1mOkiF9/D1KfVMV2eLZAh7CZdljJ7d8YV0F7KBiHIMwfYYKbaf3YvDGQjdjxD/3nlqDqBGfnL7Q5Fw9dDo6XSBCKzfveOhyKJmCvID3YfSdlKmPLp5f7g2HMW1+DSavqe23hcCej9E469ZGGfxfLXqt2i5EqH//jW70x/u/ypuHTUQluH0xtjh1FK5Nv0T2p7Byv1y1xvpF6RzbEczF9AGsJHZmUAjeaOgwRCNOESX0jhd5J8c6UuNXX/LfJvbmavf8ptS2+UPf8udUyrnpgB6XqVFV4u7BX9mCcIVgGeoF24jzeAeJlmKbInNtwfORILXfaTbk3IYtFAL8HYZdUW6PQ1WI/i2b23eknYR2SN70wP7MMOWuovdqQU1qVmUO0e7qG6yCW/JOwJoxxEmdWwziiiMTKVuBZvBuVs6oEB6VrYiLO82xw4+4qtqJkDa1bSLOAhzPc82bzaCzURY7aHPemDKVw/jCrCZGks7mjcAI9g6/GBMhbGN79aKMoppBg1GC1Tdws1S8zE1MLhCb6eWVgPsRD3Wy3XDzcqnMaZiZkQGXTbiJPCpx8C3A23KTrRZ7sTjuFEMQgjuXrcthinlHiZqrBb5aDv+Oe6SE+y2dXiZ7UJZoOHJ+rAtsJab+b4VYV8PyikC3GXf9HpdRxq+eq5bpREPX30zvKnHr8DC24hhag+aKxdgAxkgeueHU0tqwqLOs1igbRMbUvUgoi1ssWyVR8aWYkRxask/tzhTS9677MnQMYNsomg+fJddKV6l0+VvXKJQF8aF4rby3/3M+cwb5nYUgjXuRJ0ogOOoFzBSROBD6dmyT94zyw4UPCrNyJ+6zYcQVHqWasI5kCTa5xqfEKGEsA9hiPLxPEO2RWr9UcdH/0x/FRhhKYwv8NBIFRpHhFmlO2mEC0PELmDoIrd0vYQkWCdmTmOMOm0JQbTD9coVUYSrdOIaBjpr0Ho42YZ0zhF0rd2gaLsYj+/Fy9/Cjz1IJp+dTzpZAlfATk3LlRJBEzZX0BYZMBLt2XfO6Zack13ZA3qKQUStkugdB9fkKC3PwY/TSXEg1L35injmTNn8cQtz47zbm4+GdqmJG94+BUqfEcOR84k+2bGb11Z5R3PHRcBrWBfjHqhhsGQCIYZc9Jn70Mhkg00+J9V9nQibMVJBGIw8ioB1lFI/veiv3+xqjdvPUoPK/O/q61tWN+E3MoA+uZ9aUAPpLPp5Szi+O9ZBBi4ERgTIFkq/jMLTHx85i7HnPLrw8+FRegS2xvOqRR2WBKYSmUIr5QJhilIc9cfqYOdBPosJyP7Ib1fpNhwTxHWwfO8SJTV/9ZWI0rJG1TF2lRDWmEgNS5eJxHqbMpyXiUhai2uBmY1NtXemoRdW53zw6ydgCRR2R5BVTDE+TEYAsw2YyaydiC49Fo1umsJ5zeLVTWQXMJvaceQSF9nlS2Q7oFem9JANNYVJWEFB5IGq4fS3sFztA+diUWpQ6kWW+e9HnkutlFHqQ/A/PnEWgXlFOHqJ7KcJ7L6Md2o/9ausgj7Kbqo9on8VkWpER83hFNYBm2YpJKOqmpivEVias6bqJkDdUEyF/Rpcqk8gY33ngxcjPcwg13ww2k4mf/x5LnQRpXYV1eTpSBcSRKhAj1t8/xSQ3icLynLELv+8XM+kLVKkvAgK9iUWc3ELzWQKcb09dfV1hIs1r6PGFA5MFRdReZGBtwnrXVszhXa5dOjaVOoc6k8vXhhK1RdASpDUxwwS7vcueu9wQ4UCSrRcCOqhpTbe24124djQWS4nOQVmzV3NdxBYcSepo5l4JvJY/5nwfdlHbVt1bEcGO0KY3I8u2uzOOscZYdTNwrHXIWDxbj2r872DnnG/it/clWL4noM6d6Mv77E/HGGIAxi9M4PpMAJcvs5OUiQNhMxjcKe0lep4phfp1qHs09PVTfcgJrSxCl33U9vJzkDwzTQL+BTZ00zl98DG3cBtNdFnA+lNeP53KLnzTuVRdlMWL5TVPE4vdQ7H93nKg9bAvtzTvC9tbmmhxtYV0nqPBtu0fv1GEBpfImSVI1Nn07sg6BEMhqye0GG0rV6fWpgb2bCrk7i2TK3GoHZvupuyqO7I7/YdT+8jQ/zNuZejz7olGAPxufVbU1srAeYQx44raOLwzlS7JAw/zvEpIf3H6UK+1m6L4E6KLydsqUNeOpyTSruBVHY/9Ua/SPFinckut87iC3+RcL29aT8CqAEqTfCO3VDJGShMKKQVrsV9EEaE0OYWdnrOR0yvuw4UWmqet+i89erauhc+/GmEqRZiZM+xc+hmvQvh8kHSjVyButdeW4mQN5oeb8HLEiF4AIo4hoaiRzcDJ6ikww8LQvAv8vH0J/gKOlWG+O6VkF7+CgtcdhkLjaDOs38uf9TBbjcgCmocrmIReD3jMkBdqhluvbJK/JNNM6hdTc7jbXenxzDWTeE+3UwBjf5VFHzedH/IR4Y2DlwcRYVK7Ko8M5nb3NaugnhakPHeoG2K6VHI7vON96SnN5OfHq3MalyOB8bG0jPMhUjfS7xu92BnOjx1CljbP/qD/Xs7VSi/uPW3YgfTy3KCEM+niYDaj+LB6oQm/zqKG8k8O69uwq7oJhzUnthyH6lf7gq34llUx324YBhHvLymbCkMuh16+qXb/b9z1gXCk9yezZTsa4GgkkqMXDVQMxMYgQah0rL25wKkKtJMAHoRg2kMGqFPVIUCty4WIHaOQQV54HtlrKPkpWwLE29Qu1bFeIEE+oYbqaWBRd3xCvh1sgHyfLQJCHm1XKslUQRetDu4hXAOfkbs4TMI6kKL0fB18eA3rtHRzANZPHYgdyQvzSGTLiTZKNozVCl+lMMqteczuN8F4gKuhEoGy8U4I8MvRGSF/bS/joPvWouJ/Qo4aAGNtN48UheUkAVC8GJEDoFxm65lBbC1/pXxAVUaCHk3GiyCTOhTjIA+660p7OoR1K1ZldM4IriHgiDDtR51bhjq7I9zxW7lcxpImVgLm6iccK3WsWa4CJuPc0habv/hpNAfYJLn9DY+IbM9MefxDIEyDq84jm78MukGBy7Dv5ODZQIeMcrILKCfh9c1O0HMlcjF5IdDVjgnAVgoISjF78oQJYrJNSK8bM01tB+sEax9C9Q+oi4VDnYzUKEpnmlp0JErE2gUMIzh6DaOYWqEIIgpdMuWEp2An1azkZ2gALftLh58tmHOZbzOkyiKFtPp5Y7Zf/aHb1ybX36WcvubY1LIjd/4LHvk5e50cT4GwALheuOCJ66SAhBDnpnextHwqM2yn5Y7yuV3pmyVBgqLLrIU97rws2UWSIXOnwzHpFK/hHbHzGiWCJoGLleAxSVSaU+hPw/tnZjm8PlnIizTIc4gUPu70WUXYPcMnJlChhnEViO7pmE9ciaxYOcpnXme3DjWrdIAOYl+PopxCyiBhHdRDpMETD6nACC/3Hz8f8Pe3NyxJb8ztpjImEBgyFZeVYNvDfraMwhuRnE5YSLhe8Pvp7pjpP5j0i8zmVfgYRpQg9VC1rUgyhda4a8a1Zw5eebQcsyRh8dpQTQNgMtCrLxG5BKUTPP+VM1sOjDQjl9MJ8YXvQGdgOn0Cz7/+ckXYlHNEtan1/NBvAilehYeFrliAvOsLzm08pMihtRCA5aUUEFb9eNlkCQEzfKLi8/OO+tClFxZibFKaskp0BgB8SoiICkOWfW6Vo+gNvyz919IPzjxCxYN2hxg5s1HyARXjQp1GkPU6FWsm7RQcYW/OsuJ8DcdIu5fnHw5/fWZX3I7xilUliok3uhrDwF/AVhMke4xBHTulQ2Bq0+H+6lcM9gdO8EMC3CaRXBk7BQWWutb8R1DnVo5E+Lq8q1MYRqV7733o/RXx/4elqsSAV9nOIxupGLUaGLk2cw0Ojh2BeW8cEVZBvF/s5BeoIPwQdUY2LxbHXv+/RvuI7r+C3hMboB3HkrfO/x8OkJew9EKDV0AAMA90LQ3fWv3V9NOhKG3sZ7+7Nw73Lc7/TZ8on4fz3UcSM91H2CHwKsS5KD5kAX+xY6n0pOkrQMb0o86D6SXzvyKnQRPR3dXKAwwhloSFgfye+jGK35cYlGpdgvBVSon7ZYnL33MH5b4yyXy/jXsQo8zrm/ufCptxrbwPikQ/+rkT9PJiTNBxT90Jw/VT30LgSnf2PM0dXzvjcJ0r559K1wavnb3kxic1mBFHUx/efIAwuKxiApzn6iBEDxC5Ne/vudLuDqvT6eGutN3TryAk14HvbZDDLLUdZ/rAnZF6Y69wC44GS7cjtVFh1aIIg2AJOCTGTvvyofclyrmCZQLqpBk66w6rxpzFLeOyiBoPAxyvSB7E1TcHc0FPEvefhLL0plqtoKHWu9Lf/zQs+F+fmF0IP2HN7+fuq/iJsEcSLCik8WDy95jOsq+35aPIlkhXDDG/Gw/LPEqQtqkBt4jlQtKt2RP+AUgifRS/CzAzaSNzc0Rtmes6g4QZNUqHJbY7qpLQSK6m1r+fVUDBhP6ZuT9pnWklEPY8kozI7SiaTF/e2C7GB/8Oenl1rRG3Syro5sKUD5VXyKR3UkOjQg8apQIAnkMqDA+UzbAYssijYiveBCYE5Ox5ODySdrUM7SO3cXiYvXkXqcp/NNrU/MqNCJSPv4FMMsQMZ4DZqwh1eIWDFhEt+JnXk8eS1IRkuOzlZSAmMEwCKHlwo8dKTGPwXdUqpr3Hb8OdKtxI2hdsY5nsOexQ0qfy49gE+mCQdqyPMoc+R8sEOO3d4HcyArCo/zILJJ+TNAN0v7pb3OV6jaZMous8vO4DhtZE3MN1IS1OlY1U+bw5nny+BuYd2KmIBDw++TQaWWONTx6fLrsDYB38qNjPo1Jlae2ajjr2MxunIsp8tclDxe3ASjTsAtSWa/n5gBE2bzGvSKAeON26/AMunA7lk+V/RyHxMzB9zEVtMWECSg876bRbsgiCIxZADM9zff6jD7O0iw66isAOYcC0l8uFPksqSkVsmKdv4sCkcyWe7I3KJ1wEQYmRBf540TnngfFsb8+2JV6/afi4uvv+ZYYoMYgS9qYqMnjCkByRwmksQ8xeqkrgh7AkhA4OA1EM8IerDP94hyxw/LpRirVA7OrsEhXgFeeNO6PRShsYJ2Aj922lNI0hjDVnQGw6PT1bjoUkVwkdDKc/7iOPjiEiELjXBgbOSehun7Ydy+nbW+PHzgT/Xc+xBphLKA4uC7ujgv5xMPiI21a+9aUhN4ly6pMJWu6gEBdwU67nJ7mtrM3ZjA7PXg2/aTul2nHxc/pHBtbU7Gzx0DK/oCS7HLUaoIn7xzpZsjZwbfskmU/OngpdztBBw04ZbVB4YYn9LDsB0lEewIYgJN87BmMND9feShtxKxvavH2/tMgPupN9NRSircJcJ4gYMTr1QoIfIW91/sOp4vw5I7rfbLmXja3O/u6k+ITMq5n6rLYUedGrUMgBL/xPXh65jkmN2Yt/izeEh98Noe73jST1k5wSF1PXTjJ9WBg6xkbhG1gEdCXCAFkoYuoUkn7I5Katu+VnndTH2OZ49ojfafCMW4FvkYGaxtd1YHhJ2uxfBqUFiLQjbHuF92HQ53ZCwvUieEr4OuiirAyr/3w4QLw5frIQnVm8SI+gsXjv1ggpVuFrAvUQI9KdgoFblkco72qoVrGHntuHjkiHrsEmGwPuzuBLh1pQ3dL1AwewuenG/uCMQ+GgM7DMy3X7duO9HZotnqWdB0/TwudrmZ67UovTeiHweaoGCBboSq+rFVZYqQ33WhzXqWPeSc5ME8dOccJAAgC16DOupfizlvInquKrItFMAKvbmFj6zKZ8k834vMYp66iuof7iXR5LWRA3okTmgWMteAeG+9Irw0cxrHpME8CiGz1W9Ex7yIDsFZPixGcwvPQFBbFIcKuJTPzbgJUmhCcJxH2zoC8g7jPRofzn+LyD707ptilJAQYdI5SrlRdtf49dSDt3uYd+A9tgTu5hvUYd2rSmWwj+KIR91oX7Xm0V/3ovhN2r6u0obDZxbhPYAGWRgbMUL8Ka5EvKD3Uc4rc+2dGutJKnNPUhJlORCOdu1reXZaewKDy9tkP9HEjznp3UULJhFdT+C8dId+m3qgeUmT/ridof1/THjwvybs/R155os/ugL007z96rtRNWvWTWGm1Liz9VFrhB6ujd5J2vRLPU63OYQfhDn8LKuNDlzg+BtILLl/FIS2U2rlqoXzFad5NwFSBBiG2IsiAJWw+6shWSlqAYrnqDV4QiMUWvtS9uSdBN4R1ROxEzKw8KUaTr+99GivfXtpcSAcvnITS92KMIQUIBqOjUHoT0N5DoYAHMWAZh/nCmTfJADwX12ygSvcwKrw//eC5cDGWIdesYqDLw633p6/tfgKXhzV4BU6k7x97Pr3WMwIEnAh4VZ5/D05Z39j9FM5RGykgN5b+8tTL6cK5g0HBaMQrgaQwoeOSybJj8SttqTdXHglvUr63sSi/uv3z6XfIRy86vEXOx24MSE/vfDCSofbx+djIOXJfriNqbEewKa/1HCEC7Dnihi+6L3GXlFdYZ4IkW2bJof0Y1/7Nvi+HW8Q5Uu39oP3F9Nbg+4FAsVi4pbynzk7gua3xgzhQg7bqIVKb/BFpQurR7/fgEjH5wY/S4YHjeXHRhuO5hxQg//b+b5A1rR534AvpB8dfSV/bS+QUDmZz7F4Hscj+5/e+G2xS9LUEn0XYgB9qxh4hDc23934ZSl/LPE2m//Srv0iHLx6JSDyLtEUHy2BbfPyYSF/cxrvzxR8prQYbVV+qjQKB+U0ARSf9UEywH5c64pr8A4SFy9meYk7UEX/EggGIHiEoKjSK/SIe91arReFdn3JMJagn7SVteQ+SP6YsruM713isYGuVz1T74rlqOoA2NIS0eIxsBP0KId0bOBnCNyrF6KP8sIdjEaqlbkcdWc+jXquE74zsC8JLPbRj9frSOPKH0hfHwvMlAl5i8AXApo/cWyIKrvaI3JJ3sq9cZ30WmIRoytSGpAaFCMkqgOzwbllZ4HNddvkZCpMm0VW4jXHRnkK4CyPCQoGTNbIAbcYj331EvDLyV8FCatRaCRxBg+hMJSrHFfjRlB+h5ec6nwlIAr71tE/yxDxe+mrFGp/tsBx2ccQ802vZHxUHJDlE3+DDcoSZPvz2n5bjKLu1dCa/OT2/3uGI+W/BgYdZ3fJTE5Oo8oTGJzgEMlPF9kuy/3V70Uagh13icEDBf9MJlDShD5d4WUp+irIx7VPdqZriZ6osjxBk3Ie/e9tsK0V366I8jiwPj0lV0zWpjpjckzM6Zk2luknSXs9D6XFcO0/pH111c1YxFg886AA88VnunYY96iPj7jBRWuGtKTh46cU4hGOW2XzryBPTi6vCEDrmGQRJRGX6jKSh9ZNXZFmI8ZYPME+VMoE65xDQQViR5SK7Tw/j6IOdMvDj3KV+/HMupLUX0c5Q2KIfSn8SVmYAWaQCfyQdzo5RHmiqciayk5kCxQzD1WhIFtDdzyH4SUnn2ZWHYIP6YPusv9pP+7JFCs8kJIy8oNUWaENjkgXfEsLT1QU9LdHHx74MgTpNoMvhqQ6i5qgcCPwG5m7MhiB8+maH08GLZ8JJzkSw3Vz3HipZW5lB6H6foBKxPc/vddhcRUExhzHNaK1r2BPOUi3+KGWKGubqKes5lgZoVytzTuciEVoaF285rZ+TtUAo1p7au9K/f+KPSj7eIA0D1+1zmIm1EscnxPlY4QpmEvgWEietJHFo0fVReMTvtr9AMMfraCmgmlwj9dGc3UJgdSN89jgeh0Pw7ZvqNlFZZAOoc43oqP7oX0NEHa1Aj4xVFUvgRnj+z5E22nKRp8hNr1VxxyoqrOgrTpS+EVcuwg3wm0w3wd7IAiyCelID1hB3a2jbZfNPInRmvjdPkEi1mowKBqfMwmtepnaUlUdM+60u3fpIRkqJKA6Av2VH/uZfA54NGFdIn0I/PcoCM/CjzkgliI7RQvZrK6XqDSIZZtx9jLWB7X4nfL/zcnq0G359ksRXa3JKQwxRo2QyqAOubbBBamqswao1Wx94o5NmMeaZwnAV/W8mTfg8eUSti2V+zwpkgdxnFyJED6LUQkpyK75cwBNUCr2neWuoTkd47gk8VdUYFYcsYD1szfZVd+C/1BhBMGcunY0oKTNEI5qjmKAW7jz5LCX1pUN4rCKwZktDa1xzHoIj4O6i5Ogqak9ZpLp97DgLAkEY+FcwT3nLLVq4/n7LlN5tRm1EdTSWtyURzievBFAaRW774TPjXxbAFGzUkcPAAFvIOr9foT9PbH6E+q+PB/ANdfvhyVepAbWXNHQ72QqpOUV90+dO/AykVbpXDlkgA1pr+sLGB9PDXGfq6VcwVlnc1/R8bSDQIALWq13vpD0EQDxClJPI9RLhfM+ffR232CEWtxOigMkCvWml6+OuOd7nyF7tAAn+2c6n0v1EISlA/qzrvfT8mQNohajQ6L0ZkNcJEwgkwj+28f70pa3Ub0IgNqbUNHpHKWYxjuuD1MXwv9/acHf62o4nCSJpxvV3KNyhTVm+n3Qeqjh/hYfnIMj4+3hVrgapdfAyO8R63JWfwuh2CSp9EAvq810HcAUYYJRAmR3B9ODf3Pd76aF1puqeTwcpKvH99r+LNBuBjPRZu8UXtz9O3bBHSTB7GVfo42mQdCjPkG7cwJceim1cVkBm4QVbxT26gu/FsewP9jzD/dQPhqf/u1PX0pM7HiHdH/npURG/U388fbfzuWDtlEBkzZobG9M3dzxDlNh+XBeInDr7NgLydPry1t8Oj1qrJv63D4hfRviPwgxwSLK3Sx23jPRyM2YE606jUR2ukQpzTvs8wM+IafduzxFryaZosFqVWalhtQG9aCxUsYls8uBeuxqX4vWoK/WnX0ntqro1jWkd+Wk2YFhS87EJNV4tDk4iil6a4thKjD2tTY1c0xyCrIhVBzvR1rQepG+E7akm7I7UgLrBYqVUe7Ia70QzNqi+VG8v1Vr6kHe2++yO9NE4U4u3mfLP+8xNU1mLDKERt7wJPjtUxWZ5dRFnLfdYYWUd+nYFP4VbF5MPUChsYoHeQVSSRqk5nN0G2K1MZ7iBfpuycFudgerzqZHdqpkdQxe7zQ3DpN9gF2HM9SB9M0UWVgAPjUUuZp9vSKAhgy3kEpoDLmsbZAfBJGMJgKksfxUp0pvqgA+LzLTmGvjMRtBKfxp5FjQ3tErlMFK7tgav160I+fWmIgQ+9aubiVDbGAtNQbatkbBBrgtdPkNVJpWwbmSemwmUN265qYG6uPzb2ETtL9pbCevVBJt2jmeG3LU8d6P0cGuHHVAXe3HhYvrb9lfTH+//BnwbIW7F7QqFn8ahdQs7wySOYz/qehN3XnLssO3knYe5R46pgeqH8GSKNzpZO4uLL85hVUy6virqr0W+YCTEKnjCGgShKiaZVVvCNNOQ2A4v2+G+KhaXO4ptyFOTEDde8exlEV589CG52fjIsxX0FRT9pYrdSsEvvvin1Fa+K98Ygi7npXLCPRztQNmQ9hiJ//1jX3y5DhRcTZVXY9t8N2Jqxk4DnBWOgfFIxecAmiV+vEaNQSXwqmbcNmMf8qv0SWGYC2ug9jlfPLf5k32mLSvR5HaALs/Rd6aK6LUqvVr5LRYoly8e3Jfh47jcuYGp44w2eTZzzT5OG1l3L/yybgFjFNdKeO1ABTUOahlCBUW2lF9V9APe3HmfwUe/LnXcMk8fN0MFdOhXwLK+6V6S/5jGwS06rHeLT/CRvn6dw5m4flRjdp7EMtlJ5tsBot3ntbS5viD/4BIAITHryta0i9ItUuZx4jaPT3TgQ7KBeFLiYbm2a/w8aTuIzoFXjZ7R32qoxV4q3G1p3hiurqcQnvQh2QWvuZpdbAL++CxlHtcQlL2d7diJOk10k0EdehdmkC4H1uv9F1nrgdF2qhRuZveR3eqkPz34gquFEHY3IEZgMNCDlMqC7cR/vxmKOITA1z7RmetolZp3J2nBXWA3untdB0bhu/tg4VbheLe1CV0+CN2BjUB+fQ/Z3FZDmcep4dWBsNvIuPYSBK/1umu8j1R7XRmRadt+1VIjamfTtrSlsQX5aQ6BeSh1UOPJBRt95r0a6rqV8qr7mrbjxUq9rcmzUbx5X8vu7PyHz/s7qBBnZnLCrgJaa6m7eyfldtbQZxUBJ6mvtYnSR7KBcyiOO+Dpz1GTVrVjRrsOcQAAAXtJREFUOJwBoVpwb+uabWlPyybzoadDI6dwwLuSHl63L1UzX1MUfn73/An8ichRj4v1PPJnJUmfljo+HtIv1cJv8DkR+YYXY1lkWUozlKlSHuTib5/KmFmOxXovoxfF8z/Os4t7im4W9xbnl/teXF/+fvM95b8t9/nXuWe5tm4+b9tF/2/+7Va/L70UbvXufwDXLQIQCiyeFxNWDK3clffm34prPs33xf59jIcsd8/N52/+vtQjbuWam+/7de65uY3lvt+Otv9RI30BQN9FaIXV4nMB9P8XiF48+7P3TwcC/6iR/gaQlii95wrEV3332fEPCQKZZ/0M6UtzKjgKyu8pZWVVX4tHicdf/H7bPvCMou2yx9225j9rKEMg1FxZaHJuPzuWgMDNi2CJSz479RsKgc8o/XITV2J3Flkd+P3Fo6DMiyc++/CbBIHPkP4jZqtgdwLHS8Kul4v+Nwi4H7kIVIsWd/mOlig36JfPjk8NAgC5ZPOIR5TRrP8NTlPy0Td5v6sAAAAASUVORK5CYII=">