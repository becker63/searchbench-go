package pipeline

import "strings"

// Classification groups pipeline results into stable failure categories.
type Classification struct {
	GenerationFailures     []StepResult
	FormatErrors           []StepResult
	TypeErrors             []StepResult
	LintErrors             []StepResult
	TestFailures           []StepResult
	InfrastructureFailures []StepResult
	PassedSteps            []StepResult
}

// HasFailures reports whether the classification contains any failed steps.
func (c Classification) HasFailures() bool {
	return len(c.GenerationFailures) > 0 ||
		len(c.FormatErrors) > 0 ||
		len(c.TypeErrors) > 0 ||
		len(c.LintErrors) > 0 ||
		len(c.TestFailures) > 0 ||
		len(c.InfrastructureFailures) > 0
}

// Classify groups step results into stable pipeline categories.
func Classify(results []StepResult) Classification {
	classification := Classification{}
	for _, result := range results {
		if result.Passed {
			classification.PassedSteps = append(classification.PassedSteps, result)
			continue
		}
		if result.Skipped {
			continue
		}
		if result.InfrastructureError != nil {
			classification.InfrastructureFailures = append(classification.InfrastructureFailures, result)
			continue
		}

		switch result.Name {
		case "templ_generate":
			classification.GenerationFailures = append(classification.GenerationFailures, result)
		case "gofmt_check":
			classification.FormatErrors = append(classification.FormatErrors, result)
		case "go_vet":
			if looksLikeTypeError(result) {
				classification.TypeErrors = append(classification.TypeErrors, result)
			} else {
				classification.LintErrors = append(classification.LintErrors, result)
			}
		case "go_test":
			if looksLikeTypeError(result) {
				classification.TypeErrors = append(classification.TypeErrors, result)
			} else {
				classification.TestFailures = append(classification.TestFailures, result)
			}
		default:
			classification.InfrastructureFailures = append(classification.InfrastructureFailures, result)
		}
	}
	return classification
}

func looksLikeTypeError(result StepResult) bool {
	combined := strings.ToLower(result.Stdout + "\n" + result.Stderr)
	patterns := []string{
		"[build failed]",
		"undefined:",
		"cannot use ",
		"no required module provides package",
		"syntax error:",
		"too many arguments in call",
		"not enough arguments in call",
		"missing return",
		"invalid operation:",
		"build constraints exclude all go files",
	}
	for _, pattern := range patterns {
		if strings.Contains(combined, pattern) {
			return true
		}
	}
	return false
}
