package round

import (
	"path/filepath"

	"github.com/becker63/searchbench-go/internal/pure/liveconfig"
)

func applyRuntimeDefaultsForPlan(plan Plan) {
	if plan.ManifestPath == "" {
		return
	}
	cfg, ok := liveconfig.ConfigFromManifest(plan.ManifestPath)
	if !ok {
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(plan.ManifestPath), "..", "..", ".."))
		cfg = liveconfig.Default(repoRoot)
	}
	liveconfig.ApplyRuntimeDefaults(cfg)
}
