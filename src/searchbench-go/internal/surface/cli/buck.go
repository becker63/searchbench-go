package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/becker63/searchbench-go/internal/adapters/dataset/lca"
	"github.com/becker63/searchbench-go/internal/app/round"
	"github.com/becker63/searchbench-go/internal/pure/liveconfig"
)

// BuckCmd is the private Buck implementation surface (#91). Not a public API.
type BuckCmd struct {
	Round          BuckRoundCmd          `cmd:"round" help:"Repo-owned round operations for Buck targets."`
	ValidateBundle BuckValidateBundleCmd `cmd:"validate-bundle" help:"Validate a completed bundle directory."`
	Dataset        BuckDatasetCmd        `cmd:"dataset" help:"Dataset operations for Buck targets."`
}

// BuckRoundCmd runs one repo-owned round mode.
type BuckRoundCmd struct {
	Mode              string `enum:"validate,run,live_smoke,evaluate_n,stability_probe,validate_bundle" required:"" help:"Buck target mode."`
	Manifest          string `help:"Path to round.pkl." type:"path"`
	ArtifactRoot      string `help:"Artifact root for bundle output." type:"path"`
	BundlePath        string `help:"Completed bundle path (validate_bundle)." type:"path"`
	EvaluateAttempts  int    `default:"3" help:"Fresh attempts for evaluate_n."`
	StabilityAttempts int    `default:"5" help:"Fresh attempts for stability_probe."`
	DatasetConfig     string `default:"py" help:"LCA HF config."`
	DatasetSplit      string `default:"dev" help:"LCA HF split."`
	DatasetMaxItems   int    `default:"1" help:"LCA export row limit."`
	DatasetSkip       int    `default:"50" help:"LCA HF rows to skip."`
	RepoRoot          string `help:"Monorepo root for secrets and defaults." type:"path"`
}

// Run dispatches a Buck-owned round mode.
func (c *BuckRoundCmd) Run(ctx context.Context, app *App) error {
	_ = app
	repoRoot, err := c.resolveRepoRoot()
	if err != nil {
		return err
	}
	cfg := liveconfig.Default(repoRoot)
	manifest := strings.TrimSpace(c.Manifest)
	if manifest == "" {
		manifest = cfg.ManifestPath
	}
	artifactRoot := strings.TrimSpace(c.ArtifactRoot)
	if artifactRoot == "" {
		artifactRoot = cfg.ArtifactRoot
	}
	bundlePath := strings.TrimSpace(c.BundlePath)
	if bundlePath == "" {
		bundlePath = cfg.BundleDest
	}

	switch c.Mode {
	case "validate":
		if err := round.ValidateLiveManifest(ctx, manifest); err != nil {
			return err
		}
		_, err = fmt.Fprintf(app.stdout(), "ok manifest=%s\n", manifest)
		return err
	case "validate_bundle":
		canonical, err := validateBundle(bundlePath)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(app.stdout(), "ok bundle=%s passed=%v mode=%s\n", bundlePath, canonical.Passed, canonical.Mode)
		return err
	case "run":
		return (&RunCmd{
			Manifest:   manifest,
			BundleRoot: artifactRoot,
			BundleID:   liveconfig.RoundID,
		}).Run(ctx, app)
	case liveconfig.ModeLiveSmoke:
		return round.RunLiveSmoke(ctx, round.LiveModeInput{
			ManifestPath:               manifest,
			ArtifactRoot:               artifactRoot,
			BundlePath:                 bundlePath,
			RoundID:                    liveconfig.RoundID,
			RepoRoot:                   repoRoot,
			DatasetMaterializeCacheDir: cfg.MaterializeDir,
		})
	case liveconfig.ModeEvaluateN:
		return round.RunEvaluateN(ctx, round.LiveModeInput{
			ManifestPath:               manifest,
			ArtifactRoot:               artifactRoot,
			BundlePath:                 bundlePath,
			RoundID:                    liveconfig.RoundID,
			RepoRoot:                   repoRoot,
			DatasetMaterializeCacheDir: cfg.MaterializeDir,
		}, c.EvaluateAttempts)
	case liveconfig.ModeStabilityProbe:
		return round.RunStabilityProbe(ctx, round.LiveModeInput{
			ManifestPath:               manifest,
			ArtifactRoot:               artifactRoot,
			BundlePath:                 bundlePath,
			RoundID:                    liveconfig.RoundID,
			RepoRoot:                   repoRoot,
			DatasetMaterializeCacheDir: cfg.MaterializeDir,
		}, c.StabilityAttempts)
	default:
		return fmt.Errorf("buck round: unknown mode %q", c.Mode)
	}
}

func (c *BuckRoundCmd) resolveRepoRoot() (string, error) {
	if root := strings.TrimSpace(c.RepoRoot); root != "" {
		return root, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "flake.nix")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return wd, nil
}

// BuckValidateBundleCmd validates a completed bundle (Buck alias).
type BuckValidateBundleCmd struct {
	BundlePath string `required:"" help:"Path to completed round bundle." type:"path"`
}

// Run validates bundle artifacts.
func (c *BuckValidateBundleCmd) Run(ctx context.Context, app *App) error {
	_ = ctx
	canonical, err := validateBundle(c.BundlePath)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(app.stdout(), "ok bundle=%s passed=%v mode=%s\n", c.BundlePath, canonical.Passed, canonical.Mode)
	return err
}

// BuckDatasetCmd groups Buck dataset commands.
type BuckDatasetCmd struct {
	MaterializeLCA BuckDatasetMaterializeLCACmd `cmd:"materialize-lca" help:"Export LCA slice from Hugging Face."`
}

// BuckDatasetMaterializeLCACmd exports LCA JSONL for Buck materialize_dataset.
type BuckDatasetMaterializeLCACmd struct {
	ManifestDir string `required:"" help:"Round manifest directory." type:"path"`
	Config      string `default:"py" help:"HF dataset config."`
	Split       string `default:"dev" help:"HF dataset split."`
	MaxItems    int    `default:"1" help:"Maximum rows to export."`
	Skip        int    `default:"50" help:"HF rows to skip before export."`
}

// Run exports the configured LCA slice.
func (c *BuckDatasetMaterializeLCACmd) Run(ctx context.Context, app *App) error {
	path, err := lca.ExportFromHuggingFace(ctx, lca.ExportOptions{
		ManifestDir: c.ManifestDir,
		Config:      c.Config,
		Split:       c.Split,
		MaxItems:    c.MaxItems,
		Skip:        c.Skip,
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(app.stdout(), "dataset=%s\n", path)
	return err
}
