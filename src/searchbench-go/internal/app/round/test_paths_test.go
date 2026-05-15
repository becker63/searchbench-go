package round

import (
	"testing"

	"github.com/becker63/searchbench-go/internal/testing/reporoot"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	return reporoot.MonorepoRoot(t)
}
