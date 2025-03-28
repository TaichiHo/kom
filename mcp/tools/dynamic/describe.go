package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
)

func GetDynamicResourceDescribe() mcp.Tool {
	return mcp.NewTool(
		"describe_k8s_resource",
		mcp.WithDescription("Retrieve Kubernetes resource details by cluster, namespace, and name"),
		mcp.WithString("cluster", mcp.Description("Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("Name of the resource")),
		mcp.WithString("group", mcp.Description("API group of the resource")),
		mcp.WithString("version", mcp.Description("API version of the resource")),
		mcp.WithString("kind", mcp.Description("Kind of the resource")),
	)
}

func GetDynamicResourceDescribeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get resource metadata
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}
	var describeResult []byte
	err = kom.Cluster(meta.Cluster).WithContext(ctx).CRD(meta.Group, meta.Version, meta.Kind).Namespace(meta.Namespace).Name(meta.Name).RemoveManagedFields().Describe(&describeResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get item [%s/%s] type of  [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
	}
	return tools.TextResult(describeResult, meta)
}
