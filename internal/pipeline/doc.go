// Package pipeline owns small typed local validation pipelines.
//
// It provides allowlisted CLI step execution, step result capture, failure
// classification, and bounded feedback formatting for writer, repair, and
// candidate validation flows.
//
// It does not own model execution, prompt rendering, backend runtimes, or the
// public SearchBench CLI surface.
package pipeline
