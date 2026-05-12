//go:build !cgo

package treesitter

import (
	"errors"

	"github.com/becker63/searchbench-go/internal/pure/codegraph"
)

// ErrCGODisabled indicates the tree-sitter indexer was compiled without CGO.
var ErrCGODisabled = errors.New("codegraph/treesitter: tree-sitter indexer requires CGO_ENABLED=1")

// BuildDirectory is unavailable without CGO.
func BuildDirectory(absRepoRoot string) (*codegraph.Store, error) {
	return nil, ErrCGODisabled
}
