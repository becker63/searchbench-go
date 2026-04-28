// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package generated

import (
	"context"

	"github.com/apple/pkl-go/pkl"
	"github.com/becker63/searchbench-go/internal/surface/config/generated/runmode"
)

type Experiment struct {
	Name string `pkl:"name"`

	Mode runmode.RunMode `pkl:"mode"`

	Dataset Dataset `pkl:"dataset"`

	Systems Systems `pkl:"systems"`

	Evaluator Evaluator `pkl:"evaluator"`

	Writer *Writer `pkl:"writer"`

	Scoring Scoring `pkl:"scoring"`

	OutputConfig Output `pkl:"outputConfig"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Experiment
func LoadFromPath(ctx context.Context, path string) (ret Experiment, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Experiment
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Experiment, error) {
	var ret Experiment
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
