package compare

import (
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/run"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// Logger is the narrow event sink compare uses for orchestration visibility.
//
// Surface packages may implement this with richer developer or structured
// logging, but compare itself should not depend on presentation packages.
type Logger interface {
	ComparisonStarted(planID string, baseline domain.SystemRef, candidate domain.SystemRef, taskCount int, parallelismMode string, maxWorkers int)
	ComparisonCompleted(report report.CandidateReport)
	TaskStarted(task domain.TaskSpec)
	TaskCompleted(task domain.TaskSpec, baselineSucceeded bool, candidateSucceeded bool, regressionCount int)
	RunStarted(role domain.Role, spec run.Spec)
	RunExecuted(role domain.Role, executed run.ExecutedRun)
	RunScored(role domain.Role, executed run.ExecutedRun, scores score.ScoreSet)
	RunFailed(role domain.Role, failure run.RunFailure)
	ReportCreated(report report.CandidateReport)
}

type nopLogger struct{}

func (nopLogger) ComparisonStarted(string, domain.SystemRef, domain.SystemRef, int, string, int) {}
func (nopLogger) ComparisonCompleted(report.CandidateReport)                                     {}
func (nopLogger) TaskStarted(domain.TaskSpec)                                                    {}
func (nopLogger) TaskCompleted(domain.TaskSpec, bool, bool, int)                                 {}
func (nopLogger) RunStarted(domain.Role, run.Spec)                                               {}
func (nopLogger) RunExecuted(domain.Role, run.ExecutedRun)                                       {}
func (nopLogger) RunScored(domain.Role, run.ExecutedRun, score.ScoreSet)                         {}
func (nopLogger) RunFailed(domain.Role, run.RunFailure)                                          {}
func (nopLogger) ReportCreated(report.CandidateReport)                                           {}
