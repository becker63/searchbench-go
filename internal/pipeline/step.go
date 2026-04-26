package pipeline

import (
	"strings"
	"time"
)

// CommandSpec defines one allowlisted local validation step.
type CommandSpec struct {
	Name    string
	Command []string
	CWD     string
	Timeout time.Duration
}

// StepResult records the typed result of one local validation step.
type StepResult struct {
	Name                string
	Command             []string
	CWD                 string
	Passed              bool
	ExitCode            int
	Stdout              string
	Stderr              string
	Duration            time.Duration
	Skipped             bool
	InfrastructureError error
	TimedOut            bool
}

// Failed reports whether the step ended in a classified failure.
func (r StepResult) Failed() bool {
	return !r.Passed && !r.Skipped
}

// CommandString joins the recorded argv for deterministic rendering.
func (r StepResult) CommandString() string {
	return strings.Join(r.Command, " ")
}

// StdoutTail returns a bounded tail of stdout for feedback formatting.
func (r StepResult) StdoutTail(maxChars int) string {
	return tailString(r.Stdout, maxChars)
}

// StderrTail returns a bounded tail of stderr for feedback formatting.
func (r StepResult) StderrTail(maxChars int) string {
	return tailString(r.Stderr, maxChars)
}

func tailString(value string, maxChars int) string {
	value = strings.TrimSpace(value)
	if maxChars <= 0 || value == "" {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= maxChars {
		return value
	}
	return "..." + string(runes[len(runes)-maxChars:])
}
