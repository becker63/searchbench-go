package optimizer

import (
	"context"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestRenderIncludesExpectedSections(t *testing.T) {
	t.Parallel()

	spec := sampleOptimizerSpec()
	input, err := InputFromSpec(spec)
	if err != nil {
		t.Fatalf("InputFromSpec() error = %v", err)
	}

	prompt, err := Render(context.Background(), input)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{
		"<searchbench-prompt>",
		"challenger-policy-round-002",
		"challenger_policy.round-002.py",
		"iterative_context.selection_policy.v1",
		"Improve policy using only parent evidence.",
		"def score(task):",
		"localization-v1",
		"PROMOTE",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestRenderOmitsDeniedEvidence(t *testing.T) {
	t.Parallel()

	spec := sampleOptimizerSpec()
	input, err := InputFromSpec(spec)
	if err != nil {
		t.Fatalf("InputFromSpec() error = %v", err)
	}

	prompt, err := Render(context.Background(), input)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, forbidden := range []string{
		"src/search_target.go",
		"gold_files",
	} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt unexpectedly contains %q:\n%s", forbidden, prompt)
		}
	}
}

func sampleOptimizerSpec() pureoptimizer.Spec {
	return pureoptimizer.Spec{
		Target: pureoptimizer.Target{
			InputArtifactID:  domain.ArtifactID("challenger-policy-round-001"),
			OutputArtifactID: domain.ArtifactID("challenger-policy-round-002"),
			OutputName:       "challenger_policy.round-002.py",
			InterfaceID:      "iterative_context.selection_policy.v1",
		},
		Agent: pureoptimizer.AgentConfig{
			SystemPrompt: "Improve policy using only parent evidence.",
		},
		Evidence: pureoptimizer.Evidence{
			ParentRound: pureoptimizer.ParentRoundRef{
				ArtifactID: domain.ArtifactID("parent-eval"),
				BundleID:   "example-round-001",
			},
			IncludedKinds: []string{"report_summary", "round_evidence", "objective_result", "challenger_policy"},
			DeniedKinds:   []string{"gold_labels", "oracle_files", "raw_dataset_answers"},
			InputPolicy: pureoptimizer.PolicySource{
				ArtifactID:  domain.ArtifactID("challenger-policy-round-001"),
				InterfaceID: "iterative_context.selection_policy.v1",
				Source:      "def score(task):\n    return []\n",
			},
			ReportSummary: &pureoptimizer.ReportSummary{
				ReportID:       domain.ReportID("report-example-round-001"),
				Decision:       "PROMOTE",
				DecisionReason: "candidate improves the composite score",
			},
			ScoreEvidence: &score.ScoreEvidenceDocument{
				SchemaVersion: score.EvidenceSchemaVersion,
				ReportID:      domain.ReportID("report-example-round-001"),
			},
			ObjectiveResult: &score.ObjectiveResult{
				SchemaVersion: score.ObjectiveSchemaVersion,
				ObjectiveID:   "localization-v1",
				Values: []score.ObjectiveValue{{
					Name:  "final",
					Value: 0.86,
				}},
				Final: "final",
			},
		},
	}
}
