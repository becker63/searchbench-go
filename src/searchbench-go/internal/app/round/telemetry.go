package round

import (
	evaluatormodel "github.com/becker63/searchbench-go/internal/adapters/providers/evaluatormodel"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/liveconfig"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/usage"
)

// buildCanonicalTelemetry merges live fingerprints, seeds, generation config, and hashes (#91).
func buildCanonicalTelemetry(
	plan Plan,
	executions []EvaluatorExecution,
	registry *usage.HashRegistry,
) report.CanonicalReport {
	extra := report.CanonicalReport{}
	if registry != nil {
		extra.RequestHashes, extra.ResponseHashes = registry.Snapshot()
	}

	cfg, live := liveconfig.ConfigFromManifest(plan.ManifestPath)
	if live {
		if fp, err := liveconfig.InputFingerprint(cfg); err == nil && fp != "" {
			extra.InputFingerprint = fp
		}
	}

	spec, attempt := firstRunSpec(plan, executions)
	if spec != nil {
		extra.ModelSeed = liveconfig.ModelSeed(
			string(plan.Round.ID),
			string(spec.Match.ID),
			string(roleForSpec(*spec)),
			attempt,
		)
		gen := evaluatormodel.DefaultLiveGenerationConfig(*spec, attempt)
		extra.GenerationConfig = map[string]any{
			"temperature": gen.Temperature,
			"top_p":       gen.TopP,
			"seed":        gen.Seed,
		}
	}
	return extra
}

func firstRunSpec(plan Plan, executions []EvaluatorExecution) (*run.Spec, int) {
	if len(executions) > 0 {
		ex := executions[0]
		for m := range plan.Matches.Values() {
			if m.ID == ex.MatchID {
				spec := run.Spec{
					ID:    ex.RunID,
					Match: m,
				}
				spec.System = plan.Policies.Challenger
				return &spec, 1
			}
		}
	}
	return nil, 1
}
