package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ListDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"list_k8s_resource",
		mcp.WithDescription("List Kubernetes resources by cluster and resource type"),
		mcp.WithString("cluster", mcp.Description("Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("Namespace of the resources (optional for cluster-scoped resources)")),
		mcp.WithString("group", mcp.Description("API group of the resource")),
		mcp.WithString("version", mcp.Description("API version of the resource")),
		mcp.WithString("kind", mcp.Description("Kind of the resource")),
		mcp.WithString("label", mcp.Description("Label selector to filter resources (e.g. app=k8m)")),
		mcp.WithString("field", mcp.Description("Field selector to filter resources (e.g. metadata.name=test-deploy)")),
	)
}

func ListDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get resource metadata
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	// Get label selector
	label, _ := request.Params.Arguments["label"].(string)
	field, _ := request.Params.Arguments["field"].(string)

	// Get resource list
	var list []*unstructured.Unstructured
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).CRD(meta.Group, meta.Version, meta.Kind).Namespace(meta.Namespace).RemoveManagedFields()
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	}
	if label != "" {
		kubectl = kubectl.WithLabelSelector(label)
	}
	if field != "" {
		kubectl = kubectl.WithFieldSelector(field)
	}
	err = kubectl.List(&list).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list items type of [%s%s%s]: %v", meta.Group, meta.Version, meta.Kind, err)
	}

	// Extract name and namespace information
	var result []map[string]string
	for _, item := range list {
		ret := map[string]string{
			"name": item.GetName(),
		}
		if item.GetNamespace() != "" {
			ret["namespace"] = item.GetNamespace()
		}

		result = append(result, ret)
	}

	return tools.TextResult(result, meta)
}
