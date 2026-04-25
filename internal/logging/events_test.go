package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/score"
)

func TestEventHelpersDoNotPanicWithNop(t *testing.T) {
	t.Parallel()

	logger := NewNop()
	spec, executed, scores, failure, report := sampleLogArtifacts(t)

	logger.ComparisonStarted("", spec.System.Ref(), spec.System.Ref(), 1, "sequential", 1)
	logger.TaskStarted(spec.Task)
	logger.RunStarted(domain.RoleCandidate, spec)
	logger.RunExecuted(domain.RoleCandidate, executed)
	logger.RunScored(domain.RoleCandidate, executed, scores)
	logger.RunFailed(domain.RoleCandidate, failure)
	logger.TaskCompleted(spec.Task, true, false, 1)
	logger.ReportCreated(report)
	logger.ComparisonCompleted(report)
}

func TestDevelopmentLoggerPrettyOutput(t *testing.T) {
	t.Parallel()

	spec, executed, scores, _, report := sampleLogArtifacts(t)
	var buf bytes.Buffer

	logger, cleanup, err := NewDevelopmentWithWriter(&buf, false)
	if err != nil {
		t.Fatalf("NewDevelopmentWithWriter() error = %v", err)
	}
	defer func() {
		_ = cleanup()
	}()

	logger = logger.Named("demo")
	logger.ComparisonStarted("", spec.System.Ref(), spec.System.Ref(), 2, "sequential", 1)
	logger.RunScored(domain.RoleCandidate, executed, scores)
	logger.ReportCreated(report)

	out := buf.String()
	for _, want := range []string{
		"comparison.started",
		"run.scored",
		"report.created",
		"PROMOTE",
		"gold_hop=",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("pretty output missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "def score") {
		t.Fatal("pretty output leaked policy source")
	}
	if strings.Contains(out, "\"source\"") {
		t.Fatal("pretty output leaked policy source field")
	}
}

func TestProductionLoggerJSONOutput(t *testing.T) {
	t.Parallel()

	spec, executed, scores, _, _ := sampleLogArtifacts(t)
	var buf bytes.Buffer

	logger, cleanup, err := NewProductionWithWriter(&buf)
	if err != nil {
		t.Fatalf("NewProductionWithWriter() error = %v", err)
	}
	defer func() {
		_ = cleanup()
	}()

	logger.RunScored(domain.RoleCandidate, executed, scores)
	logger.RunStarted(domain.RoleCandidate, spec)

	out := buf.String()
	for _, want := range []string{
		"run.scored",
		"gold_hop",
		"run.started",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("json output missing %q\n%s", want, out)
		}
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatal("expected json log lines")
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
}

func sampleLogArtifacts(t *testing.T) (run.Spec, run.ExecutedRun, score.ScoreSet, run.RunFailure, report.CandidateReport) {
	t.Helper()

	policy := domain.NewPythonPolicy(domain.PolicyID("policy-1"), "def score(task): return 'candidate'", "score")
	spec := run.NewSpec(
		domain.RunID("run-1"),
		domain.TaskSpec{
			ID:        domain.TaskID("task-1"),
			Benchmark: domain.BenchmarkLCA,
			Repo: domain.RepoSnapshot{
				Name: domain.RepoName("repo/example"),
				SHA:  domain.RepoSHA("abc123"),
				Path: domain.HostPath("/tmp/repo"),
			},
		},
		domain.SystemSpec{
			ID:      domain.SystemID("system-1"),
			Name:    "Candidate",
			Backend: domain.BackendIterativeContext,
			Model: domain.ModelSpec{
				Provider: "openai",
				Name:     "gpt-candidate",
			},
			PromptBundle: domain.PromptBundleRef{Name: "bundle"},
			Policy:       &policy,
		},
	)
	planned := run.NewPlanned(spec)
	prepared := run.NewPrepared(planned, domain.SessionID("session-1"))
	executed := run.NewExecuted(
		prepared,
		domain.Prediction{Files: []domain.RepoRelPath{"pkg/file.go"}},
		domain.UsageSummary{InputTokens: 10, OutputTokens: 5, TotalTokens: 15, CostUSD: 0.01},
		domain.TraceID("trace-1"),
		time.Unix(1, 0),
		time.Unix(3, 0),
	)
	scores, err := score.NewScoreSet(
		score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 1},
		score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 2},
		score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.8},
		score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.2},
		score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.9},
	)
	if err != nil {
		t.Fatalf("NewScoreSet() error = %v", err)
	}
	failure := run.NewFailure(spec, run.FailureScore, "boom")
	report := report.NewCandidateReport(
		domain.ReportID("report-1"),
		report.ComparisonSpec{},
		domain.NewPair([]score.ScoredRun{}, []score.ScoredRun{}),
		domain.NewPair([]run.RunFailure{}, []run.RunFailure{failure}),
		nil,
		nil,
		report.PromotionDecision{Decision: report.DecisionPromote},
	)
	return spec, executed, scores, failure, report
}
