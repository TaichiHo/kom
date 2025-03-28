package tools

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	"k8s.io/apimachinery/pkg/util/json"
)

// buildTextResult builds a standard text response result
func buildTextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}
}

// TextResult converts any type to a standard mcp.CallToolResult
func TextResult[T any](item T, meta *metadata.ResourceMetadata) (*mcp.CallToolResult, error) {
	switch v := any(item).(type) {
	case []byte:
		return buildTextResult(string(v)), nil
	case []string:
		var contents []mcp.Content
		for _, s := range v {
			contents = append(contents, mcp.TextContent{
				Type: "text",
				Text: s,
			})
		}
		return &mcp.CallToolResult{Content: contents}, nil
	default:
		bytes, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("failed to json marshal item [%s/%s] type of [%s%s%s]: %v",
				meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
		}
		return buildTextResult(string(bytes)), nil
	}
}

func ErrorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: err.Error(),
			},
		},
	}
}
