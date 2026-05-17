package round

import (
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/report"
)

func TestConsolidateAttemptsStableSet(t *testing.T) {
	t.Parallel()
	outcomes := []AttemptOutcome{
		{AttemptID: 1, Completed: true, Passed: true, FinalVal: 0.9, TotalTokens: 100},
		{AttemptID: 2, Completed: true, Passed: true, FinalVal: 0.85, TotalTokens: 110},
		{AttemptID: 3, Completed: true, Passed: false, FinalVal: 0.2, Failures: map[string]int{"infrastructure": 1}, TotalTokens: 90},
	}
	result := ConsolidateAttempts(outcomes, ConsolidationInput{Requested: 3})
	if result.Summary == nil || result.Summary.Count != 3 {
		t.Fatalf("summary=%+v", result.Summary)
	}
	if result.Summary.PassRate < 0.66 || result.Summary.PassRate > 0.67 {
		t.Fatalf("pass_rate=%v", result.Summary.PassRate)
	}
}

func TestConsolidateAttemptsStabilityProbe(t *testing.T) {
	t.Parallel()
	fp := "a|b"
	outcomes := []AttemptOutcome{
		{AttemptID: 1, Completed: true, Passed: true, FinalVal: 1, TotalTokens: 100, PredictionFingerprint: fp},
		{AttemptID: 2, Completed: true, Passed: true, FinalVal: 1, TotalTokens: 105, PredictionFingerprint: fp},
	}
	result := ConsolidateAttempts(outcomes, ConsolidationInput{Requested: 2, Stability: true})
	if result.Summary == nil {
		t.Fatal("missing summary")
	}
	if result.Summary.Verdict != string(stabilityStable) {
		t.Fatalf("verdict=%q want %s", result.Summary.Verdict, stabilityStable)
	}
	if result.Passed != true {
		t.Fatal("expected stability pass")
	}
}

func TestConsolidateAttemptsPromotionVerdict(t *testing.T) {
	t.Parallel()
	outcomes := []AttemptOutcome{
		{AttemptID: 1, Completed: true, Passed: true, FinalVal: 1, TotalTokens: 50},
		{AttemptID: 2, Completed: true, Passed: true, FinalVal: 1, TotalTokens: 55},
		{AttemptID: 3, Completed: true, Passed: true, FinalVal: 1, TotalTokens: 60},
	}
	result := ConsolidateAttempts(outcomes, ConsolidationInput{Requested: 3})
	if result.Summary.Verdict != verdictPromote {
		t.Fatalf("verdict=%q", result.Summary.Verdict)
	}
	if result.Decision != string(report.DecisionPromoteChallenger) {
		t.Fatalf("decision=%q", result.Decision)
	}
}
