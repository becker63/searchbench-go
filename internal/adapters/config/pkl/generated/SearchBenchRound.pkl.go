// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import (
	"context"

	"github.com/apple/pkl-go/pkl"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/runmode"
)

// Root schema for a SearchBench **round** manifest.
//
// A round binds a dataset slice, incumbent/challenger policy systems, optional artifacts,
// and agent configurations. The Go runtime loads this file to resolve bundle paths, evaluation
// plans, and (when configured) optimization / next-challenger flows.
//
// Validation and projection to execution types are implemented in `internal/adapters/config/pkl`
// and `internal/app/round`. The `go.pkl` import attaches `pkl-go` struct tags for generated Go bindings.
type SearchBenchRound struct {
	// Stable game identity for this round (used in bundle layout and reporting).
	Game Game `pkl:"game"`

	// Human-readable round name; may influence default bundle IDs and reports.
	Name string `pkl:"name"`

	// Which top-level workflow this manifest participates in (evaluation-only vs optimization).
	Mode runmode.RunMode `pkl:"mode"`

	// Dataset selection: upstream benchmark id, language/config, and split.
	Dataset Dataset `pkl:"dataset"`

	// Declares interface contracts (e.g. iterative context selection policy) referenced by artifacts and policy bindings.
	Interfaces Interfaces `pkl:"interfaces"`

	// Incumbent and challenger policy **systems** for the comparative run (ids, backends, prompt bundles, runtime caps).
	Policies Policies `pkl:"policies"`

	// Optional artifact references (policy file on disk, next-challenger output, parent bundle, etc.).
	Artifacts Artifacts `pkl:"artifacts"`

	// Agents used when resolving rounds: evaluator (always required for `evaluation` mode) and optimizer (required for `optimization` mode).
	Agents Agents `pkl:"agents"`

	// Evaluation-mode wiring: must mirror `agents.evaluator`, bind systems to artifacts, and point at the scoring objective file.
	Evaluation *Evaluation `pkl:"evaluation"`

	// Optimization-mode wiring: must mirror `agents.optimizer`, parent bundle, target artifacts, and evidence policy for next-challenger proposal.
	Optimization *Optimization `pkl:"optimization"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a SearchBenchRound
func LoadFromPath(ctx context.Context, path string) (ret SearchBenchRound, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return ret, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a SearchBenchRound
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (SearchBenchRound, error) {
	var ret SearchBenchRound
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
