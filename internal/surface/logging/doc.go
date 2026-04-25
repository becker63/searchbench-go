// Package logging provides Searchbench event logging helpers.
//
// The package exposes one event-shaped logging surface while supporting two
// output modes:
//
//   - pretty human-readable development logs for interactive use
//   - structured Zap JSON logs for machine ingestion
//
// Logging defaults to report-safe views such as SystemRef instead of
// SystemSpec, PolicyRef instead of PolicyArtifact, and task identity instead of
// scorer-only oracle fields.
//
// Searchbench uses Zap directly for structured JSON output and does not try to
// abstract over logging implementations.
package logging
