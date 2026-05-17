package cli

import (
	"context"
	"fmt"

	"github.com/becker63/searchbench-go/internal/adapters/dataset/lca"
)

// DatasetCmd groups dataset materialization commands.
type DatasetCmd struct {
	Materialize MaterializeCmd `cmd:"materialize" help:"Materialize benchmark datasets."`
}

// MaterializeCmd materializes datasets for a round manifest directory.
type MaterializeCmd struct {
	LCA MaterializeLCACmd `cmd:"lca" help:"Export JetBrains LCA slice from Hugging Face."`
}

// MaterializeLCACmd exports LCA JSONL under the manifest datasets tree (#79).
type MaterializeLCACmd struct {
	ManifestDir string `required:"" help:"Round manifest directory (configs/rounds/<name>)." type:"path"`
	Config      string `default:"py" help:"HF dataset config."`
	Split       string `default:"dev" help:"HF dataset split."`
	MaxItems    int    `default:"1" help:"Maximum rows to export."`
	Skip        int    `default:"50" help:"HF rows to skip before export."`
}

// Run exports the configured LCA slice.
func (c *MaterializeLCACmd) Run(ctx context.Context, app *App) error {
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
