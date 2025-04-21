package node

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// SystemdServiceStatusTool creates a tool to inspect systemd service status
func SystemdServiceStatusTool() mcp.Tool {
	return mcp.NewTool(
		"get_systemd_service_status",
		mcp.WithDescription("查询系统服务状态 / Query systemd service status"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 / The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
		mcp.WithString("service", mcp.Description("系统服务名称 / The name of the systemd service")),
	)
}

// SystemdServiceStatusHandler handles the systemd service status inspection request
func SystemdServiceStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get parameters
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	service := request.Params.Arguments["service"].(string)

	klog.Infof("Querying systemd service %s status on node %s in cluster %s", service, meta.Name, meta.Cluster)

	// Query systemd service status
	status, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().SystemdServiceStatus(service)
	if err != nil {
		return nil, err
	}

	return tools.TextResult(status, meta)
}

// RestartSystemdServiceTool creates a tool to restart systemd services
func RestartSystemdServiceTool() mcp.Tool {
	return mcp.NewTool(
		"restart_systemd_service",
		mcp.WithDescription("重启系统服务 / Restart systemd service"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 / The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
		mcp.WithString("service", mcp.Description("系统服务名称 / The name of the systemd service")),
	)
}

// RestartSystemdServiceHandler handles the systemd service restart request
func RestartSystemdServiceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get parameters
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	service := request.Params.Arguments["service"].(string)

	klog.Infof("Restarting systemd service %s on node %s in cluster %s", service, meta.Name, meta.Cluster)

	// Restart systemd service
	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().RestartSystemdService(service)
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully restarted systemd service", meta)
}
