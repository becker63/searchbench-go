package liveconfig

import (
	"path/filepath"
	"strings"
)

// ConfigFromManifest returns live config when manifestPath belongs to the live round.
func ConfigFromManifest(manifestPath string) (Config, bool) {
	manifestPath = filepath.Clean(manifestPath)
	if !strings.Contains(manifestPath, RoundName) {
		return Config{}, false
	}
	manifestDir := filepath.Dir(manifestPath)
	repoRoot := filepath.Clean(filepath.Join(manifestDir, "..", "..", ".."))
	cfg := Default(repoRoot)
	cfg.ManifestDir = manifestDir
	cfg.ManifestPath = manifestPath
	return cfg, true
}
