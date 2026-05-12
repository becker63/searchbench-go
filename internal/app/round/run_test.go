package round

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/testing/modeltest"
)

func TestRunGeneratedContinuationEvaluatesCurrentRound(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	optimizerModel := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"artifact_id":"next-challenger-round-002","artifact_name":"next_challenger_policy.round-002.py","interface_id":"iterative_context.selection_policy.v1","code":"def score(match):\n    return []\n","summary":"generated challenger narrows the frontier"}`, nil),
		},
	)

	result, err := Run(context.Background(), Input{
		EvaluationManifestPath: filepath.Join(repoRoot(t), "configs", "rounds", "generate-ic-from-local", "round.pkl"),
		BundleRootOverride:     filepath.Join(t.TempDir(), "artifacts"),
		RoundID:                "round-002",
		Now: func() time.Time {
			return time.Date(2026, 5, 12, 14, 0, 0, 0, time.UTC)
		},
		OptimizerModelFactory: func() (model.ToolCallingChatModel, error) {
			return optimizerModel, nil
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.RoundResult == nil {
		t.Fatal("RoundResult is nil")
	}
	if len(optimizerModel.Calls()) == 0 {
		t.Fatal("optimizer model recorded zero calls")
	}
	if got := result.RoundResult.RoundReport.Spec.Policies.Challenger.Policy; got == nil {
		t.Fatal("challenger policy was not materialized before evaluation")
	}
	if _, err := os.Stat(filepath.Join(result.RoundBundle, "policies", "next_challenger_policy.round-002.py")); err != nil {
		t.Fatalf("generated challenger artifact missing from round bundle: %v", err)
	}
}

func domainSystemRoleChallenger() domain.Role {
	return domain.RoleChallenger
}
