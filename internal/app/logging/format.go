package logging

import (
	"fmt"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/run"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func shortSHA(value string) string {
	if len(value) <= 8 {
		return value
	}
	return value[:8]
}

func shortFingerprint(value domain.SystemFingerprint) string {
	return shortSHA(string(value))
}

func shortRepoSHA(value domain.RepoSHA) string {
	return shortSHA(string(value))
}

func systemLabel(system domain.SystemRef) string {
	return fmt.Sprintf("%s/%s", system.Backend, system.Model.Name)
}

func roleLabel(role domain.Role) string {
	return string(role)
}

func taskLabel(task domain.TaskSpec) string {
	return task.ID.String()
}

func runTaskLabel(spec run.Spec) string {
	return spec.Task.ID.String()
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

func formatCost(value float64) string {
	if value > 0 && value < 0.1 {
		return fmt.Sprintf("$%.4f", value)
	}
	return fmt.Sprintf("$%.2f", value)
}

func summarizeScores(scores score.ScoreSet) string {
	return strings.Join([]string{
		"gold_hop=" + formatFloat(float64(scores.GoldHop.Value)),
		"issue_hop=" + formatFloat(float64(scores.IssueHop.Value)),
		"token_efficiency=" + formatFloat(float64(scores.TokenEfficiency.Value)),
		"cost=" + formatCost(float64(scores.Cost.Value)),
		"composite=" + formatFloat(float64(scores.Composite.Value)),
	}, " ")
}

func summarizeReport(report report.CandidateReport) string {
	return strings.Join([]string{
		string(report.Decision.Decision),
		fmt.Sprintf("baseline=%d ok/%d failed", len(report.Runs.Baseline), len(report.Failures.Baseline)),
		fmt.Sprintf("candidate=%d ok/%d failed", len(report.Runs.Candidate), len(report.Failures.Candidate)),
		fmt.Sprintf("%d regressions", len(report.Regressions)),
	}, " · ")
}

func truncate(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	if max <= 1 {
		return value[:max]
	}
	return value[:max-1] + "…"
}
