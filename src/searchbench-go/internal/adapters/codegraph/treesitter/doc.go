// Package treesitter builds SearchBench code graphs from materialized repositories
// using the tree-sitter Go grammar. Parsing runs in this adapter package; scoring
// should consume codegraph.Graph only.
//
// This package requires CGO (tree-sitter C runtime). With CGO disabled, use the
// stub implementation that returns [ErrCGODisabled].
package treesitter
