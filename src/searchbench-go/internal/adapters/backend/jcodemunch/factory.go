package jcodemunch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	run "github.com/becker63/searchbench-go/internal/pure/execution"
)

// EvaluatorToolFactory returns an [run]-compatible tool factory backed by MCP tools
// exposed through rt. Tool metadata is fetched on each factory invocation so policy
// filtering in the round executor sees the same tool names as the server.
func EvaluatorToolFactory(rt *Runtime) func(run.Spec) ([]tool.BaseTool, error) {
	return func(_ run.Spec) ([]tool.BaseTool, error) {
		if rt == nil {
			return nil, &Error{Kind: KindSetup, Op: "tool factory", Err: fmt.Errorf("nil runtime")}
		}
		ctx := context.Background()
		tools, err := rt.ListTools(ctx)
		if err != nil {
			return nil, err
		}
		var out []tool.BaseTool
		for _, mt := range tools {
			if mt == nil || strings.TrimSpace(mt.Name) == "" {
				continue
			}
			info, err := ToolInfoFromMCP(mt)
			if err != nil {
				return nil, &Error{Kind: KindSetup, Op: fmt.Sprintf("tool %q", mt.Name), Err: err}
			}
			out = append(out, mcpInvokableTool{
				rt:   rt,
				name: mt.Name,
				info: info,
			})
		}
		if len(out) == 0 {
			return nil, &Error{Kind: KindSetup, Op: "tool factory", Err: fmt.Errorf("server returned no tools")}
		}
		return out, nil
	}
}

type mcpInvokableTool struct {
	rt   *Runtime
	name string
	info *schema.ToolInfo
}

func (t mcpInvokableTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return t.info, nil
}

func (t mcpInvokableTool) InvokableRun(ctx context.Context, input string, _ ...tool.Option) (string, error) {
	raw := json.RawMessage(strings.TrimSpace(input))
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	return t.rt.CallTool(ctx, t.name, raw)
}

var _ tool.InvokableTool = mcpInvokableTool{}
