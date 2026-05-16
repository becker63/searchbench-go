package workspace

import (
	"context"

	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// SeedProvider prepares a WorkspaceSeed for candidate workspace materialization.
type SeedProvider interface {
	PrepareSeed(ctx context.Context) (optimizer.WorkspaceSeed, error)
}
