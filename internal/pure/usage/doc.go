// Package usage owns harness-level token usage accounting.
//
// It defines provider-neutral per-call records, aggregated run summaries, a
// reusable collector, and local token-estimation helpers. Execution-layer code
// such as Eino callbacks may feed raw model-call facts into this package, but
// this package does not depend on Eino callback types or tracing backends.
package usage
