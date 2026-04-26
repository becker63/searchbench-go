package pipeline

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"
)

// CommandRunner executes one local validation command.
type CommandRunner interface {
	Run(ctx context.Context, spec CommandSpec) StepResult
}

// ExecCommandRunner runs local commands through os/exec.
type ExecCommandRunner struct{}

// Allowlist is the exact argv allowlist for local validation steps.
type Allowlist struct {
	commands [][]string
}

// DefaultAllowlist returns the narrow repository-local command allowlist used
// by the evaluator validation pipeline.
func DefaultAllowlist() Allowlist {
	return Allowlist{
		commands: [][]string{
			{"templ", "generate"},
			{"gofmt", "-l", "."},
			{"go", "test", "./..."},
		},
	}
}

// Allows reports whether the exact command argv is allowed.
func (a Allowlist) Allows(command []string) bool {
	for _, allowed := range a.commands {
		if slices.Equal(allowed, command) {
			return true
		}
	}
	return false
}

// Runner executes a small allowlisted validation pipeline.
type Runner struct {
	CommandRunner CommandRunner
	Allowlist     Allowlist
}

// DefaultEvaluatorSteps returns the minimal local preflight used by the current
// evaluator proof seam.
func DefaultEvaluatorSteps(cwd string) []CommandSpec {
	return []CommandSpec{
		{
			Name:    "templ_generate",
			Command: []string{"templ", "generate"},
			CWD:     cwd,
		},
		{
			Name:    "gofmt_check",
			Command: []string{"gofmt", "-l", "."},
			CWD:     cwd,
		},
	}
}

// Run executes the supplied pipeline steps in order and stops at the first
// failed step.
func (r Runner) Run(ctx context.Context, steps []CommandSpec) []StepResult {
	commandRunner := r.CommandRunner
	if commandRunner == nil {
		commandRunner = ExecCommandRunner{}
	}
	allowlist := r.Allowlist
	if len(allowlist.commands) == 0 {
		allowlist = DefaultAllowlist()
	}

	results := make([]StepResult, 0, len(steps))
	for _, spec := range steps {
		result := StepResult{
			Name:     spec.Name,
			Command:  append([]string(nil), spec.Command...),
			CWD:      spec.CWD,
			ExitCode: -1,
		}

		if err := ctx.Err(); err != nil {
			result.InfrastructureError = err
		} else if !allowlist.Allows(spec.Command) {
			result.InfrastructureError = fmt.Errorf("command is not allowlisted: %s", strings.Join(spec.Command, " "))
		} else {
			runCtx := ctx
			cancel := func() {}
			if spec.Timeout > 0 {
				runCtx, cancel = context.WithTimeout(ctx, spec.Timeout)
			}
			result = commandRunner.Run(runCtx, spec)
			cancel()
			result.Name = spec.Name
			result.Command = append([]string(nil), spec.Command...)
			result.CWD = spec.CWD
			if result.ExitCode == 0 && result.InfrastructureError == nil {
				result.Passed = true
			}
			if err := runCtx.Err(); err != nil && result.InfrastructureError == nil {
				result.InfrastructureError = err
				result.TimedOut = errors.Is(err, context.DeadlineExceeded)
				result.Passed = false
			}
		}

		result = normalizeResult(spec, result)
		results = append(results, result)
		if result.Failed() {
			break
		}
	}

	return results
}

func (ExecCommandRunner) Run(ctx context.Context, spec CommandSpec) (result StepResult) {
	startedAt := time.Now()
	result = StepResult{
		Name:     spec.Name,
		Command:  append([]string(nil), spec.Command...),
		CWD:      spec.CWD,
		ExitCode: -1,
	}
	defer func() {
		result.Duration = time.Since(startedAt)
	}()

	if len(spec.Command) == 0 {
		result.InfrastructureError = errors.New("command is required")
		return result
	}

	cmd := exec.CommandContext(ctx, spec.Command[0], spec.Command[1:]...)
	if spec.CWD != "" {
		cmd.Dir = spec.CWD
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err == nil {
		result.ExitCode = 0
		return result
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		if ctx.Err() != nil {
			result.InfrastructureError = ctx.Err()
			result.TimedOut = errors.Is(ctx.Err(), context.DeadlineExceeded)
		}
		return result
	}

	result.InfrastructureError = err
	if ctx.Err() != nil {
		result.InfrastructureError = ctx.Err()
		result.TimedOut = errors.Is(ctx.Err(), context.DeadlineExceeded)
	}
	return result
}

func normalizeResult(spec CommandSpec, result StepResult) StepResult {
	if result.InfrastructureError != nil {
		result.Passed = false
		return result
	}

	switch spec.Name {
	case "gofmt_check":
		result.Passed = result.ExitCode == 0 && strings.TrimSpace(result.Stdout) == ""
	default:
		result.Passed = result.ExitCode == 0
	}

	return result
}
