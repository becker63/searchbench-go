package eino

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
	"github.com/becker63/searchbench-go/internal/testing/modeltest"
)

func TestOptimizerConstructionIsCold(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel()
	optimizer, err := NewOptimizer(OptimizerConfig{
		Model: model,
		ValidateProposal: func(context.Context, pureoptimizer.NextChallengerProposal) (ProposalValidationResult, *pureoptimizer.Failure) {
			return ProposalValidationResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewOptimizer() error = %v", err)
	}
	if optimizer == nil {
		t.Fatal("expected optimizer")
	}
	if calls := model.Calls(); len(calls) != 0 {
		t.Fatalf("len(model.Calls()) = %d, want 0", len(calls))
	}
}

func TestOptimizerRunSuccess(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(modeltest.ScriptedResponse{
		Message: schema.AssistantMessage(`{"artifact_id":"next-challenger-round-002","artifact_name":"next_next_challenger_policy.round-002.py","interface_id":"iterative_context.selection_policy.v1","code":"def score(task):\n    return []\n","summary":"small change"}`, nil),
	})
	optimizer, err := NewOptimizer(OptimizerConfig{
		Model: model,
		ValidateProposal: func(context.Context, pureoptimizer.NextChallengerProposal) (ProposalValidationResult, *pureoptimizer.Failure) {
			return ProposalValidationResult{
				Results: []pipeline.StepResult{{
					Name:     "python_compile",
					Command:  []string{"python3", "-m", "py_compile", "next_next_challenger_policy.round-002.py"},
					Passed:   true,
					ExitCode: 0,
				}},
				Classification: &pipeline.Classification{
					PassedSteps: []pipeline.StepResult{{
						Name: "python_compile",
					}},
				},
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewOptimizer() error = %v", err)
	}

	result := optimizer.Run(context.Background(), sampleOptimizerSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if !result.Success || result.Proposal == nil {
		t.Fatalf("result = %#v, want success proposal", result)
	}
	if got, want := result.Proposal.ArtifactName, "next_next_challenger_policy.round-002.py"; got != want {
		t.Fatalf("ArtifactName = %q, want %q", got, want)
	}
	if len(result.Attempts) != 1 || result.Attempts[0].State != pureoptimizer.AttemptStateAccepted {
		t.Fatalf("Attempts = %#v, want one accepted attempt", result.Attempts)
	}
}

func TestOptimizerRunRetriesMalformedProposal(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage("{\"artifact_id\":\"next-challenger-round-002\",\"artifact_name\":\"next_next_challenger_policy.round-002.py\",\"interface_id\":\"iterative_context.selection_policy.v1\",\"code\":\"```python\\ndef score(task):\\n    return []\\n```\"}", nil),
		},
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"artifact_id":"next-challenger-round-002","artifact_name":"next_next_challenger_policy.round-002.py","interface_id":"iterative_context.selection_policy.v1","code":"def score(task):\n    return []\n"}`, nil),
		},
	)
	optimizer, err := NewOptimizer(OptimizerConfig{
		Model: model,
		ValidateProposal: func(context.Context, pureoptimizer.NextChallengerProposal) (ProposalValidationResult, *pureoptimizer.Failure) {
			return ProposalValidationResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewOptimizer() error = %v", err)
	}

	result := optimizer.Run(context.Background(), sampleOptimizerSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if len(result.Attempts) != 2 {
		t.Fatalf("len(Attempts) = %d, want 2", len(result.Attempts))
	}
	if result.Attempts[0].Failure == nil || result.Attempts[0].Failure.Kind != pureoptimizer.FailureKindNextChallengerFailed {
		t.Fatalf("first attempt failure = %#v, want policy proposal failure", result.Attempts[0].Failure)
	}
	calls := model.Calls()
	if len(calls) != 2 {
		t.Fatalf("len(model.Calls()) = %d, want 2", len(calls))
	}
	secondPrompt := calls[1].Messages[len(calls[1].Messages)-1].Content
	if !strings.Contains(secondPrompt, "Previous attempt returned an invalid proposal") {
		t.Fatalf("second prompt missing retry feedback:\n%s", secondPrompt)
	}
}

func TestOptimizerRunRetriesPipelineFailure(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"artifact_id":"next-challenger-round-002","artifact_name":"next_next_challenger_policy.round-002.py","interface_id":"iterative_context.selection_policy.v1","code":"def score(task):\n    return []\n"}`, nil),
		},
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"artifact_id":"next-challenger-round-002","artifact_name":"next_next_challenger_policy.round-002.py","interface_id":"iterative_context.selection_policy.v1","code":"def score(task):\n    return []\n","summary":"retry success"}`, nil),
		},
	)
	callCount := 0
	optimizer, err := NewOptimizer(OptimizerConfig{
		Model: model,
		ValidateProposal: func(context.Context, pureoptimizer.NextChallengerProposal) (ProposalValidationResult, *pureoptimizer.Failure) {
			callCount++
			if callCount == 1 {
				classification := pipeline.Classification{
					TypeErrors: []pipeline.StepResult{{
						Name:     "python_compile",
						Command:  []string{"python3", "-m", "py_compile", "next_next_challenger_policy.round-002.py"},
						ExitCode: 1,
						Stderr:   "IndentationError: unexpected indent",
					}},
				}
				return ProposalValidationResult{
						Results:        []pipeline.StepResult{classification.TypeErrors[0]},
						Classification: &classification,
					}, &pureoptimizer.Failure{
						Phase:            pureoptimizer.PhaseRunPolicyPipeline,
						Kind:             pureoptimizer.FailureKindPolicyPipelineFailed,
						Message:          "policy validation pipeline failed",
						Attempt:          1,
						Retryable:        true,
						PipelineCategory: "validation",
						PipelineFeedback: pipeline.FormatPipelineFeedback(classification, 1200),
					}
			}
			return ProposalValidationResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewOptimizer() error = %v", err)
	}

	result := optimizer.Run(context.Background(), sampleOptimizerSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if len(result.Attempts) != 2 {
		t.Fatalf("len(Attempts) = %d, want 2", len(result.Attempts))
	}
	if result.Attempts[0].Failure == nil || result.Attempts[0].Failure.Kind != pureoptimizer.FailureKindPolicyPipelineFailed {
		t.Fatalf("first attempt failure = %#v, want pipeline failure", result.Attempts[0].Failure)
	}
	if result.Attempts[1].State != pureoptimizer.AttemptStateAccepted {
		t.Fatalf("second attempt = %#v, want accepted", result.Attempts[1])
	}
}

func TestFinalizeProposalRejectsTargetMismatch(t *testing.T) {
	t.Parallel()

	target := sampleOptimizerSpec().Target
	for _, tc := range []struct {
		name string
		raw  string
	}{
		{
			name: "wrong artifact id",
			raw:  `{"artifact_id":"wrong","artifact_name":"next_next_challenger_policy.round-002.py","interface_id":"iterative_context.selection_policy.v1","code":"def score(task):\n    return []\n"}`,
		},
		{
			name: "wrong artifact name",
			raw:  `{"artifact_id":"next-challenger-round-002","artifact_name":"wrong.py","interface_id":"iterative_context.selection_policy.v1","code":"def score(task):\n    return []\n"}`,
		},
		{
			name: "wrong interface",
			raw:  `{"artifact_id":"next-challenger-round-002","artifact_name":"next_next_challenger_policy.round-002.py","interface_id":"wrong","code":"def score(task):\n    return []\n"}`,
		},
		{
			name: "empty code",
			raw:  `{"artifact_id":"next-challenger-round-002","artifact_name":"next_next_challenger_policy.round-002.py","interface_id":"iterative_context.selection_policy.v1","code":""}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, failure := finalizeProposal(tc.raw, target, 1); failure == nil {
				t.Fatal("expected failure")
			}
		})
	}
}

