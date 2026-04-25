package score

import (
	"github.com/becker63/searchbench-go/internal/codegraph"
	"github.com/becker63/searchbench-go/internal/run"
)

// Input is the canonical input to the scoring engine.
//
// Scorers should not reach into backend sessions, raw traces, model messages,
// or filesystem state. Everything needed for required metrics should be here.
type Input struct {
	Run   run.ExecutedRun
	Graph codegraph.Graph
}
