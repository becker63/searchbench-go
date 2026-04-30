package cli

import (
	"context"
	"fmt"

	"github.com/becker63/searchbench-go/internal/app/localrun"
)

// RunCmd executes one manifest-driven local fake experiment run.
type RunCmd struct {
	Manifest      string `required:"" help:"Path to the experiment manifest." type:"path"`
	BundleRoot    string `help:"Optional bundle root override." type:"path"`
	BundleID      string `help:"Optional bundle identifier override."`
	NoHumanReport bool   `help:"Skip writing the optional human-readable report artifact."`
}

// Run executes the manifest-driven local fake run and prints a short summary.
func (c *RunCmd) Run(ctx context.Context, app *App) error {
	if app == nil {
		app = &App{}
	}

	result, err := localrun.Run(ctx, localrun.Request{
		ManifestPath:        c.Manifest,
		BundleRootOverride:  c.BundleRoot,
		BundleID:            c.BundleID,
		DisableRenderReport: c.NoHumanReport,
	})
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		app.stdout(),
		"bundle=%s\nreport_id=%s\nobjective=%s\nfinal=%s:%0.6f\n",
		result.Bundle.Path,
		result.ReportID,
		result.ObjectiveResult.ObjectiveID,
		result.ObjectiveResult.Final,
		mustFinalValue(result),
	); err != nil {
		return err
	}
	return nil
}

func mustFinalValue(result localrun.Result) float64 {
	if result.ObjectiveResult == nil {
		return 0
	}
	final, ok := result.ObjectiveResult.FinalValue()
	if !ok {
		return 0
	}
	return final.Value
}
