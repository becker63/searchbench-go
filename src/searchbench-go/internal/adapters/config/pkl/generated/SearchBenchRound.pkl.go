// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

// Root schema for a SearchBench round manifest.
//
// The only author-facing entry point is `round`, which either defines a round
// from scratch or continues from an explicit completed bundle path.
type SearchBenchRound struct {
	// Stable game identity for this round (used in bundle layout and reporting).
	Game Game `pkl:"game"`

	// Human-readable round name; may influence default bundle IDs and reports.
	Name string `pkl:"name"`

	// Declares interface contracts referenced by policy artifacts.
	Interfaces Interfaces `pkl:"interfaces"`

	// Continuation-backed round surface used by concise game-level manifests.
	Round *RoundManifest `pkl:"round"`
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
