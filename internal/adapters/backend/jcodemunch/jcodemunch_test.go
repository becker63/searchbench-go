package jcodemunch_test

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/becker63/searchbench-go/internal/adapters/backend/jcodemunch"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
)

func TestScriptedMCPToolSurface(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	impl := &mcp.Implementation{Name: "jcodemunch-fake-server", Version: "v0"}
	server := mcp.NewServer(impl, nil)

	objectSchema := map[string]any{"type": "object"}
	resolveSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{"type": "string", "description": "issue text"},
		},
	}
	server.AddTool(&mcp.Tool{Name: "resolve", Description: "Resolve candidate paths", InputSchema: resolveSchema},
		func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{
				StructuredContent: map[string]any{
					"paths": []string{"src/search_target.go"},
				},
			}, nil
		})
	// Admin-style tool name: still listed by MCP; round executor allow-lists gate exposure for the model.
	server.AddTool(&mcp.Tool{Name: "reindex_repo", Description: "Rebuild index", InputSchema: objectSchema},
		func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "indexed"}},
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

	rt := jcodemunch.NewRuntime(clientSession)
	factory := jcodemunch.EvaluatorToolFactory(rt)
	tools, err := factory(run.Spec{})
	if err != nil {
		t.Fatalf("tool factory: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("tools len: got %d want 2", len(tools))
	}

	out, err := rt.CallTool(ctx, "resolve", []byte(`{"query":"find widget"}`))
	if err != nil {
		t.Fatalf("call resolve: %v", err)
	}
	if out != `{"paths":["src/search_target.go"]}` {
		t.Fatalf("unexpected payload: %q", out)
	}
}

func TestCallToolSurfacesMCPResultErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	impl := &mcp.Implementation{Name: "jcodemunch-fake-server", Version: "v0"}
	server := mcp.NewServer(impl, nil)
	objectSchema := map[string]any{"type": "object"}
	server.AddTool(&mcp.Tool{Name: "broken", Description: "returns isError", InputSchema: objectSchema},
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

	rt := jcodemunch.NewRuntime(clientSession)
	_, err = rt.CallTool(ctx, "broken", []byte(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !jcodemunch.IsToolCall(err) {
		t.Fatalf("expected tool-call error, got %v", err)
	}
}

func TestOpenCommandFailsSetup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := jcodemunch.OpenCommand(ctx, jcodemunch.CommandConfig{
		Command: exec.Command("false"),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !jcodemunch.IsSetup(err) {
		t.Fatalf("expected setup error, got %v", err)
	}
}

func TestRepoRootRegisteredBeforeConnect(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	impl := &mcp.Implementation{Name: "jcodemunch-fake-server", Version: "v0"}
	server := mcp.NewServer(impl, &mcp.ServerOptions{
		RootsListChangedHandler: func(_ context.Context, req *mcp.RootsListChangedRequest) {},
	})
	objectSchema := map[string]any{"type": "object"}
	server.AddTool(&mcp.Tool{Name: "noop", Description: "noop", InputSchema: objectSchema},
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
	client.AddRoots(&mcp.Root{URI: "file:///tmp/searchbench-fake-repo"})
	clientSession, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	roots, err := serverSession.ListRoots(ctx, nil)
	if err != nil {
		t.Fatalf("list roots: %v", err)
	}
	if len(roots.Roots) != 1 || roots.Roots[0].URI != "file:///tmp/searchbench-fake-repo" {
		t.Fatalf("unexpected roots: %+v", roots.Roots)
	}
}
