package lca

import (
	"strings"

	"github.com/becker63/searchbench-go/internal/ports/dataset"
)

// JetBrainsLCADatasetName is the Hugging Face id for the LCA benchmark slice.
const JetBrainsLCADatasetName = "JetBrains-Research/lca-bug-localization"

// IsJetBrainsDataset reports whether req selects the JetBrains LCA dataset.
func IsJetBrainsDataset(r dataset.Request) bool {
	if strings.TrimSpace(r.Kind) != "lca" {
		return false
	}
	return normalizeDatasetName(r.Name) == normalizeDatasetName(JetBrainsLCADatasetName)
}

func normalizeDatasetName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func slugDatasetName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), "/", "_")
}
