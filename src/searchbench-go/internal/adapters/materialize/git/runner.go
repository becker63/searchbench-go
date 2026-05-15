package gitmaterialize

import (
	"bytes"
	"context"
	"os/exec"
)

// GitRunner runs `git` subprocesses. Production code uses [ExecGitRunner];
// tests may inject a fake runner.
type GitRunner interface {
	RunGit(ctx context.Context, workDir string, args ...string) (stdout, stderr string, exitCode int, err error)
}

// ExecGitRunner implements [GitRunner] with os/exec.
type ExecGitRunner struct{}

func (ExecGitRunner) RunGit(ctx context.Context, workDir string, args ...string) (stdout, stderr string, exitCode int, err error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	var sout, serr bytes.Buffer
	cmd.Stdout = &sout
	cmd.Stderr = &serr
	err = cmd.Run()
	stdout = sout.String()
	stderr = serr.String()
	if err == nil {
		return stdout, stderr, 0, nil
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return stdout, stderr, ee.ExitCode(), nil
	}
	return stdout, stderr, -1, err
}
