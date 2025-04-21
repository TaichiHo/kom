package node

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// KubeletServiceTool creates a tool to manage kubelet service
func KubeletServiceTool() mcp.Tool {
	return mcp.NewTool(
		"manage_kubelet_service",
		mcp.WithDescription("管理 kubelet 服务 / Manage kubelet service"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 / The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
		mcp.WithString("action", mcp.Description("操作类型: status 或 restart / Action type: status or restart")),
	)
}

// KubeletServiceHandler handles the kubelet service management requests
func KubeletServiceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	action := request.Params.Arguments["action"].(string)

	// Validate action
	if action != "status" && action != "restart" {
		return nil, fmt.Errorf("invalid action: %s. Must be either 'status' or 'restart'", action)
	}

	klog.Infof("Managing kubelet service on node %s in cluster %s: action=%s",
		meta.Name, meta.Cluster, action)

	// Get node controller
	nodeCtl := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node()

	var result string
	var handlerErr error

	// Perform the requested action
	if action == "status" {
		result, handlerErr = nodeCtl.SystemdServiceStatus("kubelet")
	} else {
		handlerErr = nodeCtl.RestartSystemdService("kubelet")
		if handlerErr == nil {
			result = fmt.Sprintf("Successfully %sed kubelet service", action)
		}
	}

	if handlerErr != nil {
		return nil, handlerErr
	}

	return tools.TextResult(result, meta)
}

// KubeletJournalTool creates a tool to read kubelet journal logs
func KubeletJournalTool() mcp.Tool {
	return mcp.NewTool(
		"read_kubelet_journal",
		mcp.WithDescription("读取 kubelet 日志 / Read kubelet journal logs"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 / The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
		mcp.WithNumber("lines", mcp.Description("日志行数 / Number of log lines to read (max 1000)")),
	)
}

// KubeletJournalHandler handles the kubelet journal log reading request
func KubeletJournalHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	lines := 100 // Default to 100 lines
	if linesVal, ok := request.Params.Arguments["lines"].(float64); ok {
		lines = int(linesVal)
	}
	if lines <= 0 || lines > 1000 {
		lines = 100 // Default to 100 lines if not specified or invalid
	}

	klog.Infof("Reading kubelet journal logs on node %s in cluster %s: lines=%d",
		meta.Name, meta.Cluster, lines)

	// Get node controller
	nodeCtl := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node()

	// Execute SSH command to get journal logs
	result, err := nodeCtl.JournalLogs("kubelet", lines)
	if err != nil {
		return nil, fmt.Errorf("failed to get journal logs: %v", err)
	}

	return tools.TextResult(result, meta)
}
