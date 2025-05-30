package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
)

func DeleteDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"delete_k8s_resource",
		mcp.WithDescription("Delete Kubernetes resource by cluster, namespace, and name"),
		mcp.WithString("cluster", mcp.Description("Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("Name of the resource")),
		mcp.WithString("group", mcp.Description("API group of the resource")),
		mcp.WithString("version", mcp.Description("API version of the resource")),
		mcp.WithString("kind", mcp.Description("Kind of the resource")),
		mcp.WithBoolean("force", mcp.Description("Force delete the resource")),
	)
}

func DeleteDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get resource metadata
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	// Delete resource
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).CRD(meta.Group, meta.Version, meta.Kind).Namespace(meta.Namespace)
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	}
	if force, ok := request.Params.Arguments["force"].(bool); ok && force {
		err = kubectl.Name(meta.Name).ForceDelete().Error
	} else {
		err = kubectl.Name(meta.Name).Delete().Error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to delete item [%s/%s] type of [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
	}
	result := fmt.Sprintf("Successfully deleted resource [%s/%s] of type [%s%s%s]", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind)
	return tools.TextResult(result, meta)
}
