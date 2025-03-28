package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LabelsManager struct containing shared labels
type LabelsManager struct {
	Labels map[string]string
}

// NewLabelsManager constructor, initializes and returns a LabelsManager
func NewLabelsManager(labels map[string]string) *LabelsManager {
	return &LabelsManager{
		Labels: labels,
	}
}

// AddLabels adds shared labels to any Kubernetes resource object
func (lm *LabelsManager) AddLabels(meta *metav1.ObjectMeta) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	for k, v := range lm.Labels {
		meta.Labels[k] = v
	}
}

// AddCustomLabel dynamically adds user-specified labels
func (lm *LabelsManager) AddCustomLabel(meta *metav1.ObjectMeta, key, value string) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[key] = value
}
