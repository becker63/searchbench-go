// Package cli defines the small Searchbench command-line surface.
//
// The CLI is intentionally thin: it parses user intent, configures logging and
// rendering, and calls the existing comparison and reporting packages. It does
// not own the Searchbench model, backend runtime, scoring algorithms, or
// report data structures.
package cli
