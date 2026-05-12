package round

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	iterativecontext "github.com/becker63/searchbench-go/internal/adapters/backend/iterativecontext"
	evaluatoreino "github.com/becker63/searchbench-go/internal/agents/evaluator/eino"
	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
)

func TestDefaultPreparedTools_selectsFakeBackend(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	spec := testEvaluatorSpec(t, domain.BackendFake, nil)
	tools, cleanup, err := defaultPreparedToolFactory()(ctx, spec)
	if cleanup != nil {
		t.Fatal("expected nil cleanup for fake backend")
	}
	if err != nil {
		t.Fatalf("tool factory: %v", err)
	}
	if len(tools) != len(evaluatorfake.LocalEvaluatorToolNames()) {
		t.Fatalf("tools len %d, want %d", len(tools), len(evaluatorfake.LocalEvaluatorToolNames()))
	}
}

func TestDefaultPreparedTools_jCodeMunchRequiresEnv(t *testing.T) {
	t.Setenv(envJCodeMunchCommand, "")
	ctx := context.Background()
	spec := testEvaluatorSpec(t, domain.BackendJCodeMunch, nil)
	_, _, err := defaultPreparedToolFactory()(ctx, spec)
	if err == nil {
		t.Fatal("expected error when jCodeMunch command env is unset")
	}
}

func TestDefaultPreparedTools_iterativeContextRequiresPolicy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	spec := testEvaluatorSpec(t, domain.BackendIterativeContext, nil)
	_, _, err := defaultPreparedToolFactory()(ctx, spec)
	if err == nil {
		t.Fatal("expected error without policy")
	}
}

func TestDefaultPreparedTools_iterativeContextRequiresEnv(t *testing.T) {
	t.Setenv(envIterativeContextCommand, "")
	ctx := context.Background()
	pol := domain.NewPythonPolicy(domain.PolicyID("p1"), "def score_fn(x):\n    return []\n", "score_fn")
	spec := testEvaluatorSpec(t, domain.BackendIterativeContext, &pol)
	_, cleanup, err := defaultPreparedToolFactory()(ctx, spec)
	if cleanup != nil {
		cleanup()
		t.Fatal("unexpected cleanup")
	}
	if err == nil {
		t.Fatal("expected error when Iterative Context command env is unset")
	}
}

func TestEvaluatorExecutor_fakeBackendUsesDefaultPreparedTools(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	spec := testEvaluatorSpec(t, domain.BackendFake, nil)

	allowed := map[string]struct{}{}
	for _, name := range evaluatorfake.LocalEvaluatorDefaultAllowedToolNames() {
		allowed[name] = struct{}{}
	}

	ex := &evaluatorExecutor{
		modelFactory:      evaluatorfake.ModelFactory,
		toolFactory:       defaultPreparedToolFactory(),
		allowedTools:      allowed,
		retryPolicy:       evaluatoreino.RetryPolicy{MaxAttempts: 1},
		evaluatorAppendix: run.EvaluatorRunAppendix{},
	}
	executed, err := ex.Execute(ctx, spec)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(executed.Prediction.Files) == 0 {
		t.Fatalf("expected prediction files")
	}

	recs := ex.executions()
	if len(recs) != 1 {
		t.Fatalf("execution records len %d", len(recs))
	}
	if len(recs[0].Result.Phases) == 0 {
		t.Fatal("expected evaluator phases on success path")
	}
}

func TestEvaluatorExecutor_toolFactoryErrorSkipsEvaluatorRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	spec := testEvaluatorSpec(t, domain.BackendFake, nil)
	allowed := map[string]struct{}{}
	for _, name := range evaluatorfake.LocalEvaluatorDefaultAllowedToolNames() {
		allowed[name] = struct{}{}
	}
	ex := &evaluatorExecutor{
		modelFactory: evaluatorfake.ModelFactory,
		toolFactory: func(context.Context, run.Spec) ([]tool.BaseTool, func(), error) {
			return nil, nil, fmt.Errorf("injected tool factory failure")
		},
		allowedTools: allowed,
		retryPolicy:  evaluatoreino.RetryPolicy{MaxAttempts: 1},
	}
	_, err := ex.Execute(ctx, spec)
	if err == nil {
		t.Fatal("expected error")
	}
	if len(ex.executions()) != 0 {
		t.Fatal("expected no recorded executions when tool preparation fails")
	}
}

func TestIterativeContextEvaluatorToolSurfaceOmitsAdminTools(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	impl := &mcp.Implementation{Name: "ic-round-test", Version: "v0"}
	server := mcp.NewServer(impl, nil)
	obj := map[string]any{"type": "object"}
	server.AddTool(&mcp.Tool{Name: "resolve", Description: "r", InputSchema: obj},
		func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{StructuredContent: map[string]any{"ok": true}}, nil
		})
	server.AddTool(&mcp.Tool{Name: "install_score", Description: "a", InputSchema: obj},
		func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{StructuredContent: map[string]any{"ok": true}}, nil
		})

	st, ct := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })

	client := mcp.NewClient(impl, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	rt := iterativecontext.NewRuntime(cs)
	factory := iterativecontext.EvaluatorToolFactory(rt)
	tools, err := factory(run.Spec{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tl := range tools {
		info, ierr := tl.Info(ctx)
		if ierr != nil {
			t.Fatal(ierr)
		}
		switch info.Name {
		case "install_score", "verify_score":
			t.Fatalf("admin tool %q leaked to evaluator-visible factory", info.Name)
		}
	}
}

func TestLegacyToolFactoryWrapperPassesThrough(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	spec := testEvaluatorSpec(t, domain.BackendFake, nil)
	called := false
	wrapped := wrapLegacyEvaluatorToolFactory(func(s run.Spec) ([]tool.BaseTool, error) {
		called = true
		if s.ID != spec.ID {
			t.Fatalf("spec id mismatch")
		}
		return evaluatorfake.ToolFactory(s)
	})
	tools, cleanup, err := wrapped(ctx, spec)
	if cleanup != nil {
		t.Fatal("expected nil cleanup")
	}
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected legacy factory invoked")
	}
	if len(tools) == 0 {
		t.Fatal("expected tools")
	}
}

func testEvaluatorSpec(t *testing.T, backend domain.BackendKind, policy *domain.PolicyArtifact) run.Spec {
	t.Helper()
	dir := t.TempDir()
	sys := domain.SystemSpec{
		ID:           domain.SystemID("sys-1"),
		Name:         "test-system",
		Backend:      backend,
		Model:        domain.ModelSpec{Provider: "fake", Name: "fake"},
		PromptBundle: domain.PromptBundleRef{Name: "default"},
		Policy:       policy,
	}
	if err := sys.Validate(); err != nil {
		t.Fatalf("system validate: %v", err)
	}
	match := domain.MatchSpec{
		ID:        domain.MatchID("match-1"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: "demo/repo",
			Path: domain.HostPath(dir),
		},
		Input: domain.MatchInput{Title: "locate bug"},
		Oracle: domain.MatchOracle{
			GoldFiles: []domain.RepoRelPath{"src/search_target.go"},
		},
	}
	return run.NewSpec(domain.RunID("incumbent-match-1-sys"), match, sys)
}
