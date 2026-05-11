package cli

import (
	"context"
	"fmt"

	"github.com/becker63/searchbench-go/internal/app/round"
)

// RoundCmd groups public round commands.
type RoundCmd struct {
	Run RunCmd `cmd:"run" help:"Run one Pkl round manifest."`
}

// RunCmd executes one manifest-driven round.
type RunCmd struct {
	Manifest      string `required:"" help:"Path to the round manifest." type:"path"`
	BundleRoot    string `help:"Optional bundle root override." type:"path"`
	BundleID      string `help:"Optional bundle identifier override."`
	NoHumanReport bool   `help:"Skip writing the optional human-readable report artifact."`
}

// Run executes the manifest-driven round and prints a short summary.
func (c *RunCmd) Run(ctx context.Context, app *App) error {
	if app == nil {
		app = &App{}
	}

	record, err := round.Run(ctx, round.Input{
		EvaluationManifestPath: c.Manifest,
		BundleRootOverride:     c.BundleRoot,
		RoundID:                c.BundleID,
		DisableRenderReport:    c.NoHumanReport,
	})
	if err != nil {
		return err
	}

	result := record.RoundResult
	if result == nil {
		return fmt.Errorf("round: missing evaluation result")
	}

	objectiveID := ""
	if result.ObjectiveResult != nil {
		objectiveID = string(result.ObjectiveResult.ObjectiveID)
	}
	if _, err := fmt.Fprintf(
		app.stdout(),
		"bundle=%s\nreport_id=%s\nobjective=%s\nfinal=%s:%0.6f\n",
		result.Bundle.Path,
		result.ReportID,
		objectiveID,
		finalLabel(result),
		finalValue(result),
	); err != nil {
		return err
	}
	return nil
}

func finalLabel(result *round.Result) string {
	if result == nil || result.ObjectiveResult == nil {
		return "final"
	}
	return string(result.ObjectiveResult.Final)
}

func finalValue(result *round.Result) float64 {
	if result == nil || result.ObjectiveResult == nil {
		return 0
	}
	final, ok := result.ObjectiveResult.FinalValue()
	if !ok {
		return 0
	}
	return final.Value
}
