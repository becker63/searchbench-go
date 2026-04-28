// Package compare coordinates baseline/candidate comparisons.
//
// It owns orchestration only: validating executable plans, running task
// comparisons, applying task-level parallelism policy, accumulating results,
// and emitting candidate reports.
//
// It does not own concrete backend sessions, graph ingestion, metric
// algorithms, prompt construction, or telemetry integrations. Those enter
// through small interfaces such as Executor, GraphProvider, Scorer, and
// Decider.
package compare
