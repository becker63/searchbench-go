package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/becker63/searchbench-go/internal/compare"
	"github.com/becker63/searchbench-go/internal/surface/console"
)

// DemoReportCmd runs the fake end-to-end Searchbench demo comparison.
type DemoReportCmd struct {
	Tasks      int    `default:"2" help:"Number of fake tasks to compare."`
	MaxWorkers int    `default:"1" help:"Maximum task-level workers for the fake comparison."`
	Output     string `enum:"pretty,json,both" default:"pretty" help:"Output format for the candidate report."`
}

// Run executes the demo comparison and prints the resulting candidate report.
func (c *DemoReportCmd) Run(ctx context.Context, app *App) error {
	if c.Tasks <= 0 {
		return errors.New("tasks must be greater than zero")
	}
	if c.MaxWorkers <= 0 {
		return errors.New("max workers must be greater than zero")
	}

	if app == nil {
		app = &App{}
	}

	plan := demoPlan(c.Tasks)
	runner := demoRunner(demoTime(), app.Log, c.MaxWorkers)
	if c.MaxWorkers > 1 {
		runner.Parallelism = compare.Parallelism{
			Mode:       compare.ExecutionParallel,
			MaxWorkers: c.MaxWorkers,
		}
	}

	out, err := runner.Run(ctx, plan)
	if err != nil {
		return err
	}

	opts := console.Options{
		Color:   !app.NoColor,
		Width:   app.Width,
		Verbose: false,
	}

	switch c.Output {
	case "pretty":
		_, err = fmt.Fprintln(app.stdout(), console.RenderCandidateReport(out, opts))
		return err
	case "json":
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(app.stdout(), string(data))
		return err
	case "both":
		if _, err := fmt.Fprintln(app.stdout(), console.RenderCandidateReport(out, opts)); err != nil {
			return err
		}
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(app.stdout(), "\n"+string(data))
		return err
	default:
		return fmt.Errorf("unsupported output format %q", c.Output)
	}
}

func (a *App) stdout() io.Writer {
	if a == nil || a.Stdout == nil {
		return nilWriter{}
	}
	return a.Stdout
}

type nilWriter struct{}

func (nilWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
