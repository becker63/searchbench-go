package compare

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/app/logging"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
)

func TestRunnerExampleComparison(t *testing.T) {
	t.Parallel()

	policySource := "def score(task):\n    return 'candidate'\n"

	// Define the systems being compared and the benchmark tasks.
	plan := NewPlan(
		domain.NewPair(
			exampleBaselineSystem(),
			exampleCandidateSystem(policySource),
		),
		domain.NewNonEmpty(
			exampleTask(domain.TaskID("task-1"), domain.RepoRelPath("pkg/bug1.go")),
			exampleTask(domain.TaskID("task-2"), domain.RepoRelPath("pkg/bug2.go")),
		),
	)

	// Inject fake runtime dependencies and run the comparison.
	got, err := exampleRunner(fixedTestTime()).Run(context.Background(), plan)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Inspect the resulting report at a high level.
	if got.Decision.Decision != report.DecisionPromote {
		t.Fatalf("Decision = %q, want %q", got.Decision.Decision, report.DecisionPromote)
	}
	if len(got.Runs.Baseline) != 2 || len(got.Runs.Candidate) != 2 {
		t.Fatalf("unexpected run counts: baseline=%d candidate=%d", len(got.Runs.Baseline), len(got.Runs.Candidate))
	}
	if len(got.Failures.Baseline) != 0 || len(got.Failures.Candidate) != 0 {
		t.Fatal("expected no failures")
	}
	if len(got.Regressions) != 0 {
		t.Fatal("expected no regressions")
	}
	if got.Spec.Systems.Candidate.Policy == nil {
		t.Fatal("expected report-safe candidate policy ref")
	}

	assertReportDoesNotLeakPolicySource(t, got, policySource)
}

func exampleBaselineSystem() domain.SystemSpec {
	return domain.SystemSpec{
		ID:      domain.SystemID("baseline-system"),
		Name:    "Baseline",
		Backend: domain.BackendJCodeMunch,
		Model: domain.ModelSpec{
			Provider: "openai",
			Name:     "gpt-baseline",
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

func exampleCandidateSystem(policySource string) domain.SystemSpec {
	policy := domain.NewPythonPolicy(domain.PolicyID("policy-1"), policySource, "score")
	return domain.SystemSpec{
		ID:      domain.SystemID("candidate-system"),
		Name:    "Candidate",
		Backend: domain.BackendIterativeContext,
		Model: domain.ModelSpec{
			Provider: "openai",
			Name:     "gpt-candidate",
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

func exampleTask(id domain.TaskID, gold domain.RepoRelPath) domain.TaskSpec {
	return domain.TaskSpec{
		ID:        id,
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("repo/example"),
			SHA:  domain.RepoSHA("abc123"),
			Path: domain.HostPath("/tmp/repo"),
		},
		Input: domain.TaskInput{
			Title: "Find issue " + id.String(),
			Body:  "Locate bug for " + id.String(),
		},
		Oracle: domain.TaskOracle{
			GoldFiles: []domain.RepoRelPath{gold},
		},
	}
}

func exampleRunner(now time.Time) Runner {
	return Runner{
		Executor:      fakeExecutor{now: now},
		GraphProvider: fakeGraphProvider{},
		Scorer:        fakeScorer{},
		Decider:       fakeDecider{},
		NewRunID: func(role domain.Role, task domain.TaskSpec, system domain.SystemRef) domain.RunID {
			return domain.RunID(fmt.Sprintf("%s-%s-%s", role, task.ID, system.ID))
		},
		NewReportID: func() domain.ReportID {
			return domain.ReportID("report-1")
		},
		Now: func() time.Time {
			return now
		},
		Logger: logging.NewNop(),
	}
}

func fixedTestTime() time.Time {
	return time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
}

func assertReportDoesNotLeakPolicySource(t *testing.T, got report.CandidateReport, policySource string) {
	t.Helper()

	data, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), policySource) {
		t.Fatal("report leaked policy source")
	}
	if strings.Contains(string(data), "\"source\"") {
		t.Fatal("report JSON should not include policy source field")
	}
}
