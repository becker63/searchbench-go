// Package pipeline owns small typed local validation pipelines.
//
// It provides allowlisted CLI step execution, step result capture, failure
// classification, and bounded feedback formatting for evaluator-side retry and
// repair flows.
//
// The default evaluator step list in this package is the current local
// evaluator preflight seam used to prove typed CLI validation before model
// execution. It is not yet the final dataset-scale lifecycle for SearchBench
// runs; later work may choose different scheduling such as once-per-process
// preflight or post-write candidate validation.
//
// It does not own model execution, prompt rendering, backend runtimes, or the
// public SearchBench CLI surface.
package pipeline
