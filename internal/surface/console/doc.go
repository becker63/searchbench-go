// Package console renders Searchbench artifacts for human terminal output.
//
// It owns terminal formatting only. Structured report data remains in package
// report, and orchestration remains in package compare.
//
// Renderers must use report-safe identities and must not print executable
// policy source, task oracle details, or raw oversized artifacts by default.
package console
