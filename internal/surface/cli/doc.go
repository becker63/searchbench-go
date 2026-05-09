// Package cli defines the small Searchbench command-line surface.
//
// The CLI is intentionally thin: it parses user intent, configures logging and
// rendering, and calls the app round/evaluation packages. It does not own the
// SearchBench game model, backend runtime, scoring algorithms, or round report
// data structures.
package cli
