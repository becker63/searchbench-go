// Package localrun composes the existing SearchBench-Go seams into the
// smallest manifest-driven local fake end-to-end path.
//
// It is intentionally a surface-layer composition package. Pkl remains a
// configuration and scoring surface; Go remains the source of truth for fake
// execution, report construction, score evidence projection, objective
// invocation, and bundle writing.
package localrun
