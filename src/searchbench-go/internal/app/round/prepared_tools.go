package round

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudwego/eino/components/tool"

	iterativecontext "github.com/becker63/searchbench-go/internal/adapters/backend/iterativecontext"
	jcodemunch "github.com/becker63/searchbench-go/internal/adapters/backend/jcodemunch"
	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
)

// Environment variables name MCP server launch commands (stdin/stdout JSON-RPC).
// Values are executed as `sh -c <value>` so quoting behaves like a shell command line.
const (
	envJCodeMunchCommand       = "SEARCHBENCH_JCODEMUNCH_COMMAND"
	envIterativeContextCommand = "SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND"
)

// preparedToolFactory builds evaluator tools for one spec and returns an optional
// cleanup callback once the evaluator run finishes (MCP sessions, temp policy files).
type preparedToolFactory func(ctx context.Context, spec run.Spec) ([]tool.BaseTool, func(), error)

func wrapLegacyEvaluatorToolFactory(f EvaluatorToolFactory) preparedToolFactory {
	if f == nil {
		return nil
	}
	return func(ctx context.Context, spec run.Spec) ([]tool.BaseTool, func(), error) {
		tools, err := f(spec)
		return tools, nil, err
	}
}

func defaultPreparedToolFactory() preparedToolFactory {
	return func(ctx context.Context, spec run.Spec) ([]tool.BaseTool, func(), error) {
		switch spec.System.Backend {
		case domain.BackendFake:
			tools, err := evaluatorfake.ToolFactory(spec)
			return tools, nil, err
		case domain.BackendJCodeMunch:
			return openJCodeMunchTools(ctx, spec)
		case domain.BackendIterativeContext:
			return openIterativeContextTools(ctx, spec)
		default:
			return nil, nil, fmt.Errorf("unsupported backend %q for default evaluator tools", spec.System.Backend)
		}
	}
}

func resolvePreparedToolFactory(req evaluationRequest) preparedToolFactory {
	if req.EvaluatorToolFactory != nil {
		return wrapLegacyEvaluatorToolFactory(req.EvaluatorToolFactory)
	}
	return defaultPreparedToolFactory()
}

func openJCodeMunchTools(ctx context.Context, spec run.Spec) ([]tool.BaseTool, func(), error) {
	line := strings.TrimSpace(os.Getenv(envJCodeMunchCommand))
	if line == "" {
		return nil, nil, fmt.Errorf("%s must be set to launch the jCodeMunch MCP server for backend %q",
			envJCodeMunchCommand, spec.System.Backend)
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", line)
	rt, err := jcodemunch.OpenCommand(ctx, jcodemunch.CommandConfig{
		Command:  cmd,
		RepoPath: strings.TrimSpace(string(spec.Match.Repo.Path)),
	})
	if err != nil {
		return nil, nil, err
	}
	tools, ferr := jcodemunch.EvaluatorToolFactory(rt)(spec)
	if ferr != nil {
		_ = rt.Close()
		return nil, nil, ferr
	}
	return tools, func() { _ = rt.Close() }, nil
}

func openIterativeContextTools(ctx context.Context, spec run.Spec) ([]tool.BaseTool, func(), error) {
	if spec.System.Policy == nil {
		return nil, nil, fmt.Errorf("iterative-context backend requires system.policy for score preparation")
	}
	line := strings.TrimSpace(os.Getenv(envIterativeContextCommand))
	if line == "" {
		return nil, nil, fmt.Errorf("%s must be set to launch the Iterative Context MCP server for backend %q",
			envIterativeContextCommand, spec.System.Backend)
	}
	policyPath, removePolicy, err := writePolicyTempFile(spec.System.Policy)
	if err != nil {
		return nil, nil, fmt.Errorf("write temporary policy file: %w", err)
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", line)
	symbol := strings.TrimSpace(spec.System.Policy.Entrypoint)
	install := iterativecontext.ScoreInstallParams{
		PolicyPath: policyPath,
		PolicyID:   spec.System.Policy.ID.String(),
		Symbol:     symbol,
	}
	rt, err := iterativecontext.OpenCommand(ctx, iterativecontext.CommandConfig{
		Command:      cmd,
		RepoPath:     strings.TrimSpace(string(spec.Match.Repo.Path)),
		ScoreInstall: &install,
	})
	if err != nil {
		removePolicy()
		return nil, nil, err
	}
	tools, ferr := iterativecontext.EvaluatorToolFactory(rt)(spec)
	if ferr != nil {
		_ = rt.Close()
		removePolicy()
		return nil, nil, ferr
	}
	cleanup := func() {
		_ = rt.Close()
		removePolicy()
	}
	return tools, cleanup, nil
}

func writePolicyTempFile(p *domain.PolicyArtifact) (path string, cleanup func(), err error) {
	f, err := os.CreateTemp("", "searchbench-ic-policy-*.py")
	if err != nil {
		return "", nil, err
	}
	path = f.Name()
	cleanup = func() { _ = os.Remove(path) }
	if _, err := f.WriteString(p.Source); err != nil {
		_ = f.Close()
		cleanup()
		return "", nil, err
	}
	if err := f.Close(); err != nil {
		cleanup()
		return "", nil, err
	}
	return path, cleanup, nil
}
