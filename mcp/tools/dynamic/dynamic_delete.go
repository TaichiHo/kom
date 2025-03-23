package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools/metadata"
)

func DeleteDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"delete_k8s_resource",
		mcp.WithDescription("Delete Kubernetes resource by cluster, namespace, and name"),
		mcp.WithString("cluster", mcp.Description("Cluster where the resource is running")),
		mcp.WithString("namespace", mcp.Description("Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("Name of the resource")),
		mcp.WithString("group", mcp.Description("API group of the resource (optional if resourceType is provided)")),
		mcp.WithString("version", mcp.Description("API version of the resource (optional if resourceType is provided)")),
		mcp.WithString("kind", mcp.Description("Kind of the resource (optional if resourceType is provided)")),
	)
}

func DeleteDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	// 删除资源
	err = kom.Cluster(meta.Cluster).WithContext(ctx).CRD(meta.Group, meta.Version, meta.Kind).Namespace(meta.Namespace).Name(meta.Name).Delete().Error
	if err != nil {
		return nil, fmt.Errorf("failed to delete item [%s/%s] type of [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
	}

	// 构建返回结果
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully deleted resource [%s/%s] of type [%s%s%s]", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind),
			},
		},
	}, nil
}
