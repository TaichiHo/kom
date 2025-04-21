package node

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {
	s.AddTool(
		TaintNodeTool(),
		TaintNodeHandler,
	)
	s.AddTool(
		UnTaintNodeTool(),
		UnTaintNodeHandler,
	)
	s.AddTool(
		CordonNodeTool(),
		CordonNodeHandler,
	)
	s.AddTool(
		UnCordonNodeTool(),
		UnCordonNodeHandler,
	)

	s.AddTool(
		DrainNodeTool(),
		DrainNodeHandler,
	)
	s.AddTool(
		NodeResourceUsageTool(),
		NodeResourceUsageHandler,
	)

	s.AddTool(
		NodeIPUsageTool(),
		NodeIPUsageHandler,
	)
	s.AddTool(
		NodePodCountTool(),
		NodePodCountHandler,
	)

	s.AddTool(
		SystemdServiceStatusTool(),
		SystemdServiceStatusHandler,
	)
	s.AddTool(
		RestartSystemdServiceTool(),
		RestartSystemdServiceHandler,
	)

	s.AddTool(
		KubeletServiceTool(),
		KubeletServiceHandler,
	)
	s.AddTool(
		KubeletJournalTool(),
		KubeletJournalHandler,
	)
	s.AddTool(
		ContainerdServiceTool(),
		ContainerdServiceHandler,
	)
	s.AddTool(
		ContainerdJournalTool(),
		ContainerdJournalHandler,
	)
}
