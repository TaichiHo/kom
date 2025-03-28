package kom

type ctl struct {
	kubectl *Kubectl
}

func (c *ctl) Deployment() *deploy {
	return &deploy{
		kubectl: c.kubectl,
	}
}
func (c *ctl) ReplicationController() *replicationController {
	return &replicationController{
		kubectl: c.kubectl,
	}
}
func (c *ctl) ReplicaSet() *replicaSet {
	return &replicaSet{
		kubectl: c.kubectl,
	}
}
func (c *ctl) StatefulSet() *statefulSet {
	return &statefulSet{
		kubectl: c.kubectl,
	}
}
func (c *ctl) DaemonSet() *daemonSet {
	return &daemonSet{
		kubectl: c.kubectl,
	}
}
func (c *ctl) Pod() *pod {
	return &pod{
		kubectl: c.kubectl,
	}
}
func (c *ctl) Node() *node {
	return &node{
		kubectl: c.kubectl,
	}
}
func (c *ctl) CronJob() *cronJob {
	return &cronJob{
		kubectl: c.kubectl,
	}
}

func (c *ctl) StorageClass() *storageClass {
	return &storageClass{
		kubectl: c.kubectl,
	}
}

func (c *ctl) IngressClass() *ingressClass {
	return &ingressClass{
		kubectl: c.kubectl,
	}
}
func (c *ctl) Rollout() *rollout {
	return &rollout{
		kubectl: c.kubectl,
	}
}

// Deprecated: use ctl().Scaler().Scale() instead.
func (c *ctl) Scale(replicas int32) error {
	item := &scale{
		kubectl: c.kubectl,
	}
	return item.Scale(replicas)
}
func (c *ctl) Scaler() *scale {
	return &scale{
		kubectl: c.kubectl,
	}
}

// Label updates labels
// Add label: x=y
// Delete label: x-
func (c *ctl) Label(str string) error {
	item := &label{
		kubectl: c.kubectl,
	}
	return item.Label(str)
}

// Annotate updates annotations
// Add annotation: x=y
// Delete annotation: x-
func (c *ctl) Annotate(str string) error {
	item := &annotate{
		kubectl: c.kubectl,
	}
	return item.Annotate(str)
}

func isSupportedKind(kind string, supportedKinds []string) bool {
	for _, k := range supportedKinds {
		if kind == k {
			return true
		}
	}
	return false
}
