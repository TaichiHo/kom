package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/tools/pod"
	"k8s.io/klog/v2"
)

func RunMCPServer(port int) {
	// 创建一个新的 MCP 服务器
	s := server.NewMCPServer(
		"kom mcp server",
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	// 注册Pod相关工具
	pod.RegisterPodTools(s)

	// 创建 SSE 服务器
	sseServer := server.NewSSEServer(s)

	// 启动服务器
	err := sseServer.Start(fmt.Sprintf(":%d", port))
	if err != nil {
		klog.Errorf("MCP Server error: %v\n", err)
	}
}
