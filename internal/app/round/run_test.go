package round

import (
        "context"
        "encoding/json"
        "os"
        "os/exec"
        "path/filepath"
        "strings"
        "testing"
        "time"

        "github.com/cloudwego/eino/components/model"
        "github.com/cloudwego/eino/schema"

        "github.com/becker63/searchbench-go/internal/adapters/executor/eino"
        "github.com/becker63/searchbench-go/internal/pure/domain"
        run "github.com/becker63/searchbench-go/internal/pure/execution"
        "github.com/becker63/searchbench-go/internal/testing/modeltest"
)

func TestRoundAdvancesNextChallengerFromEvidence(t *testing.T) {
        t.Parallel()

        requirePkl(t)

        root := repoRoot(t)
        evaluationManifest := filepath.Join(root, "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")
        optimizationManifest := filepath.Join(root, "configs", "rounds", "optimize-ic", "round.pkl")
        inputPolicyPath := filepath.Join(root, "configs", "rounds", "optimize-ic", "policies", "challenger_policy.py")
        bundleRoot := filepath.Join(t.TempDir(), "artifacts")

        evaluationManifestBefore := mustReadFile(t, evaluationManifest)
        optimizationManifestBefore := mustReadFile(t, optimizationManifest)
        inputPolicyBefore := mustReadFile(t, inputPolicyPath)

        var evaluatorModels []*modeltest.ScriptedModel
        evaluatorFactory := func(spec run.Spec) (model.ToolCallingChatModel, error) {
                prediction := `{"predicted_files":["src/incumbent_guess.go"],"reasoning":"incumbent fake evaluator stayed conservative"}`
                if spec.System.Backend == domain.BackendIterativeContext {
                        prediction = `{"predicted_files":["src/search_target.go"],"reasoning":"challenger fake evaluator used structural evidence to localize the issue"}`
                }
                scripted := modeltest.NewScriptedModel(
                        modeltest.ScriptedResponse{
                                Message: schema.AssistantMessage("", []schema.ToolCall{{
                                        ID: "call-1",
                                        Function: schema.FunctionCall{
                                                Name:      "resolve_and_expand",
                                                Arguments: `{"query":"retry interceptor"}`,
                                        },
                                }}),
                        },
                        modeltest.ScriptedResponse{
                                Message: schema.AssistantMessage(prediction, nil),
                        },
                )
                evaluatorModels = append(evaluatorModels, scripted)
                return scripted, nil
        }

        optimizerModel := modeltest.NewScriptedModel(
                modeltest.ScriptedResponse{
                        Message: schema.AssistantMessage(`{"artifact_id":"next-challenger-round-002","artifact_name":"next_challenger_policy.round-002.py","interface_id":"iterative_context.selection_policy.v1","code":"def score(task):\n    return []\n","summary":"challenger narrows the frontier using the parent evidence"}`, nil),
                },
        )

        result, err := Run(context.Background(), Input{
                EvaluationManifestPath:   evaluationManifest,
                OptimizationManifestPath: optimizationManifest,
                BundleRootOverride:       bundleRoot,
                RoundID:                  "round-001",
                OptimizerBundleID:        "fakee2e-optimizer",
                Now: func() time.Time {
                        return time.Date(2026, 5, 8, 20, 30, 0, 0, time.UTC)
                },
                EvaluatorModelFactory: evaluatorFactory,
                OptimizerModelFactory: func() (model.ToolCallingChatModel, error) {
                        return optimizerModel, nil
                },
        })
        if err != nil {
                t.Fatalf("Run() error = %v", err)
        }
        if result.RoundResult == nil || result.NextChallengerResult == nil {
                t.Fatalf("result = %#v, want round evaluation and next challenger results", result)
        }

        for _, path := range []string{
                filepath.Join(result.RoundBundle, "COMPLETE"),
                filepath.Join(result.RoundBundle, "resolved-round.json"),
                filepath.Join(result.RoundBundle, "round-report.json"),
                filepath.Join(result.RoundBundle, "round-report.txt"),
                filepath.Join(result.RoundBundle, "evidence.pkl"),
                filepath.Join(result.RoundBundle, "decision.json"),
                filepath.Join(result.RoundBundle, "objective.json"),
                filepath.Join(result.OptimizerBundle, "COMPLETE"),
                filepath.Join(result.OptimizerBundle, "next_challenger_policy.round-002.py"),
                filepath.Join(result.OptimizerBundle, "optimizer_result.json"),
        } {
                if _, err := os.Stat(path); err != nil {
                        t.Fatalf("os.Stat(%q) error = %v", path, err)
                }
        }

        if !strings.HasPrefix(result.RoundBundle, filepath.Join(bundleRoot, "games", "code-localization", "rounds")) {
                t.Fatalf("RoundBundle = %q, want game-scoped round bundle root under %q", result.RoundBundle, bundleRoot)
        }
        if !strings.HasPrefix(result.OptimizerBundle, filepath.Join(bundleRoot, "optimizer")) {
                t.Fatalf("OptimizerBundle = %q, want Go-owned bundle root under %q", result.OptimizerBundle, bundleRoot)
        }

        var optimizerResolved struct {
                ParentRound struct {
                        BundlePath string `json:"bundle_path"`
                } `json:"parent_round"`
        }
        decodeJSONFile(t, filepath.Join(result.OptimizerBundle, "resolved-next-challenger.json"), &optimizerResolved)
        if got, want := optimizerResolved.ParentRound.BundlePath, result.RoundBundle; got != want {
                t.Fatalf("optimizer parent bundle path = %q, want %q", got, want)
        }

        if got := string(mustReadFile(t, evaluationManifest)); got != string(evaluationManifestBefore) {
                t.Fatal("evaluation manifest was mutated")
        }
        if got := string(mustReadFile(t, optimizationManifest)); got != string(optimizationManifestBefore) {
                t.Fatal("optimization manifest was mutated")
        }
        if got := string(mustReadFile(t, inputPolicyPath)); got != string(inputPolicyBefore) {
                t.Fatal("input policy was mutated")
        }

        if len(evaluatorModels) != 2 {
                t.Fatalf("len(evaluatorModels) = %d, want 2", len(evaluatorModels))
        }
        for _, scripted := range evaluatorModels {
                if len(scripted.Calls()) == 0 {
                        t.Fatal("scripted evaluator model recorded zero calls")
                }
        }
        if len(optimizerModel.Calls()) == 0 {
                t.Fatal("scripted optimizer model recorded zero calls")
        }

        if len(result.RoundResult.EvaluatorExecutions) != 2 {
                t.Fatalf("len(EvaluatorExecutions) = %d, want 2", len(result.RoundResult.EvaluatorExecutions))
        }
        for _, execution := range result.RoundResult.EvaluatorExecutions {
                if !containsEvaluatorPhase(execution.Result.Phases, eino.PhaseRunEvaluator) {
                        t.Fatalf("execution phases = %#v, want run_evaluator", execution.Result.Phases)
                }
                if !containsEvaluatorPhase(execution.Result.Phases, eino.PhaseComplete) {
                        t.Fatalf("execution phases = %#v, want complete", execution.Result.Phases)
                }
        }

        if !strings.Contains(result.NextChallengerResult.Optimizer.RenderedPrompt, "<objective-result>") {
                t.Fatalf("optimizer prompt missing objective evidence summary:\n%s", result.NextChallengerResult.Optimizer.RenderedPrompt)
        }
        if strings.Contains(result.NextChallengerResult.Optimizer.RenderedPrompt, "oracle_files") {
                t.Fatalf("optimizer prompt leaked denied evidence:\n%s", result.NextChallengerResult.Optimizer.RenderedPrompt)
        }
}

