package pod

import (
	"context"
	"io"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
)

// GetPodLogsTool creates a tool for getting Pod logs
func GetPodLogsTool() mcp.Tool {
	return mcp.NewTool(
		"get_pod_logs",
		mcp.WithDescription("Get pod logs by cluster, namespace and name with tail lines limit"),
		mcp.WithString("cluster", mcp.Description("The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("The namespace of the pod")),
		mcp.WithString("name", mcp.Description("The name of the pod")),
		mcp.WithNumber("container", mcp.Description("Name of the container in the pod (must be specified if there are more than one container in Pod, only one container could use empty string)")),
		mcp.WithNumber("tail", mcp.Description("Number of lines from the end of the logs to show (default 100)")),
		mcp.WithBoolean("previous", mcp.Description("Whether to get logs from the previous container (default false)")),
	)
}

// GetPodLogsHandler handles requests to get Pod logs
func GetPodLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get parameters
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	tailLines := int64(100)
	if tailLinesVal, ok := request.Params.Arguments["tail"].(float64); ok {
		tailLines = int64(tailLinesVal)
	}
	containerName := ""
	if containerNameVal, ok := request.Params.Arguments["container"].(string); ok {
		containerName = containerNameVal
	}
	var stream io.ReadCloser
	opt := &v1.PodLogOptions{}
	opt.TailLines = utils.Ptr(tailLines)
	// 设置是否获取上一个容器的日志
	if previous, ok := request.Params.Arguments["previous"].(bool); ok && previous {
		opt.Previous = true
	}
	err = kom.Cluster(meta.Cluster).WithContext(ctx).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().ContainerName(containerName).GetLogs(&stream, opt).Error
	if err != nil {
		return nil, err
	}
	// Read all log content
	var logs []byte
	logs, err = io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	return tools.TextResult(string(logs), meta)
}
