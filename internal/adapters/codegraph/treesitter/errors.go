package treesitter

import "errors"

// ErrCGODisabled is returned by [BuildDirectory] when the package was compiled
// without CGO (stub implementation).
var ErrCGODisabled = errors.New("codegraph/treesitter: tree-sitter indexer requires CGO_ENABLED=1")
