// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import (
	"context"

	"github.com/apple/pkl-go/pkl"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/runmode"
)

type SearchBenchRound struct {
	Game Game `pkl:"game"`

	Name string `pkl:"name"`

	Mode runmode.RunMode `pkl:"mode"`

	Dataset Dataset `pkl:"dataset"`

	Interfaces Interfaces `pkl:"interfaces"`

	Policies Policies `pkl:"policies"`

	Artifacts Artifacts `pkl:"artifacts"`

	Agents Agents `pkl:"agents"`

	Evaluation *Evaluation `pkl:"evaluation"`

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
