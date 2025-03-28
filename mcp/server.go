package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/tools/cluster"
	"github.com/weibaohui/kom/mcp/tools/dynamic"
	"github.com/weibaohui/kom/mcp/tools/event"
	"github.com/weibaohui/kom/mcp/tools/pod"
	"k8s.io/klog/v2"
)

func RunMCPServer(name, version string, port int) {
	// Create a new MCP server
	s := server.NewMCPServer(
		name,
		version,
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithLogging(),
	)

	// Register general resource manager
	dynamic.RegisterTools(s)
	// Register Pod-related tools
	pod.RegisterTools(s)
	// Register cluster-related tools
	cluster.RegisterTools(s)
	// Register event resources
	event.RegisterTools(s)

	// Create SSE server
	sseServer := server.NewSSEServer(s)

	// Start server
	err := sseServer.Start(fmt.Sprintf(":%d", port))
	if err != nil {
		klog.Errorf("MCP Server error: %v\n", err)
	}
}
