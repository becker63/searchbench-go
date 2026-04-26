package pipeline

import (
	"fmt"
	"strings"
)

const defaultStepTailChars = 400

// FormatPipelineFeedback converts a classification into deterministic, bounded
// retry-safe text.
func FormatPipelineFeedback(classification Classification, maxChars int) string {
	if maxChars <= 0 {
		maxChars = 2000
	}

	sections := []struct {
		heading string
		steps   []StepResult
	}{
		{heading: "## INFRASTRUCTURE FAILURES", steps: classification.InfrastructureFailures},
		{heading: "## GENERATION FAILURES", steps: classification.GenerationFailures},
		{heading: "## FORMAT ERRORS", steps: classification.FormatErrors},
		{heading: "## TYPE ERRORS", steps: classification.TypeErrors},
		{heading: "## LINT ERRORS", steps: classification.LintErrors},
		{heading: "## TEST FAILURES", steps: classification.TestFailures},
		{heading: "## PASSED STEPS", steps: classification.PassedSteps},
	}

	var builder strings.Builder
	for _, section := range sections {
		if len(section.steps) == 0 {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(section.heading)
		for _, step := range section.steps {
			builder.WriteString("\n")
			builder.WriteString(formatStep(step))
			if builder.Len() >= maxChars {
				return truncateString(builder.String(), maxChars)
			}
		}
	}

	return truncateString(builder.String(), maxChars)
}

func formatStep(step StepResult) string {
	var builder strings.Builder
	status := fmt.Sprintf("- %s: %s", step.Name, step.CommandString())
	if step.ExitCode >= 0 {
		status += fmt.Sprintf(" (exit %d)", step.ExitCode)
	}
	if step.InfrastructureError != nil {
		status += fmt.Sprintf(" infrastructure=%q", step.InfrastructureError.Error())
	}
	builder.WriteString(status)

	if stdout := step.StdoutTail(defaultStepTailChars); stdout != "" {
		builder.WriteString("\n  stdout: ")
		builder.WriteString(stdout)
	}
	if stderr := step.StderrTail(defaultStepTailChars); stderr != "" {
		builder.WriteString("\n  stderr: ")
		builder.WriteString(stderr)
	}

	return builder.String()
}

func truncateString(value string, maxChars int) string {
	if maxChars <= 0 {
		return ""
	}

	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= maxChars {
		return string(runes)
	}
	if maxChars <= 3 {
		return string(runes[:maxChars])
	}
	return string(runes[:maxChars-3]) + "..."
}
