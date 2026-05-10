// Package compare coordinates incumbent/challenger comparisons.
//
// It owns orchestration only: validating executable plans, running task
// comparisons, applying task-level parallelism policy, accumulating results,
// and emitting round reports.
//
// It does not own concrete backend sessions, graph ingestion, metric
// algorithms, prompt construction, or telemetry integrations. Those enter
// through small interfaces such as Executor, GraphProvider, Scorer, and
// Decider.
package compare
