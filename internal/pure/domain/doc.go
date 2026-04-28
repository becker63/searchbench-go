// Package domain defines the stable Searchbench vocabulary.
//
// It owns identifiers, tasks, systems, policy refs, repo snapshots,
// predictions, usage summaries, artifact references, and other value types that
// are shared across execution, scoring, and reporting.
//
// It does not own run lifecycle, score computation, report generation,
// orchestration, backend sessions, or graph ingestion. Those higher-level
// concerns build on the domain model rather than living inside it.
package domain
