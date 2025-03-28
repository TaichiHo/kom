package kom

import (
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type tools struct {
	kubectl *Kubectl
}

func (u *tools) ClearCache() {
	u.kubectl.ClusterCache().Clear()
}

// ConvertRuntimeObjectToTypedObject is a generic conversion function that converts a runtime.Object to a specified target type
func (u *tools) ConvertRuntimeObjectToTypedObject(obj runtime.Object, target interface{}) error {
	// Assert obj as *unstructured.Unstructured type
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unable to convert object to *unstructured.Unstructured type")
	}

	// Use DefaultUnstructuredConverter to convert unstructured data to concrete type
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, target)
	if err != nil {
		return fmt.Errorf("unable to convert object to target type: %v", err)
	}

	return nil
}
func (u *tools) ConvertRuntimeObjectToUnstructuredObject(obj runtime.Object) (*unstructured.Unstructured, error) {
	// Assert obj as *unstructured.Unstructured type
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("unable to convert object to *unstructured.Unstructured type")
	}

	return unstructuredObj, nil
}
func (u *tools) GetGVRByGVK(gvk schema.GroupVersionKind) (gvr schema.GroupVersionResource, namespaced bool) {
	apiResources := u.kubectl.Status().APIResources()
	for _, resource := range apiResources {
		if resource.Kind == gvk.Kind &&
			resource.Version == gvk.Version &&
			resource.Group == gvk.Group {
			gvr = schema.GroupVersionResource{
				Group:    resource.Group,
				Version:  resource.Version,
				Resource: resource.Name, // Usually the plural form of Kind
			}
			return gvr, resource.Namespaced
		}
	}
	return schema.GroupVersionResource{}, false
}

// GetGVRByKind returns the GroupVersionResource for the corresponding string
// Gets values from k8s API interface
// If multiple versions exist simultaneously, returns the first one
// Therefore the version might not be correct
func (u *tools) GetGVRByKind(kind string) (gvr schema.GroupVersionResource, namespaced bool) {
	apiResources := u.kubectl.Status().APIResources()
	for _, resource := range apiResources {
		if resource.Kind == kind {
			version := resource.Version
			gvr = schema.GroupVersionResource{
				Group:    resource.Group,
				Version:  version,
				Resource: resource.Name, // Usually the plural form of Kind
			}
			return gvr, resource.Namespaced
		}
	}
	return schema.GroupVersionResource{}, false
}

// IsBuiltinResource checks if the given resource kind is a built-in resource.
// This function works by iterating through the apiResources list and comparing each item's Kind property with the given kind parameter.
// If a match is found, indicating the resource kind is built-in, the function returns true; otherwise, it returns false.
// This function is primarily used for quick validation of resource kinds to determine if they belong to predefined built-in types.
//
// Parameters:
//
//	kind (string): The name of the resource kind to check.
//
// Returns:
//
//	bool: Returns true if kind is one of the built-in resource kinds; otherwise returns false.
func (u *tools) IsBuiltinResource(kind string) bool {
	apiResources := u.kubectl.Status().APIResources()
	for _, list := range apiResources {
		if list.Kind == kind {
			return true
		}
	}
	return false
}

func (u *tools) GetCRD(kind string, group string) (*unstructured.Unstructured, error) {

	crdList := u.kubectl.Status().CRDList()
	for _, crd := range crdList {
		spec, found, err := unstructured.NestedMap(crd.Object, "spec")
		if err != nil || !found {
			continue
		}
		crdKind, found, err := unstructured.NestedString(spec, "names", "kind")
		if err != nil || !found {
			continue
		}
		crdGroup, found, err := unstructured.NestedString(spec, "group")
		if err != nil || !found {
			continue
		}
		if crdKind != kind || crdGroup != group {
			continue
		}
		return crd, nil
	}
	return nil, fmt.Errorf("crd %s.%s not found", kind, group)
}

// GetGVKFromObj gets the GroupVersionKind from an object
func (u *tools) GetGVKFromObj(obj interface{}) (schema.GroupVersionKind, error) {
	switch o := obj.(type) {
	case *unstructured.Unstructured:
		return o.GroupVersionKind(), nil
	case runtime.Object:
		return o.GetObjectKind().GroupVersionKind(), nil
	default:
		return schema.GroupVersionKind{}, fmt.Errorf("unsupported type %v", o)
	}
}

