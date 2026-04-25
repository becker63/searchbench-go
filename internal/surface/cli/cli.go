package cli

import (
	"context"
	"io"
	"os"

	"github.com/alecthomas/kong"

	"github.com/becker63/searchbench-go/internal/surface/logging"
)

// CLI is the root Searchbench command tree.
type CLI struct {
	DemoReport DemoReportCmd `cmd:"demo-report" help:"Run a complete fake comparison and render the candidate report."`

	LogFormat string `enum:"dev,json,none" default:"dev" help:"Log format."`
	NoColor   bool   `help:"Disable color in rendered reports."`
	Width     int    `default:"100" help:"Pretty report width."`
	Quiet     bool   `help:"Suppress structured logs."`
}

// App carries runtime services shared across CLI command handlers.
type App struct {
	Log     logging.Logger
	NoColor bool
	Width   int
	Quiet   bool
	Stdout  io.Writer
}

// Run executes the Searchbench CLI against the process stdio streams.
func Run(ctx context.Context, args []string) error {
	return RunWithWriters(ctx, args, os.Stdout, os.Stderr)
}

// RunWithWriters executes the Searchbench CLI against injected output streams.
//
// It exists so tests can capture pretty and JSON report output without relying
// on process-global stdout/stderr.
func RunWithWriters(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	var cli CLI
	parser, err := kong.New(
		&cli,
		kong.Name("searchbench"),
		kong.Description("Compare agentic code-search systems and render candidate reports."),
		kong.Writers(stdout, stderr),
	)
	if err != nil {
		return err
	}

	kctx, err := parser.Parse(args)
	if err != nil {
		return err
	}
	kctx.BindTo(ctx, (*context.Context)(nil))

	logger, cleanup, err := newLogger(cli.LogFormat, cli.Quiet, stderr, !cli.NoColor)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer func() {
			_ = cleanup()
		}()
	}

	app := &App{
		Log:     logger,
		NoColor: cli.NoColor,
		Width:   cli.Width,
		Quiet:   cli.Quiet,
		Stdout:  stdout,
	}
	return kctx.Run(ctx, app)
}

func newLogger(format string, quiet bool, stderr io.Writer, color bool) (logging.Logger, func() error, error) {
	if quiet || format == "none" {
		return logging.NewNop(), nil, nil
	}

	switch format {
	case "dev":
		return logging.NewDevelopmentWithWriter(stderr, color)
	case "json":
		return logging.NewProductionWithWriter(stderr)
	default:
		return logging.NewNop(), nil, nil
	}
}
