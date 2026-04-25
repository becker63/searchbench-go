package logging

import (
	"fmt"
	"time"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/score"
)

// ComparisonStarted logs the start of a baseline/candidate comparison.
func (l Logger) ComparisonStarted(planID string, baseline domain.SystemRef, candidate domain.SystemRef, taskCount int, parallelismMode string, maxWorkers int) {
	if l.mode == ModeDev {
		parallelism := parallelismMode
		if maxWorkers > 1 {
			parallelism = fmt.Sprintf("%s x%d", parallelismMode, maxWorkers)
		}
		l.devInfo(
			"comparison.started",
			fmt.Sprintf("%d tasks", taskCount),
			parallelism,
			"baseline="+systemLabel(baseline),
			"candidate="+systemLabel(candidate),
		)
		return
	}
	args := AppendKV(
		[]any{
			"task_count", taskCount,
			"parallelism_mode", parallelismMode,
			"max_workers", maxWorkers,
		},
		prefixedKV("baseline_", SystemRefKV(baseline)),
		prefixedKV("candidate_", SystemRefKV(candidate)),
	)
	if planID != "" {
		args = append([]any{"plan_id", planID}, args...)
	}
	l.base().Infow("comparison.started", args...)
}

// ComparisonCompleted logs completion of a baseline/candidate comparison.
func (l Logger) ComparisonCompleted(report report.CandidateReport) {
	if l.mode == ModeDev {
		return
	}
	l.base().Infow("comparison.completed", ReportSummaryKV(report)...)
}

// TaskStarted logs the start of one task comparison.
func (l Logger) TaskStarted(task domain.TaskSpec) {
	if l.mode == ModeDev {
		l.devInfo(
			"task.started",
			taskLabel(task),
			fmt.Sprintf("%s@%s", task.Repo.Name, shortRepoSHA(task.Repo.SHA)),
			string(task.Benchmark),
		)
		return
	}
	l.base().Infow("task.started", TaskKV(task)...)
}

// TaskCompleted logs the high-level outcome of one task comparison.
func (l Logger) TaskCompleted(task domain.TaskSpec, baselineSucceeded bool, candidateSucceeded bool, regressionCount int) {
	if l.mode == ModeDev {
		l.devInfo(
			"task.completed",
			taskLabel(task),
			"baseline "+l.renderStatus(baselineSucceeded),
			"candidate "+l.renderStatus(candidateSucceeded),
			fmt.Sprintf("%d regressions", regressionCount),
		)
		return
	}
	l.base().Infow(
		"task.completed",
		AppendKV(
			TaskKV(task),
			[]any{
				"baseline_succeeded", baselineSucceeded,
				"candidate_succeeded", candidateSucceeded,
				"regression_count", regressionCount,
			},
		)...,
	)
}

// RunStarted logs the beginning of one role-specific run.
func (l Logger) RunStarted(role domain.Role, spec run.Spec) {
	if l.mode == ModeDev {
		l.devInfo(
			"run.started",
			l.renderRole(role),
			runTaskLabel(spec),
			systemLabel(spec.System.Ref()),
		)
		return
	}
	l.base().Infow("run.started", AppendKV(RoleKV(role), RunSpecKV(spec))...)
}

// RunExecuted logs the successful execution of one role-specific run.
func (l Logger) RunExecuted(role domain.Role, executed run.ExecutedRun) {
	if l.mode == ModeDev {
		parts := []string{
			l.renderRole(role),
			runTaskLabel(executed.Spec()),
			fmt.Sprintf("%d files", len(executed.Prediction.Files)),
			fmt.Sprintf("%d tokens", executed.Usage.TotalTokens),
			formatCost(executed.Usage.CostUSD),
		}
		if !executed.StartedAt.IsZero() && !executed.FinishedAt.IsZero() {
			parts = append(parts, executed.FinishedAt.Sub(executed.StartedAt).Round(time.Millisecond).String())
		}
		l.devInfo("run.executed", parts...)
		return
	}
	args := AppendKV(
		RoleKV(role),
		RunSpecKV(executed.Spec()),
		[]any{
			"predicted_files_count", len(executed.Prediction.Files),
			"input_tokens", executed.Usage.InputTokens,
			"output_tokens", executed.Usage.OutputTokens,
			"total_tokens", executed.Usage.TotalTokens,
			"cost_usd", executed.Usage.CostUSD,
		},
	)
	if !executed.StartedAt.IsZero() && !executed.FinishedAt.IsZero() {
		args = append(args, "duration_seconds", executed.FinishedAt.Sub(executed.StartedAt).Seconds())
	}
	l.base().Infow("run.executed", args...)
}

// RunScored logs the successful scoring of one executed run.
func (l Logger) RunScored(role domain.Role, executed run.ExecutedRun, scores score.ScoreSet) {
	if l.mode == ModeDev {
		l.devInfo(
			"run.scored",
			l.renderRole(role),
			runTaskLabel(executed.Spec()),
			l.renderMetric("gold_hop", formatFloat(float64(scores.GoldHop.Value))),
			l.renderMetric("issue_hop", formatFloat(float64(scores.IssueHop.Value))),
			l.renderMetric("token_efficiency", formatFloat(float64(scores.TokenEfficiency.Value))),
			l.renderMetric("cost", formatCost(float64(scores.Cost.Value))),
			l.renderMetric("composite", formatFloat(float64(scores.Composite.Value))),
		)
		return
	}
	l.base().Infow(
		"run.scored",
		AppendKV(RoleKV(role), RunSpecKV(executed.Spec()), ScoreSetKV(scores))...,
	)
}

// RunFailed logs one report-facing run failure.
func (l Logger) RunFailed(role domain.Role, failure run.RunFailure) {
	if l.mode == ModeDev {
		l.devWarn(
			"run.failed",
			l.renderRole(role),
			failure.TaskID.String(),
			string(failure.Stage),
			truncate(failure.Message, 120),
		)
		return
	}
	l.base().Warnw("run.failed", AppendKV(RoleKV(role), FailureKV(failure))...)
}

// ReportCreated logs the creation of the final candidate report.
func (l Logger) ReportCreated(report report.CandidateReport) {
	if l.mode == ModeDev {
		l.devInfo(
			"report.created",
			report.ID.String(),
			summarizeReport(report),
		)
		return
	}
	l.base().Infow("report.created", ReportSummaryKV(report)...)
}

func prefixedKV(prefix string, kv []any) []any {
	out := make([]any, 0, len(kv))
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			continue
		}
		out = append(out, prefix+key, kv[i+1])
	}
	return out
}