func TestRoundResolutionFailurePreventsNextChallenger(t *testing.T) {
        t.Parallel()

        requirePkl(t)

        optimizerFactoryCalled := false
        _, err := Run(context.Background(), Input{
                EvaluationManifestPath:   filepath.Join(repoRoot(t), "configs", "rounds", "missing", "round.pkl"),
                OptimizationManifestPath: filepath.Join(repoRoot(t), "configs", "rounds", "optimize-ic", "round.pkl"),
                OptimizerModelFactory: func() (model.ToolCallingChatModel, error) {
                        optimizerFactoryCalled = true
                        return modeltest.NewScriptedModel(), nil
                },
        })
        if err == nil {
                t.Fatal("expected error")
        }
        if !strings.Contains(err.Error(), "Cannot find module") && !strings.Contains(err.Error(), "no such file or directory") {
                t.Fatalf("err = %v, want missing manifest failure", err)
        }
        if optimizerFactoryCalled {
                t.Fatal("optimizer factory should not run after parent evaluation failure")
        }
}

func TestRoundRequiresOptimizerFactoryWhenOptimizationConfigured(t *testing.T) {
        t.Parallel()

        requirePkl(t)

        root := repoRoot(t)
        _, err := Run(context.Background(), Input{
                EvaluationManifestPath:   filepath.Join(root, "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl"),
                OptimizationManifestPath: filepath.Join(root, "configs", "rounds", "optimize-ic", "round.pkl"),
                BundleRootOverride:       filepath.Join(t.TempDir(), "artifacts"),
                RoundID:                  "missing-optimizer-factory",
                Now: func() time.Time {
                        return time.Date(2026, 5, 8, 20, 30, 0, 0, time.UTC)
                },
        })
        if err == nil {
                t.Fatal("expected error when optimization manifest is configured without an optimizer factory")
        }
        if !strings.Contains(err.Error(), "OptimizerModelFactory is required") {
                t.Fatalf("err = %v, want OptimizerModelFactory required failure", err)
        }
}

func containsEvaluatorPhase(phases []eino.Phase, want eino.Phase) bool {
        for _, phase := range phases {
                if phase == want {
                        return true
                }
        }
        return false
}

func requirePkl(t *testing.T) {
        t.Helper()
        if _, err := exec.LookPath("pkl"); err != nil {
                t.Skip("pkl CLI not available on PATH")
        }
}

func repoRoot(t *testing.T) string {
        t.Helper()
        root, err := filepath.Abs(filepath.Join("..", "..", ".."))
        if err != nil {
                t.Fatalf("filepath.Abs(repo root) error = %v", err)
        }
        return root
}

func mustReadFile(t *testing.T, path string) []byte {
        t.Helper()
        data, err := os.ReadFile(path)
        if err != nil {
                t.Fatalf("os.ReadFile(%q) error = %v", path, err)
        }
        return data
}

func decodeJSONFile(t *testing.T, path string, target any) {
        t.Helper()
        if err := json.Unmarshal(mustReadFile(t, path), target); err != nil {
                t.Fatalf("json.Unmarshal(%q) error = %v", path, err)
        }
}
