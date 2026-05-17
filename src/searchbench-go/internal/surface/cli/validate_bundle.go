package cli

import (
	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	"github.com/becker63/searchbench-go/internal/pure/report"
)

func validateBundle(path string) (report.CanonicalReport, error) {
	return bundlefs.ValidateCompletedBundle(path)
}
