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

	policySource := "def score(task):\n    return 'challenger'\n"

	// Define the policies being compared and the benchmark tasks.
	plan := NewPlan(
		domain.NewPair(
			exampleIncumbentPolicy(),
			exampleChallengerPolicy(policySource),
		),
		domain.NewNonEmpty(
			exampleTask(domain.MatchID("task-1"), domain.RepoRelPath("pkg/bug1.go")),
			exampleTask(domain.MatchID("task-2"), domain.RepoRelPath("pkg/bug2.go")),
		),
	)

	// Inject fake runtime dependencies and run the comparison.
	got, err := exampleRunner(fixedTestTime()).Run(context.Background(), plan)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Inspect the resulting report at a high level.
	if got.Decision.Decision != report.DecisionPromoteChallenger {
		t.Fatalf("Decision = %q, want %q", got.Decision.Decision, report.DecisionPromoteChallenger)
	}
	if len(got.Runs.Incumbent) != 2 || len(got.Runs.Challenger) != 2 {
		t.Fatalf("unexpected run counts: incumbent=%d challenger=%d", len(got.Runs.Incumbent), len(got.Runs.Challenger))
	}
	if len(got.Failures.Incumbent) != 0 || len(got.Failures.Challenger) != 0 {
		t.Fatal("expected no failures")
	}
	if len(got.Regressions) != 0 {
		t.Fatal("expected no regressions")
	}
	if got.Spec.Policies.Challenger.Policy == nil {
		t.Fatal("expected report-safe challenger policy ref")
	}

	assertReportDoesNotLeakPolicySource(t, got, policySource)
}

func exampleIncumbentPolicy() domain.SystemSpec {
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

func exampleChallengerPolicy(policySource string) domain.SystemSpec {
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

func exampleTask(id domain.MatchID, gold domain.RepoRelPath) domain.MatchSpec {
	return domain.MatchSpec{
		ID:        id,
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("repo/example"),
			SHA:  domain.RepoSHA("abc123"),
			Path: domain.HostPath("/tmp/repo"),
		},
		Input: domain.MatchInput{
			Title: "Find issue " + id.String(),
			Body:  "Locate bug for " + id.String(),
		},
		Oracle: domain.MatchOracle{
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
		NewRunID: func(role domain.Role, task domain.MatchSpec, system domain.SystemRef) domain.RunID {
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

func assertReportDoesNotLeakPolicySource(t *testing.T, got report.RoundReport, policySource string) {
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
