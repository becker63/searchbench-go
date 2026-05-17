package cli

import (
	"context"
	"fmt"

	configpkl "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
)

// ValidateManifestCmd validates a Pkl round manifest without running it.
type ValidateManifestCmd struct {
	Manifest string `required:"" help:"Path to the round manifest." type:"path"`
}

// Run validates manifest wiring.
func (c *ValidateManifestCmd) Run(ctx context.Context, app *App) error {
	_ = ctx
	_ = app
	spec, err := configpkl.LoadFromPath(ctx, c.Manifest)
	if err != nil {
		return err
	}
	if err := configpkl.Validate(spec); err != nil {
		return err
	}
	roundID := ""
	if spec.Round != nil {
		roundID = spec.Round.Id
	}
	_, err = fmt.Fprintf(app.stdout(), "ok manifest=%s round_id=%s\n", c.Manifest, roundID)
	return err
}

// ValidateBundleCmd validates a completed bundle directory (#83).
type ValidateBundleCmd struct {
	BundlePath string `required:"" help:"Path to completed round bundle." type:"path"`
}

// Run validates bundle artifacts and prints canonical summary path.
func (c *ValidateBundleCmd) Run(ctx context.Context, app *App) error {
	_ = ctx
	canonical, err := validateBundle(c.BundlePath)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(app.stdout(), "ok bundle=%s passed=%v mode=%s\n", c.BundlePath, canonical.Passed, canonical.Mode)
	return err
}
