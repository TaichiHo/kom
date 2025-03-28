package utils

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func SortByCreationTime(items []unstructured.Unstructured) []unstructured.Unstructured {
	sort.Slice(items, func(i, j int) bool {
		ti := items[i].GetCreationTimestamp()
		tj := items[j].GetCreationTimestamp()
		return ti.After(tj.Time)
	})
	return items
}

// RemoveManagedFields removes the metadata.managedFields field from the unstructured.Unstructured object
func RemoveManagedFields(obj *unstructured.Unstructured) {
	// Get metadata
	metadata, found, err := unstructured.NestedMap(obj.Object, "metadata")
	if err != nil || !found {
		return
	}

	// Delete managedFields
	delete(metadata, "managedFields")

	// Update metadata
	err = unstructured.SetNestedMap(obj.Object, metadata, "metadata")
	if err != nil {
		return
	}
}

// ConvertUnstructuredToYAML converts an Unstructured object to a YAML string
func ConvertUnstructuredToYAML(obj *unstructured.Unstructured) (string, error) {

	// Marshal Unstructured object to JSON
	jsonBytes, err := obj.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to serialize Unstructured object to JSON: %v", err)
	}

	// Convert JSON to YAML
	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return "", fmt.Errorf("failed to convert JSON to YAML: %v", err)
	}

	return string(yamlBytes), nil
}

// AddOrUpdateAnnotations adds or updates annotations
func AddOrUpdateAnnotations(item *unstructured.Unstructured, newAnnotations map[string]string) {
	// Get existing annotations
	annotations := item.GetAnnotations()
	if annotations == nil {
		// If not exists, initialize a map
		annotations = make(map[string]string)
	}

	// Append or override new data
	for key, value := range newAnnotations {
		annotations[key] = value
	}

	// Set back to the object
	item.SetAnnotations(annotations)
}
