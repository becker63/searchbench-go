// Package report defines the Searchbench release artifact model.
//
// It owns report-safe comparison specs, metric comparisons, regressions,
// promotion decisions, task-aligned run sets, and CandidateReport, the central
// product emitted after a baseline/candidate comparison.
//
// It does not own executable system configuration or runtime orchestration.
// Executable inputs are modeled in package compare and package domain; reports
// intentionally operate on report-safe identities instead.
package report
