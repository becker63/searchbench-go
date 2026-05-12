//go:build !cgo

package treesitter

import (
	"github.com/becker63/searchbench-go/internal/pure/codegraph"
)

// BuildDirectory is unavailable without CGO.
func BuildDirectory(absRepoRoot string) (*codegraph.Store, error) {
	return nil, ErrCGODisabled
}