func sampleOptimizerSpec() pureoptimizer.Spec {
	return pureoptimizer.Spec{
		Target: pureoptimizer.NextChallengerTarget{
			InputArtifactID:  domain.ArtifactID("challenger-policy-round-001"),
			OutputArtifactID: domain.ArtifactID("next-challenger-round-002"),
			OutputName:       "next_next_challenger_policy.round-002.py",
			InterfaceID:      "iterative_context.selection_policy.v1",
		},
		Agent: pureoptimizer.AgentConfig{
			SystemPrompt: "Improve the policy using bounded evidence.",
		},
		Evidence: pureoptimizer.NextChallengerEvidence{
			ParentRound: pureoptimizer.ParentRoundRef{
				ArtifactID: domain.ArtifactID("parent-bundle"),
				BundleID:   "example-round-001",
			},
			IncludedKinds: []string{"report_summary", "round_evidence", "objective_result", "challenger_policy"},
			DeniedKinds:   []string{"gold_labels", "oracle_files", "raw_dataset_answers"},
			InputPolicy: pureoptimizer.PolicySource{
				ArtifactID:  domain.ArtifactID("challenger-policy-round-001"),
				InterfaceID: "iterative_context.selection_policy.v1",
				Source:      "def score(task):\n    return []\n",
			},
		},
	}
}
