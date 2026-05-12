package iterativecontext

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CommandConfig starts an MCP server subprocess over stdio and completes the MCP handshake.
type CommandConfig struct {
	// Command is the MCP server command (stdin/stdout JSON-RPC). Required.
	Command *exec.Cmd
	// RepoPath, when non-empty, is registered as an MCP root before connecting.
	RepoPath string
	// Client, when non-empty, is used as the MCP client; otherwise a default SearchBench client is created.
	Client *mcp.Client
	// ScoreInstall, when non-nil, runs [PrepareScore] immediately after a successful connect.
	ScoreInstall *ScoreInstallParams
}

// OpenCommand connects to an Iterative Context MCP server launched as a subprocess.
func OpenCommand(ctx context.Context, cfg CommandConfig) (*Runtime, error) {
	if cfg.Command == nil {
		return nil, &Error{Kind: KindSession, Op: "open command", Err: fmt.Errorf("nil command")}
	}
	client := cfg.Client
	if client == nil {
		client = mcp.NewClient(&mcp.Implementation{Name: "searchbench-go", Version: "dev"}, nil)
	}

	if cfg.RepoPath != "" {
		uri, err := filePathToRootURI(cfg.RepoPath)
		if err != nil {
			return nil, &Error{Kind: KindSession, Op: "repo root uri", Err: err}
		}
		client.AddRoots(&mcp.Root{URI: uri})
	}

	transport := &mcp.CommandTransport{Command: cfg.Command}
	sess, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, &Error{Kind: KindSession, Op: "mcp connect", Err: err}
	}
	rt := NewRuntime(sess)

	if cfg.ScoreInstall != nil {
		if err := PrepareScore(ctx, rt, *cfg.ScoreInstall); err != nil {
			_ = rt.Close()
			return nil, err
		}
	}

	return rt, nil
}
