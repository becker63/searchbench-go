package iterativecontext

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NormalizeCallToolResult turns an MCP tools/call result into a single JSON string for Eino.
//
// When the result is not an MCP error payload but [mcp.CallToolResult.IsError] is set,
// this returns an [Error] with [KindToolCall].
func NormalizeCallToolResult(res *mcp.CallToolResult) (string, error) {
	if res == nil {
		return "", fmt.Errorf("nil CallToolResult")
	}
	if res.IsError {
		return "", &Error{
			Kind: KindToolCall,
			Op:   "tool returned error",
			Err:  fmt.Errorf("%s", textFromContent(res.Content)),
		}
	}
	if res.StructuredContent != nil {
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			return "", &Error{Kind: KindToolCall, Op: "marshal structuredContent", Err: err}
		}
		return string(raw), nil
	}
	if out := textFromContent(res.Content); out != "" {
		return out, nil
	}
	if len(res.Content) > 0 {
		raw, err := json.Marshal(res.Content)
		if err != nil {
			return "", &Error{Kind: KindToolCall, Op: "marshal content", Err: err}
		}
		return string(raw), nil
	}
	return "{}", nil
}

func textFromContent(parts []mcp.Content) string {
	var b strings.Builder
	for _, part := range parts {
		if part == nil {
			continue
		}
		switch t := part.(type) {
		case *mcp.TextContent:
			if b.Len() > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(t.Text)
		}
	}
	return b.String()
}
