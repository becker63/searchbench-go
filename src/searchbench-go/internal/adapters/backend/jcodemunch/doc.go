// Package jcodemunch connects SearchBench evaluator runs to a jCodeMunch MCP server.
//
// The adapter owns MCP session lifecycle, repo root registration, MCP tool listing,
// conversion into Eino tools, and normalization of tool results for the evaluator.
// Setup failures (connect, list tools during tool construction) are distinct from
// per-invocation tool failures.
//
// Production use starts a subprocess transport via [OpenCommand]. Tests should use
// [NewRuntime] with an in-memory MCP session from [github.com/modelcontextprotocol/go-sdk/mcp.NewInMemoryTransports].
package jcodemunch
