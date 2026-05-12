package iterativecontext

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool names reserved for harness-side score setup; never exposed to the evaluator tool list.
var excludedEvaluatorToolNames = map[string]struct{}{
	"install_score": {},
	"verify_score":  {},
}

// ScoreInstallParams selects a policy module on disk and the deterministic identity used by the harness.
type ScoreInstallParams struct {
	PolicyPath string
	PolicyID   string
	// Symbol names the callable inside the policy module; empty means the server default (score_fn).
	Symbol string
}

// Runtime holds a live MCP client session for Iterative Context tool servers.
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

// listMCPTools returns the full paginated tools/list payload from the server.
func (r *Runtime) listMCPTools(ctx context.Context) ([]*mcp.Tool, error) {
	if r == nil || r.session == nil {
		return nil, &Error{Kind: KindToolSetup, Op: "list tools", Err: fmt.Errorf("nil runtime")}
	}
	var (
		all    []*mcp.Tool
		cursor string
	)
	for {
		res, err := r.session.ListTools(ctx, &mcp.ListToolsParams{Cursor: cursor})
		if err != nil {
			return nil, &Error{Kind: KindToolSetup, Op: "mcp tools/list", Err: err}
		}
		all = append(all, res.Tools...)
		if res.NextCursor == "" {
			break
		}
		cursor = res.NextCursor
	}
	return all, nil
}

// ListEvaluatorTools returns MCP tools intended for the evaluator after stripping harness-admin tools.
func (r *Runtime) ListEvaluatorTools(ctx context.Context) ([]*mcp.Tool, error) {
	all, err := r.listMCPTools(ctx)
	if err != nil {
		return nil, err
	}
	var out []*mcp.Tool
	for _, t := range all {
		if t == nil || strings.TrimSpace(t.Name) == "" {
			continue
		}
		if _, hide := excludedEvaluatorToolNames[t.Name]; hide {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

// PrepareScore installs then verifies the active score identity for this MCP session.
func PrepareScore(ctx context.Context, rt *Runtime, p ScoreInstallParams) error {
	if err := rt.InstallScore(ctx, p); err != nil {
		return err
	}
	return rt.VerifyScore(ctx, strings.TrimSpace(p.PolicyID))
}

// InstallScore invokes MCP install_score for this session.
func (r *Runtime) InstallScore(ctx context.Context, p ScoreInstallParams) error {
	if r == nil || r.session == nil {
		return &Error{Kind: KindInstall, Op: "install_score", Err: fmt.Errorf("nil runtime")}
	}
	path := strings.TrimSpace(p.PolicyPath)
	id := strings.TrimSpace(p.PolicyID)
	if path == "" || id == "" {
		return &Error{Kind: KindInstall, Op: "install_score", Err: fmt.Errorf("policy_path and policy_id are required")}
	}
	args := map[string]any{
		"policy_path": path,
		"policy_id":   id,
	}
	if sym := strings.TrimSpace(p.Symbol); sym != "" {
		args["symbol"] = sym
	}
	_, err := r.callAdminTool(ctx, KindInstall, "install_score", "install_score", args)
	return err
}

// VerifyScore invokes MCP verify_score for this session.
func (r *Runtime) VerifyScore(ctx context.Context, policyID string) error {
	if r == nil || r.session == nil {
		return &Error{Kind: KindVerify, Op: "verify_score", Err: fmt.Errorf("nil runtime")}
	}
	id := strings.TrimSpace(policyID)
	if id == "" {
		return &Error{Kind: KindVerify, Op: "verify_score", Err: fmt.Errorf("policy_id is required")}
	}
	args := map[string]any{"policy_id": id}
	_, err := r.callAdminTool(ctx, KindVerify, "verify_score", "verify_score", args)
	return err
}

func (r *Runtime) callAdminTool(ctx context.Context, kind Kind, op, name string, args map[string]any) (string, error) {
	res, err := r.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return "", &Error{Kind: kind, Op: op, Err: err}
	}
	out, normErr := NormalizeCallToolResult(res)
	if normErr != nil {
		var je *Error
		if errors.As(normErr, &je) && je != nil && je.Err != nil {
			return "", &Error{Kind: kind, Op: op, Err: je.Err}
		}
		return "", &Error{Kind: kind, Op: op, Err: normErr}
	}
	if err := interpretScoreJSON(kind, op, out); err != nil {
		return "", err
	}
	return out, nil
}

func interpretScoreJSON(kind Kind, op, payloadJSON string) error {
	var m map[string]any
	if err := json.Unmarshal([]byte(payloadJSON), &m); err != nil {
		return &Error{Kind: kind, Op: op, Err: fmt.Errorf("parse response JSON: %w", err)}
	}
	if msg, ok := m["error"].(string); ok && strings.TrimSpace(msg) != "" {
		return &Error{Kind: kind, Op: op, Err: errors.New(strings.TrimSpace(msg))}
	}
	if ok, _ := m["ok"].(bool); !ok {
		return &Error{Kind: kind, Op: op, Err: fmt.Errorf("response missing ok=true")}
	}
	return nil
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
