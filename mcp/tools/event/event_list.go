package event

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	v1 "k8s.io/api/events/v1"
)

func ListEventResource() mcp.Tool {
	return mcp.NewTool(
		"list_k8s_event",
		mcp.WithDescription("List Kubernetes events by cluster and namespace"),
		mcp.WithString("cluster", mcp.Description("Cluster where the events are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("Namespace of the events (optional)")),
		mcp.WithString("involvedObjectName", mcp.Description("Filter events by involved object name")),
	)
}

func ListEventResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get resource metadata
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	// Get label selector and involved object name
	involvedObjectName, _ := request.Params.Arguments["involvedObjectName"].(string)

	// Get event list
	var list []*v1.Event
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).CRD("events.k8s.io", "v1", "Event").Namespace(meta.Namespace).RemoveManagedFields()
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	}

	if involvedObjectName != "" {
		// kubectl = kubectl.WithFieldSelector("involvedObject.name=" + involvedObjectName)
		kubectl = kubectl.WithFieldSelector("regarding.name=" + involvedObjectName)
	}
	err = kubectl.List(&list).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %v", err)
	}

	return tools.TextResult(list, meta)
}
