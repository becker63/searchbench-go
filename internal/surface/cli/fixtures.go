package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/becker63/searchbench-go/internal/app/compare"
	"github.com/becker63/searchbench-go/internal/app/logging"
	"github.com/becker63/searchbench-go/internal/pure/codegraph"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

const demoPolicySource = "def score(task):\n    return \"challenger\"\n"

func demoPlan(taskCount int) compare.Plan {
	tasks := make([]domain.MatchSpec, 0, taskCount)
	for i := 1; i <= taskCount; i++ {
		tasks = append(tasks, demoTask(i))
	}

	return compare.NewPlan(
		domain.NewPair(
			demoIncumbentPolicy(),
			demoChallengerPolicy(demoPolicySource),
		),
		domain.NewNonEmpty(tasks[0], tasks[1:]...),
	)
}

func demoIncumbentPolicy() domain.SystemSpec {
	return domain.SystemSpec{
		ID:      domain.SystemID("incumbent-system"),
		Name:    "Incumbent",
		Backend: domain.BackendJCodeMunch,
		Model: domain.ModelSpec{
			Provider: "openai",
			Name:     "gpt-incumbent",
		},
		PromptBundle: domain.PromptBundleRef{
			Name:    "bundle",
			Version: "v1",
		},
		Runtime: domain.RuntimeConfig{
			MaxSteps: 5,
		},
	}
}

func demoChallengerPolicy(policySource string) domain.SystemSpec {
	policy := domain.NewPythonPolicy(domain.PolicyID("policy-1"), policySource, "score")
	return domain.SystemSpec{
		ID:      domain.SystemID("challenger-system"),
		Name:    "Challenger",
		Backend: domain.BackendIterativeContext,
		Model: domain.ModelSpec{
			Provider: "openai",
			Name:     "gpt-challenger",
		},
		PromptBundle: domain.PromptBundleRef{
			Name:    "bundle",
			Version: "v2",
		},
		Policy: &policy,
		Runtime: domain.RuntimeConfig{
			MaxSteps: 7,
		},
	}
}

func demoTask(index int) domain.MatchSpec {
	taskID := domain.MatchID(fmt.Sprintf("task-%d", index))
	gold := domain.RepoRelPath(fmt.Sprintf("pkg/bug%d.go", index))

	return domain.MatchSpec{
		ID:        taskID,
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("repo/example"),
			SHA:  domain.RepoSHA("abc123"),
			Path: domain.HostPath("/tmp/repo"),
		},
		Input: domain.MatchInput{
			Title: "Find issue " + taskID.String(),
			Body:  "Locate bug for " + taskID.String(),
		},
		Oracle: domain.MatchOracle{
			GoldFiles: []domain.RepoRelPath{gold},
		},
	}
}

func demoRunner(now time.Time, logger logging.Logger, maxWorkers int) compare.Runner {
	return compare.Runner{
		Executor:      demoExecutor{now: now},
		GraphProvider: demoGraphProvider{},
		Scorer:        demoScorer{},
		Decider:       demoDecider{},
		NewRunID: func(role domain.Role, task domain.MatchSpec, system domain.SystemRef) domain.RunID {
			return domain.RunID(fmt.Sprintf("%s-%s-%s", role, task.ID, system.ID))
		},
		NewReportID: func() domain.ReportID {
			return domain.ReportID("demo-report")
		},
		Now: func() time.Time {
			return now
		},
		Logger: logger.Named("demo"),
	}
}

func demoTime() time.Time {
	return time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
}

type demoExecutor struct {
	now time.Time
}

func (d demoExecutor) Execute(_ context.Context, spec run.Spec) (run.ExecutedRun, error) {
	planned := run.NewPlanned(spec)
	prepared := run.NewPrepared(planned, domain.SessionID("session-"+spec.ID.String()))

	predictionFile := domain.RepoRelPath("pkg/incumbent.go")
	reasoning := "incumbent path"
	if spec.System.Backend == domain.BackendIterativeContext {
		predictionFile = domain.RepoRelPath("pkg/challenger.go")
		reasoning = "challenger path"
	}

	return run.NewExecuted(
		prepared,
		domain.Prediction{
			Files:     []domain.RepoRelPath{predictionFile},
			Reasoning: reasoning,
		},
		domain.UsageSummary{
			InputTokens:  100,
			OutputTokens: 20,
			TotalTokens:  120,
			CostUSD:      0.05,
		},
		domain.TraceID("trace-"+spec.ID.String()),
		d.now,
		d.now.Add(2*time.Second),
	), nil
}

type demoGraphProvider struct{}

func (demoGraphProvider) GraphForTask(_ context.Context, task domain.MatchSpec) (codegraph.Graph, error) {
	store := codegraph.NewStore()
	fileID := codegraph.NodeID("file-" + task.ID.String())
	fnID := codegraph.NodeID("fn-" + task.ID.String())

	if err := store.AddNode(codegraph.NewFileNode(fileID, task.Oracle.GoldFiles[0])); err != nil {
		return nil, err
	}
	if err := store.AddNode(codegraph.NewFunctionNode(fnID, task.Oracle.GoldFiles[0], "score", 1, 10)); err != nil {
		return nil, err
	}
	if err := store.AddEdge(codegraph.NewEdge(fileID, fnID, codegraph.EdgeContains)); err != nil {
		return nil, err
	}
	return store.Build()
}

type demoScorer struct{}

func (demoScorer) Score(_ context.Context, input score.Input) (score.ScoreSet, error) {
	if input.Run.Spec().System.Backend == domain.BackendIterativeContext {
		return score.NewScoreSet(
			score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 1},
			score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 1},
			score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.9},
			score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.1},
			score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.95},
		)
	}

	return score.NewScoreSet(
		score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 4},
		score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 5},
		score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.4},
		score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.6},
		score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.3},
	)
}

type demoDecider struct{}

func (demoDecider) Decide(_ []report.ScoreComparison, regressions []report.Regression) report.Decision {
	if len(regressions) == 0 {
		return report.Decision{
			Decision: report.DecisionPromoteChallenger,
			Reason:   "challenger improved every required metric",
		}
	}

	return report.Decision{
		Decision: report.DecisionReview,
		Reason:   "challenger has regressions",
	}
}
