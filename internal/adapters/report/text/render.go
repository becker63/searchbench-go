// Package text adapts surface/console rendering for app-layer consumers
// without dragging the surface layer into app/round.
package text

import (
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/surface/console"
)

// Options configures rendering.
type Options struct {
	Color bool
	Width int
}

// RenderRoundReport renders a round report as plain text suitable for
// embedding in a durable bundle.
func RenderRoundReport(r report.RoundReport, opts Options) string {
	return console.RenderRoundReport(r, console.Options{
		Color: opts.Color,
		Width: opts.Width,
	})
}
