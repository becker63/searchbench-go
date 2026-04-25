// Package score defines Searchbench scoring vocabulary and helpers.
//
// It owns typed metric values, complete ScoreSet validation, metric direction
// semantics, score comparisons, and ScoredRun, which represents a run whose
// required metrics were computed successfully.
//
// It does not own backend execution, report generation, or concrete scoring
// algorithms beyond the stable interfaces and helper shapes used by the rest of
// the system.
package score
