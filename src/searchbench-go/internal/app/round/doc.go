// Package round is the single canonical app-layer entry point for executing
// a SearchBench round.
//
// It composes manifest resolution, fake match evaluation, evidence
// construction, decision capture, durable bundle writing, and an optional
// next-challenger proposal into one Run function. Callers (CLI surface,
// future schedulers) depend only on this package; the older sibling
// packages internal/app/evaluation and internal/app/compare have been
// folded in here and behind round/internal/compare.
//
// Pkl remains a configuration and scoring surface; Go remains the source
// of truth for fake execution, report construction, round evidence
// projection, objective invocation, and bundle writing.
package round
