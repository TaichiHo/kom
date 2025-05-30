package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/tools/cluster"
	"github.com/weibaohui/kom/mcp/tools/deployment"
	"github.com/weibaohui/kom/mcp/tools/dynamic"
	"github.com/weibaohui/kom/mcp/tools/event"
	"github.com/weibaohui/kom/mcp/tools/ingressclass"
	"github.com/weibaohui/kom/mcp/tools/node"
	"github.com/weibaohui/kom/mcp/tools/pod"
	"github.com/weibaohui/kom/mcp/tools/storageclass"
	"github.com/weibaohui/kom/mcp/tools/yaml"
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

	// register tools
	dynamic.RegisterTools(s)
	pod.RegisterTools(s)
	cluster.RegisterTools(s)
	event.RegisterTools(s)
	deployment.RegisterTools(s)
	node.RegisterTools(s)
	storageclass.RegisterTools(s)
	ingressclass.RegisterTools(s)
	yaml.RegisterTools(s)

	// Create SSE server
	sseServer := server.NewSSEServer(s)

	// Start server
	err := sseServer.Start(fmt.Sprintf(":%d", port))
	if err != nil {
		klog.Errorf("MCP Server error: %v\n", err)
	}
}
