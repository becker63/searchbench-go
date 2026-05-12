package jcodemunch

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Runtime holds a live MCP client session for jCodeMunch-style tool servers.
type Runtime struct {
	session *mcp.ClientSession
}

// NewRuntime wraps an initialized MCP client session. The caller must have already
// completed the MCP initialize handshake (for example via [mcp.Client.Connect]).
func NewRuntime(session *mcp.ClientSession) *Runtime {
	if session == nil {
		return nil
	}
	return &Runtime{session: session}
}

// Session returns the underlying MCP session for advanced call sites.
func (r *Runtime) Session() *mcp.ClientSession {
	if r == nil {
		return nil
	}
	return r.session
}

// Close ends the MCP session.
func (r *Runtime) Close() error {
	if r == nil || r.session == nil {
		return nil
	}
	return r.session.Close()
}

// ListTools returns the full paginated tool list advertised by the server.
func (r *Runtime) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	if r == nil || r.session == nil {
		return nil, &Error{Kind: KindSetup, Op: "list tools", Err: fmt.Errorf("nil runtime")}
	}
	var (
		all    []*mcp.Tool
		cursor string
	)
	for {
		res, err := r.session.ListTools(ctx, &mcp.ListToolsParams{Cursor: cursor})
		if err != nil {
			return nil, &Error{Kind: KindSetup, Op: "mcp tools/list", Err: err}
		}
		all = append(all, res.Tools...)
		if res.NextCursor == "" {
			break
		}
		cursor = res.NextCursor
	}
	return all, nil
}

// CallTool invokes MCP tools/call and normalizes the payload for the evaluator tool surface.
func (r *Runtime) CallTool(ctx context.Context, name string, arguments json.RawMessage) (string, error) {
	if r == nil || r.session == nil {
		return "", &Error{Kind: KindToolCall, Op: "call tool", Err: fmt.Errorf("nil runtime")}
	}
	var args any
	if len(arguments) > 0 {
		if err := json.Unmarshal(arguments, &args); err != nil {
			return "", &Error{Kind: KindToolCall, Op: "parse tool arguments", Err: err}
		}
	}
	res, err := r.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return "", &Error{Kind: KindToolCall, Op: "mcp tools/call", Err: err}
	}
	return NormalizeCallToolResult(res)
}

// CommandConfig starts an MCP server subprocess over stdio and completes the MCP handshake.
type CommandConfig struct {
	// Command is the MCP server command (stdin/stdout JSON-RPC). Required.
	Command *exec.Cmd
	// RepoPath, when non-empty, is registered as an MCP root before connecting.
	RepoPath string
	// Client, when non-empty, is used as the MCP client; otherwise a default SearchBench client is created.
	Client *mcp.Client
	// Setup is invoked after a successful connect/index window for optional repo indexing tools
	// or other admin calls. Failures should wrap or return [Error] with [KindSetup] where appropriate.
	Setup func(ctx context.Context, rt *Runtime) error
}

// OpenCommand connects to a jCodeMunch MCP server launched as a subprocess.
func OpenCommand(ctx context.Context, cfg CommandConfig) (*Runtime, error) {
	if cfg.Command == nil {
		return nil, &Error{Kind: KindSetup, Op: "open command", Err: fmt.Errorf("nil command")}
	}
	client := cfg.Client
	if client == nil {
		client = mcp.NewClient(&mcp.Implementation{Name: "searchbench-go", Version: "dev"}, nil)
	}

	if cfg.RepoPath != "" {
		uri, err := filePathToRootURI(cfg.RepoPath)
		if err != nil {
			return nil, &Error{Kind: KindSetup, Op: "repo root uri", Err: err}
		}
		client.AddRoots(&mcp.Root{URI: uri})
	}

	transport := &mcp.CommandTransport{Command: cfg.Command}
	sess, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, &Error{Kind: KindSetup, Op: "mcp connect", Err: err}
	}
	rt := NewRuntime(sess)

	if cfg.Setup != nil {
		if err := cfg.Setup(ctx, rt); err != nil {
			_ = rt.Close()
			if je, ok := err.(*Error); ok && je != nil {
				return nil, je
			}
			return nil, &Error{Kind: KindSetup, Op: "adapter setup", Err: err}
		}
	}

	return rt, nil
}
