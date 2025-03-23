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
		mcp.WithDescription("Delete Kubernetes resource by cluster, namespace, and name / 通过集群、命名空间和名称删除Kubernetes资源"),
		mcp.WithString("cluster", mcp.Description("Cluster where the resources are running (use empty string for default cluster) / 运行资源的集群（使用空字符串表示默认集群）")),
		mcp.WithString("namespace", mcp.Description("Namespace of the resource (optional for cluster-scoped resources) / 资源所在的命名空间（集群范围资源可选）")),
		mcp.WithString("name", mcp.Description("Name of the resource / 资源的名称")),
		mcp.WithString("group", mcp.Description("API group of the resource / 资源的API组")),
		mcp.WithString("version", mcp.Description("API version of the resource / 资源的API版本")),
		mcp.WithString("kind", mcp.Description("Kind of the resource / 资源的类型")),
	)
}

func DeleteDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	// 删除资源
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).CRD(meta.Group, meta.Version, meta.Kind).Namespace(meta.Namespace)
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	}
	err = kubectl.Name(meta.Name).Delete().Error
	if err != nil {
		return nil, fmt.Errorf("failed to delete item [%s/%s] type of [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
	}
	result := fmt.Sprintf("Successfully deleted resource [%s/%s] of type [%s%s%s]", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind)
	return tools.TextResult(result, meta)

}
