// Package fake collapses deterministic local-round stand-ins behind the evaluator
// path (model/tool factories, GraphProvider, Scorer, Decider, dataset.MatchSource).
// Together they simulate one full local Manifest-driven Evaluation without wiring
// real backends or datasets.
package fake
