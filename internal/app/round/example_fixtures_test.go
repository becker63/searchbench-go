package round

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/testing/modeltest"
)

func TestWriteExampleRoundFixtures(t *testing.T) {
	if os.Getenv("SEARCHBENCH_WRITE_EXAMPLE_FIXTURES") != "1" {
		t.Skip("set SEARCHBENCH_WRITE_EXAMPLE_FIXTURES=1 to regenerate checked-in example bundles")
	}

	requirePkl(t)

	root := repoRoot(t)
	localBundleRoot := filepath.Join(root, "configs", "rounds", "local-ic-vs-jcodemunch", "artifacts")
	optimizeBundleRoot := filepath.Join(root, "configs", "rounds", "optimize-ic", "artifacts")

	resetExampleBundleDir(t, filepath.Join(localBundleRoot, "games", "code-localization", "rounds", "round-001"))
	resetExampleBundleDir(t, filepath.Join(optimizeBundleRoot, "games", "code-localization", "rounds", "round-002"))

	if _, err := Run(context.Background(), Input{
		EvaluationManifestPath: filepath.Join(root, "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl"),
		BundleRootOverride:     localBundleRoot,
		RoundID:                "round-001",
		Now: func() time.Time {
			return time.Date(2026, 5, 12, 13, 0, 0, 0, time.UTC)
		},
	}); err != nil {
		t.Fatalf("write local example bundle: %v", err)
	}

	if _, err := Run(context.Background(), Input{
		EvaluationManifestPath: filepath.Join(root, "configs", "rounds", "optimize-ic", "round.pkl"),
		BundleRootOverride:     optimizeBundleRoot,
		RoundID:                "round-002",
		Now: func() time.Time {
			return time.Date(2026, 5, 12, 14, 0, 0, 0, time.UTC)
		},
		OptimizerModelFactory: scriptedExampleOptimizerFactory(),
	}); err != nil {
		t.Fatalf("write optimize example bundle: %v", err)
	}
}

func resetExampleBundleDir(t *testing.T, path string) {
	t.Helper()
	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("RemoveAll(%q) error = %v", path, err)
	}
}

func scriptedExampleOptimizerFactory() OptimizerModelFactory {
	return func() (model.ToolCallingChatModel, error) {
		return modeltest.NewScriptedModel(
			modeltest.ScriptedResponse{
				Message: schema.AssistantMessage(`{"artifact_id":"next-challenger-round-002","artifact_name":"next_challenger_policy.round-002.py","interface_id":"iterative_context.selection_policy.v1","code":"def score(match):\n    return []\n","summary":"generated challenger narrows the frontier"}`, nil),
			},
		), nil
	}
}
