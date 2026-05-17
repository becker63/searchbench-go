package liveconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// InputFingerprint hashes pinned manifest and dataset paths for live reports (#86).
func InputFingerprint(cfg Config) (string, error) {
	h := sha256.New()
	_, _ = h.Write([]byte("manifest\x00"))
	if data, err := os.ReadFile(cfg.ManifestPath); err == nil {
		_, _ = h.Write(data)
	}
	_, _ = h.Write([]byte("\ndataset\x00"))
	slice := DatasetSlicePath(cfg)
	if data, err := os.ReadFile(slice); err == nil {
		_, _ = h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// DatasetSlicePath returns the default LCA JSONL path for the live round.
func DatasetSlicePath(cfg Config) string {
	return fmt.Sprintf("%s/datasets/JetBrains-Research_lca-bug-localization/%s/%s.jsonl",
		cfg.ManifestDir,
		cfg.LCAConfig,
		cfg.LCASplit,
	)
}

// ModelSeed derives a stable seed string for reports (#87).
func ModelSeed(roundID, matchID, role string, attempt int) string {
	raw := fmt.Sprintf("%s|%s|%s|%d", strings.TrimSpace(roundID), strings.TrimSpace(matchID), strings.TrimSpace(role), attempt)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:8])
}
