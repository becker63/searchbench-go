package iterativecontext_test

import (
	"context"
	"encoding/json"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/becker63/searchbench-go/internal/adapters/backend/iterativecontext"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
)

func objectSchema() map[string]any {
	return map[string]any{"type": "object"}
}

func TestEvaluatorToolListHidesAdminTools(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	impl := &mcp.Implementation{Name: "ic-fake-server", Version: "v0"}
	server := mcp.NewServer(impl, nil)

	server.AddTool(&mcp.Tool{Name: "resolve", Description: "Resolve symbol", InputSchema: objectSchema()},
		func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{StructuredContent: map[string]any{"ok": true}}, nil
		})
	server.AddTool(&mcp.Tool{Name: "install_score", Description: "admin", InputSchema: objectSchema()},
		func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{StructuredContent: map[string]any{"ok": true}}, nil
		})
	server.AddTool(&mcp.Tool{Name: "verify_score", Description: "admin", InputSchema: objectSchema()},
		func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{StructuredContent: map[string]any{"ok": true}}, nil
		})

	st, ct := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(impl, nil)
	clientSession, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	rt := iterativecontext.NewRuntime(clientSession)
	factory := iterativecontext.EvaluatorToolFactory(rt)
	tools, err := factory(run.Spec{})
	if err != nil {
		t.Fatalf("tool factory: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("tools len: got %d want 1 (admin tools filtered)", len(tools))
	}
	info, err := tools[0].Info(ctx)
	if err != nil {
		t.Fatalf("tool info: %v", err)
	}
	if info.Name != "resolve" {
		t.Fatalf("unexpected tool name %q", info.Name)
	}
}

func TestPrepareScoreInstallBeforeVerify(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	impl := &mcp.Implementation{Name: "ic-fake-server", Version: "v0"}
	server := mcp.NewServer(impl, nil)

	var mu sync.Mutex
	installedPolicy := ""

	osch := objectSchema()
	server.AddTool(&mcp.Tool{Name: "install_score", Description: "install", InputSchema: osch},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args map[string]any
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return &mcp.CallToolResult{StructuredContent: map[string]any{"error": err.Error()}}, nil
			}
			pid, _ := args["policy_id"].(string)
			mu.Lock()
			installedPolicy = pid
			mu.Unlock()
			return &mcp.CallToolResult{StructuredContent: map[string]any{
				"ok": true, "policy_id": pid,
			}}, nil
		})
	server.AddTool(&mcp.Tool{Name: "verify_score", Description: "verify", InputSchema: osch},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args map[string]any
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return &mcp.CallToolResult{StructuredContent: map[string]any{"error": err.Error()}}, nil
			}
			want, _ := args["policy_id"].(string)
			mu.Lock()
			got := installedPolicy
			mu.Unlock()
			if got == "" {
				return &mcp.CallToolResult{StructuredContent: map[string]any{
					"error": "no score installed for this runtime session",
				}}, nil
			}
			if got != want {
				return &mcp.CallToolResult{StructuredContent: map[string]any{
					"error": "policy mismatch",
				}}, nil
			}
			return &mcp.CallToolResult{StructuredContent: map[string]any{"ok": true, "policy_id": got}}, nil
		})

	st, ct := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(impl, nil)
	clientSession, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	rt := iterativecontext.NewRuntime(clientSession)

	if err := rt.VerifyScore(ctx, "pol-1"); err == nil {
		t.Fatal("verify before install: expected error")
	} else if !iterativecontext.IsVerify(err) {
		t.Fatalf("expected verify error kind, got %v", err)
	}

	if err := iterativecontext.PrepareScore(ctx, rt, iterativecontext.ScoreInstallParams{
		PolicyPath: "/tmp/fake_policy.py",
		PolicyID:   "pol-1",
	}); err != nil {
		t.Fatalf("PrepareScore: %v", err)
	}

	factory := iterativecontext.EvaluatorToolFactory(rt)
	_, ferr := factory(run.Spec{})
	if ferr == nil {
		t.Fatal("expected error: no evaluator tools registered on fake server")
	}
	if !iterativecontext.IsToolSetup(ferr) {
		t.Fatalf("expected tool_setup error, got %v", ferr)
	}
}

func TestInstallScoreSurfacesJSONErrorKind(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	impl := &mcp.Implementation{Name: "ic-fake-server", Version: "v0"}
	server := mcp.NewServer(impl, nil)

	server.AddTool(&mcp.Tool{Name: "install_score", Description: "install", InputSchema: objectSchema()},
		func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{StructuredContent: map[string]any{"error": "bad module"}}, nil
		})

	st, ct := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(impl, nil)
	clientSession, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	rt := iterativecontext.NewRuntime(clientSession)
	err = rt.InstallScore(ctx, iterativecontext.ScoreInstallParams{PolicyPath: "/x", PolicyID: "p"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !iterativecontext.IsInstall(err) {
		t.Fatalf("expected install error kind, got %v", err)
	}
}

func TestCallToolEvaluatorSurfacesToolKind(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	impl := &mcp.Implementation{Name: "ic-fake-server", Version: "v0"}
	server := mcp.NewServer(impl, nil)

	server.AddTool(&mcp.Tool{Name: "resolve", Description: "Resolve", InputSchema: objectSchema()},
		func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "boom"}},
			}, nil
		})

	st, ct := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(impl, nil)
	clientSession, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	rt := iterativecontext.NewRuntime(clientSession)
	_, err = rt.CallTool(ctx, "resolve", []byte(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !iterativecontext.IsToolCall(err) {
		t.Fatalf("expected tool-call error, got %v", err)
	}
}

func TestOpenCommandFailsSession(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := iterativecontext.OpenCommand(ctx, iterativecontext.CommandConfig{
		Command: exec.Command("false"),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !iterativecontext.IsSession(err) {
		t.Fatalf("expected session error, got %v", err)
	}
}
