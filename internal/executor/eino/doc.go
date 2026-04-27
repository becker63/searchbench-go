// Package eino implements the bounded Eino-backed evaluator execution path.
//
// One evaluator Run may include multiple model turns and multiple tool calls
// inside Eino's internal agent loop, but it still returns exactly one final
// normalized prediction or one typed evaluator failure.
//
// SearchBench-Go owns prompt rendering orchestration, retry policy, strict
// prediction finalization, evaluator-local phase/failure typing, and context
// cancellation propagation. Eino owns the internal model/tool loop. The
// harness can supply a run-level turn bound through SystemSpec.Runtime.MaxSteps,
// which the evaluator maps to Eino's iteration limit.
//
// Per-run Eino callbacks are composed through the sibling callbacks package so
// observability wiring stays separate from evaluator business logic.
//
// This package does not own writer repair flows, CLI validation pipelines,
// benchmark schemas, scoring, backend adapters, or MCP integration.
package eino
