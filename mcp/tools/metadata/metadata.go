package metadata

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// ResourceMetadata encapsulates resource metadata information
type ResourceMetadata struct {
	Cluster   string
	Namespace string
	Name      string
	Group     string
	Version   string
	Kind      string
}

// ParseFromRequest parses resource metadata from the request
func ParseFromRequest(request mcp.CallToolRequest) (*ResourceMetadata, error) {
	// Validate required parameters
	// Get cluster parameter, use empty string as default if not present
	cluster := ""
	if clusterVal, ok := request.Params.Arguments["cluster"].(string); ok {
		cluster = clusterVal
	}

	// Get name parameter, return error if not present
	name := ""
	if nameVal, ok := request.Params.Arguments["name"].(string); ok {
		name = nameVal
	}

	// Get namespace parameter (optional, supports cluster-level resources)
	namespace := ""
	if ns, ok := request.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	// Get resource type information
	var group, version, kind string
	if resourceType, ok := request.Params.Arguments["kind"].(string); ok && resourceType != "" {
		// If resourceType is provided, get resource info from type.go
		if info, exists := GetResourceInfo(resourceType); exists {
			// Use user-specified GVK if provided, otherwise use default values
			group = getStringParam(request, "group", info.Group)
			version = getStringParam(request, "version", info.Version)
			kind = getStringParam(request, "kind", info.Kind)
		}
	}

	// If no information was obtained through resourceType, use directly specified GVK
	if group == "" {
		group = getStringParam(request, "group", "")
	}
	if version == "" {
		version = getStringParam(request, "version", "")
	}
	if kind == "" {
		kind = getStringParam(request, "kind", "")
	}

	return &ResourceMetadata{
		Cluster:   cluster,
		Namespace: namespace,
		Name:      name,
		Group:     group,
		Version:   version,
		Kind:      kind,
	}, nil
}

// getStringParam gets a string value from request parameters, returns default value if not present or invalid
func getStringParam(request mcp.CallToolRequest, key, defaultValue string) string {
	if value, ok := request.Params.Arguments[key].(string); ok && value != "" {
		return value
	}
	return defaultValue
}
