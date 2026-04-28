// Package config owns the Pkl-backed experiment manifest surface.
//
// The package is intentionally an adapter-edge surface. It may depend on
// pkl-go generated bindings and handwritten validation, but it does not change
// the pure SearchBench center. Domain, run, score, report, compare, and
// executor packages remain authoritative for execution semantics.
//
// One manifest resolves to one Go config struct. After that resolution, Go
// validation enforces SearchBench-specific rules such as evaluator/writer
// separation and mode-specific requirements.
package config