func (u *tools) GetGVRFromCRD(crd *unstructured.Unstructured) schema.GroupVersionResource {
	// Extract GVR
	group := crd.Object["spec"].(map[string]interface{})["group"].(string)
	version := crd.Object["spec"].(map[string]interface{})["versions"].([]interface{})[0].(map[string]interface{})["name"].(string)
	resource := crd.Object["spec"].(map[string]interface{})["names"].(map[string]interface{})["plural"].(string)

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	return gvr
}

func (u *tools) ParseGVK2GVR(gvks []schema.GroupVersionKind, versions ...string) (gvr schema.GroupVersionResource, namespaced bool) {
	// Get single GVK
	gvk := u.GetGVK(gvks, versions...)

	// Get GVR
	if u.IsBuiltinResource(gvk.Kind) {
		// Built-in resource
		return u.GetGVRByKind(gvk.Kind)
	} else {
		crd, err := u.GetCRD(gvk.Kind, gvk.Group)
		if err != nil {
			return
		}
		// Check if CRD is Namespaced
		namespaced = crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"
		gvr = u.GetGVRFromCRD(crd)
	}

	return
}

func (u *tools) GetGVK(gvks []schema.GroupVersionKind, versions ...string) (gvk schema.GroupVersionKind) {
	if len(gvks) == 0 {
		return schema.GroupVersionKind{}
	}
	if len(versions) > 0 {
		// Version specified
		v := versions[0]
		for _, g := range gvks {
			if g.Version == v {
				return schema.GroupVersionKind{
					Kind:    g.Kind,
					Group:   g.Group,
					Version: g.Version,
				}
			}
		}
	} else {
		// Take the first one
		return schema.GroupVersionKind{
			Kind:    gvks[0].Kind,
			Group:   gvks[0].Group,
			Version: gvks[0].Version,
		}
	}
	return
}

// FindGVKByTableNameInApiResources finds the corresponding GVK from the APIResource list for a table name
// APIResource includes CRD content
func (u *tools) FindGVKByTableNameInApiResources(tableName string) *schema.GroupVersionKind {

	for _, resource := range u.kubectl.parentCluster().apiResources {
		// Compare table name with resource Name or Kind
		if resource.Name == tableName || resource.Kind == tableName || resource.SingularName == tableName ||
			slice.Contain(resource.ShortNames, tableName) {
			// Build and return GroupVersionKind
			return &schema.GroupVersionKind{
				Group:   resource.Group,   // API group
				Version: resource.Version, // Version
				Kind:    resource.Kind,    // Kind
			}
		}
	}
	return nil // No matching resource found
}

// FindGVKByTableNameInCRDList finds the corresponding GVK from the CRD list for a table name
func (u *tools) FindGVKByTableNameInCRDList(tableName string) *schema.GroupVersionKind {

	for _, crd := range u.kubectl.parentCluster().crdList {
		// Get the names field under "spec" from the CRD object
		specNames, found, err := unstructured.NestedMap(crd.Object, "spec", "names")
		if err != nil || !found {
			continue // Skip current CRD if spec.names doesn't exist
		}

		// Extract kind and plural
		kind, _ := specNames["kind"].(string)
		plural, _ := specNames["plural"].(string)
		singular, _ := specNames["singular"].(string)
		shortNames, _, _ := unstructured.NestedStringSlice(crd.Object, "spec", "names", "shortNames")

		// Compare if tableName matches kind or plural
		if tableName == kind || tableName == plural || tableName == singular || slice.Contain(shortNames, tableName) {
			// Extract group and version
			group, _, _ := unstructured.NestedString(crd.Object, "spec", "group")
			versions, found, _ := unstructured.NestedSlice(crd.Object, "spec", "versions")
			if !found || len(versions) == 0 {
				continue
			}

			// Get the name field of the first version
			versionMap, ok := versions[0].(map[string]interface{})
			if !ok {
				continue
			}
			version, _ := versionMap["name"].(string)

			// Return GVK
			return &schema.GroupVersionKind{
				Group:   group,
				Version: version,
				Kind:    kind,
			}
		}
	}
	return nil // No match found
}
func (u *tools) ListAvailableTableNames() (names []string) {
	for _, resource := range u.kubectl.parentCluster().apiResources {
		// Compare table name with resource Name or Kind
		names = append(names, strings.ToLower(resource.Kind))
		for _, name := range resource.ShortNames {
			names = append(names, name)
		}
	}

	names = slice.Unique(names)
	names = slice.Filter(names, func(index int, item string) bool {
		return !strings.Contains(item, "Option")
	})
	slice.Sort(names, "asc")
	return names
}
