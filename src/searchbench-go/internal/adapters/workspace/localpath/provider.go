package localpath

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/becker63/searchbench-go/internal/adapters/workspace/materialize"
	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// Provider prepares local_path workspace seeds from a filesystem source tree.
type Provider struct {
	// Source is the IC project root (for example src/iterative-context).
	Source string
	// SeedID overrides the generated seed id when non-empty.
	SeedID string
}

// PrepareSeed implements workspace.SeedProvider.
func (p Provider) PrepareSeed(ctx context.Context) (optimizer.WorkspaceSeed, error) {
	_ = ctx
	src := strings.TrimSpace(p.Source)
	if src == "" {
		return optimizer.WorkspaceSeed{}, fmt.Errorf("local path provider source is required")
	}
	abs, err := filepath.Abs(src)
	if err != nil {
		return optimizer.WorkspaceSeed{}, err
	}
	digest, err := materialize.DigestTree(abs)
	if err != nil {
		return optimizer.WorkspaceSeed{}, fmt.Errorf("digest source tree: %w", err)
	}
	id := strings.TrimSpace(p.SeedID)
	if id == "" {
		id = "local-path-" + digest[:12]
	}
	seed := optimizer.WorkspaceSeed{
		ID:   id,
		Kind: optimizer.SeedKindLocalPath,
		Root: abs,
		Identity: optimizer.WorkspaceSeedIdentity{
			Provider: optimizer.SeedProviderLocalPath,
			Source:   abs,
			Sha256:   digest,
		},
	}
	return seed, seed.Validate()
}
