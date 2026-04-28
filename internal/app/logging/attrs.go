package logging

import (
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/run"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// AppendKV concatenates flat Zap sugar key-value slices.
func AppendKV(parts ...[]any) []any {
	total := 0
	for _, part := range parts {
		total += len(part)
	}

	out := make([]any, 0, total)
	for _, part := range parts {
		out = append(out, part...)
	}
	return out
}

// RoleKV returns the structured logging fields for one comparison role.
func RoleKV(role domain.Role) []any {
	return []any{"role", role}
}

// TaskKV returns the safe task identity fields for structured logging.
//
// It intentionally omits TaskOracle and other scorer-only task details.
func TaskKV(task domain.TaskSpec) []any {
	return []any{
		"task_id", task.ID,
		"benchmark", task.Benchmark,
		"repo_name", task.Repo.Name,
		"repo_sha", task.Repo.SHA,
		"repo_path", task.Repo.Path,
	}
}

// RepoKV returns structured logging fields for one repo snapshot.
func RepoKV(repo domain.RepoSnapshot) []any {
	return []any{
		"repo_name", repo.Name,
		"repo_sha", repo.SHA,
		"repo_path", repo.Path,
	}
}

// SystemRefKV returns safe structured logging fields for one system identity.
func SystemRefKV(system domain.SystemRef) []any {
	return AppendKV(
		[]any{
			"system_id", system.ID,
			"system_name", system.Name,
			"system_fingerprint", system.Fingerprint,
			"backend", system.Backend,
			"model_provider", system.Model.Provider,
			"model_name", system.Model.Name,
			"prompt_bundle_name", system.PromptBundle.Name,
			"prompt_bundle_version", system.PromptBundle.Version,
			"policy_present", system.Policy != nil,
		},
		PolicyRefKV(system.Policy),
	)
}

// SystemSpecKV returns safe structured logging fields for one executable system.
//
// It always logs the report-safe SystemRef view instead of the raw SystemSpec.
func SystemSpecKV(system domain.SystemSpec) []any {
	return SystemRefKV(system.Ref())
}

// PolicyRefKV returns safe structured logging fields for one policy reference.
func PolicyRefKV(policy *domain.PolicyRef) []any {
	if policy == nil {
		return nil
	}
	return []any{
		"policy_id", policy.ID,
		"policy_language", policy.Language,
		"policy_sha256", policy.SHA256,
		"policy_entrypoint", policy.Entrypoint,
	}
}

// RunSpecKV returns structured logging fields for one planned run request.
func RunSpecKV(spec run.Spec) []any {
	ref := spec.System.Ref()
	return AppendKV(
		[]any{
			"run_id", spec.ID,
			"task_id", spec.Task.ID,
		},
		SystemRefKV(ref),
	)
}

// FailureKV returns structured logging fields for one report-facing run failure.
func FailureKV(failure run.RunFailure) []any {
	return []any{
		"run_id", failure.RunID,
		"task_id", failure.TaskID,
		"system_id", failure.System,
		"stage", failure.Stage,
		"message", failure.Message,
	}
}

// ScoreSetKV returns stable metric fields for one complete score set.
func ScoreSetKV(scores score.ScoreSet) []any {
	out := make([]any, 0, 10)
	for point := range scores.Points() {
		out = append(out, string(point.Name), point.Value)
	}
	return out
}

// ReportSummaryKV returns a compact report summary suitable for structured logs.
func ReportSummaryKV(report report.CandidateReport) []any {
	out := []any{
		"report_id", report.ID,
		"decision", report.Decision.Decision,
		"baseline_runs", len(report.Runs.Baseline),
		"candidate_runs", len(report.Runs.Candidate),
		"baseline_failures", len(report.Failures.Baseline),
		"candidate_failures", len(report.Failures.Candidate),
		"regressions", len(report.Regressions),
		"comparisons", len(report.Comparisons),
	}
	if !report.CreatedAt.IsZero() {
		out = append(out, "created_at", report.CreatedAt.Format(time.RFC3339Nano))
	}
	return out
}
