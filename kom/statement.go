package kom

import (
	"context"
	"io"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

type Statement struct {
	*Kubectl            `json:"Kubectl,omitempty"`  // Base configuration
	RowsAffected        int64                       `json:"rowsAffected,omitempty"`        // Number of affected rows
	TotalCount          *int64                      `json:"totalCount,omitempty"`          // Total count for queries, used for pagination. Only effective in List query methods.
	AllNamespace        bool                        `json:"allNamespace,omitempty"`        // All namespaces
	Namespace           string                      `json:"namespace,omitempty"`           // Resource namespace
	NamespaceList       []string                    `json:"namespace_list,omitempty"`      // Multiple namespaces, used for list queries only. Cross-namespace queries only occur during list operations. When using AllNamespace, NamespaceList is not needed
	Name                string                      `json:"name,omitempty"`                // Resource name
	GVR                 schema.GroupVersionResource `json:"GVR"`                           // Resource type (GroupVersionResource)
	GVK                 schema.GroupVersionKind     `json:"GVK"`                           // Resource type (GroupVersionKind)
	Namespaced          bool                        `json:"namespaced,omitempty"`          // Whether it's a namespaced resource
	ListOptions         []metav1.ListOptions        `json:"listOptions,omitempty"`         // List query parameters, used as variadic args. By default, only the first one is used
	Context             context.Context             `json:"-"`                             // Context
	Dest                interface{}                 `json:"dest,omitempty"`                // Destination object for results, typically a struct pointer
	PatchType           types.PatchType             `json:"patchType,omitempty"`           // PATCH type
	PatchData           string                      `json:"patchData,omitempty"`           // PATCH data
	RemoveManagedFields bool                        `json:"removeManagedFields,omitempty"` // Whether to remove managed fields
	useCustomGVK        bool                        `json:"-"`                             // If GVK is set via CRD method, force its use and skip automatic GVK resolution
	ContainerName       string                      `json:"containerName,omitempty"`       // Container name, used for container log operations
	Command             string                      `json:"command,omitempty"`             // Container commands, including ls, cat, and user input commands
	Args                []string                    `json:"args,omitempty"`                // Container command arguments
	PodLogOptions       *v1.PodLogOptions           `json:"-" `                            // Used for getting container logs
	Stdin               io.Reader                   `json:"-" `                            // Set input
	Filter              Filter                      `json:"filter,omitempty"`
	StdoutCallback      func(data []byte) error     `json:"-"`
	StderrCallback      func(data []byte) error     `json:"-"`
	CacheTTL            time.Duration               `json:"cacheTTL,omitempty"`    // Cache duration
	ForceDelete         bool                        `json:"forceDelete,omitempty"` // Force delete flag
}

type Filter struct {
	Columns    []string    `json:"columns,omitempty"`
	Conditions []Condition `json:"condition,omitempty"` // xx=?
	Order      string      `json:"order,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	Offset     int         `json:"offset,omitempty"`
	Sql        string      `json:"sql,omitempty"`    // Original SQL
	Parsed     bool        `json:"parsed,omitempty"` // Whether it has been parsed
	From       string      `json:"from,omitempty"`   // From TableName
}

type Condition struct {
	Depth     int
	AndOr     string
	Field     string
	Operator  string
	Value     interface{} // Set to precise type value through detectType, before detectType it's always string
	ValueType string      // number, string, bool, time
}

func (s *Statement) ParseGVKs(gvks []schema.GroupVersionKind, versions ...string) *Statement {

	s.GVR = schema.GroupVersionResource{}
	s.GVK = schema.GroupVersionKind{}
	// Get single GVK
	gvk := s.Tools().GetGVK(gvks, versions...)
	s.GVK = gvk

	// Get GVR
	if s.Tools().IsBuiltinResource(gvk.Kind) {
		// Built-in resource
		if s.useCustomGVK {
			// CRD is set with version
			s.GVR, s.Namespaced = s.Tools().GetGVRByGVK(gvk)
		} else {
			s.GVR, s.Namespaced = s.Tools().GetGVRByKind(gvk.Kind)
		}
		klog.V(6).Infof("useCustomGVK=%v \t GVR=%v \t GVK=%v", s.useCustomGVK, s.GVR, s.GVK)
	} else {
		crd, err := s.Tools().GetCRD(gvk.Kind, gvk.Group)
		if err != nil {
			return s
		}
		// Check if CRD is Namespaced
		s.Namespaced = crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"
		s.GVR = s.Tools().GetGVRFromCRD(crd)
	}

	return s
}

func (s *Statement) ParseNsNameFromRuntimeObj(obj runtime.Object) *Statement {
	// Get metadata (like Name and Namespace)
	accessor, err := meta.Accessor(obj)
	if err != nil {
		klog.V(6).Infof("error getting meta data by meta.Accessor : %v", err)
		return s
	}
	if name := accessor.GetName(); name != "" {
		s.Name = name // Get resource name
	}
	if namespace := accessor.GetNamespace(); namespace != "" {
		s.Namespace = namespace // Get resource namespace
	}
	return s
}

func (s *Statement) ParseGVKFromRuntimeObj(obj runtime.Object) *Statement {
	// Use scheme.Scheme.ObjectKinds() to get Kind
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		klog.V(6).Infof("error getting kind by scheme.Scheme.ObjectKinds : %v", err)
		return s
	}
	s.ParseGVKs(gvks)
	return s
}

func (s *Statement) ParseFromRuntimeObj(obj runtime.Object) *Statement {
	return s.
		ParseGVKFromRuntimeObj(obj).
		ParseNsNameFromRuntimeObj(obj)
}
