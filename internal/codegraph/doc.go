// Package codegraph defines the Searchbench code graph model.
//
// It owns graph nodes, edges, paths, the mutable Builder interface used during
// ingestion, and the read-only Graph interface used during scoring.
//
// It does not own tree-sitter ingestion logic or scoring policy. Those higher-
// level concerns should depend on Builder or Graph rather than on the Store
// implementation or its gograph backing.
package codegraph
