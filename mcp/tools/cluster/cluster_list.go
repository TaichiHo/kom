package cluster

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
)

func ListClusters() mcp.Tool {
	return mcp.NewTool(
		"list_clusters",
		mcp.WithDescription("List all registered Kubernetes clusters"),
	)
}

func ListClustersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get all registered cluster names
	clusters := kom.Clusters().AllClusters()

	// Extract cluster names
	var result []map[string]string
	for clusterName, _ := range clusters {
		result = append(result, map[string]string{
			"name": clusterName,
		})
	}

	return tools.TextResult(result, nil)
}
