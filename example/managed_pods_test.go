package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
)

func TestDeployManagedRs(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: managed-pods
  labels:
    app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  revisionHistoryLimit: 0
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
`
	kom.DefaultCluster().Applier().Apply(yaml)

	rs, err := kom.DefaultCluster().Namespace("default").
		Name("managed-pods").
		Ctl().Deployment().
		ManagedLatestReplicaSet()
	if err != nil {
		t.Logf("ManagedLatestReplicaSet error: %v", err)
	}
	t.Logf("ManagedLatestReplicaSet: %v", rs.Name)
}
func TestDeployManagedPods(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: managed-pods
  labels:
    app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  revisionHistoryLimit: 0
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
`
	kom.DefaultCluster().Applier().Apply(yaml)

	list, err := kom.DefaultCluster().Namespace("default").
		Name("managed-pods").
		Ctl().Deployment().
		ManagedPods()
	if err != nil {
		t.Logf("ManagedPods error: %v", err)
	}
	t.Logf("ManagedPods Count %d", len(list))
	for _, pod := range list {
		t.Logf("ManagedPods: %v", pod.Name)
	}
}
func TestDeployManagedPod(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: managed-pods
  labels:
    app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  revisionHistoryLimit: 0
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
`
	kom.DefaultCluster().Applier().Apply(yaml)

	item, err := kom.DefaultCluster().Namespace("default").
		Name("managed-pods").
		Ctl().Deployment().
		ManagedPod()
	if err != nil {
		t.Logf("ManagedPod error: %v", err)
	}
	t.Logf("ManagedPod: %v", item.Name)
}
