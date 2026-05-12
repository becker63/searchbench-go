// Package iterativecontext connects SearchBench evaluator runs to an Iterative Context MCP server.
//
// The adapter owns MCP session lifecycle, optional repo root registration, harness-owned
// score installation ([Runtime.InstallScore], [Runtime.VerifyScore]) before evaluator tools
// are listed, and MCP tools/call dispatch for evaluator-facing localization tools.
//
// Evaluator-visible tool lists intentionally exclude admin/install/verify tools even if the
// server advertises them in tools/list.
//
// Production use starts a subprocess transport via [OpenCommand]. Tests should use [NewRuntime]
// with an in-memory MCP session from [github.com/modelcontextprotocol/go-sdk/mcp.NewInMemoryTransports].
package iterativecontext
