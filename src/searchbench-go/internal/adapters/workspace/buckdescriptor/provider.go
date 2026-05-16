package buckdescriptor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/becker63/searchbench-go/internal/adapters/workspace/materialize"
	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// Provider resolves a Buck-backed optimizable backend descriptor into a WorkspaceSeed.
type Provider struct {
	// DescriptorPath is explicit JSON path; when empty, DescriptorTarget or default discovery is used.
	DescriptorPath string
	// DescriptorTarget is the Buck label (default //src/iterative-context:optimizable_backend).
	DescriptorTarget string
	// RepoRoot anchors relative source.path values.
	RepoRoot string
	// RunBuck when true runs buck2 build on DescriptorTarget before loading output.
	RunBuck bool
}

// PrepareSeed implements workspace.SeedProvider.
func (p Provider) PrepareSeed(ctx context.Context) (optimizer.WorkspaceSeed, error) {
	path, err := p.resolveDescriptorPath(ctx)
	if err != nil {
		return optimizer.WorkspaceSeed{}, err
	}
	desc, err := LoadDescriptorFile(path)
	if err != nil {
		return optimizer.WorkspaceSeed{}, err
	}
	return p.seedFromDescriptor(desc)
}

func (p Provider) resolveDescriptorPath(ctx context.Context) (string, error) {
	if path := strings.TrimSpace(p.DescriptorPath); path != "" {
		return filepath.Abs(path)
	}
	if p.RunBuck {
		target := strings.TrimSpace(p.DescriptorTarget)
		if target == "" {
			target = "//src/iterative-context:optimizable_backend"
		}
		cmd := exec.CommandContext(ctx, "buck2", "build", target)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("buck2 build %s: %w: %s", target, err, strings.TrimSpace(string(out)))
		}
	}
	repoRoot, err := p.repoRoot()
	if err != nil {
		return "", err
	}
	defaultPath := filepath.Join(repoRoot, "src", "iterative-context", "optimizable_backend.json")
	if st, statErr := os.Stat(defaultPath); statErr == nil && !st.IsDir() {
		return defaultPath, nil
	}
	return "", fmt.Errorf("descriptor not found (set DescriptorPath or build %s)", p.DescriptorTarget)
}

func (p Provider) repoRoot() (string, error) {
	if root := strings.TrimSpace(p.RepoRoot); root != "" {
		return filepath.Abs(root)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for d := wd; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "flake.nix")); err == nil {
			return d, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
	}
	return wd, nil
}

func (p Provider) seedFromDescriptor(desc Descriptor) (optimizer.WorkspaceSeed, error) {
	repoRoot, err := p.repoRoot()
	if err != nil {
		return optimizer.WorkspaceSeed{}, err
	}
	switch desc.Source.Kind {
	case "local_path":
		srcPath := desc.Source.Path
		if !filepath.IsAbs(srcPath) {
			srcPath = filepath.Join(repoRoot, srcPath)
		}
		abs, err := filepath.Abs(srcPath)
		if err != nil {
			return optimizer.WorkspaceSeed{}, err
		}
		digest, err := materialize.DigestTree(abs)
		if err != nil {
			return optimizer.WorkspaceSeed{}, fmt.Errorf("digest descriptor source: %w", err)
		}
		declared := strings.TrimSpace(desc.Source.DeclaredBy)
		if declared == "" {
			declared = strings.TrimSpace(p.DescriptorTarget)
		}
		seed := optimizer.WorkspaceSeed{
			ID:   "buck-descriptor-" + digest[:12],
			Kind: optimizer.SeedKindBuckDescriptor,
			Root: abs,
			Identity: optimizer.WorkspaceSeedIdentity{
				Provider: optimizer.SeedProviderBuckDescriptor,
				Source:   declared,
				Sha256:   digest,
			},
		}
		return seed, seed.Validate()
	default:
		return optimizer.WorkspaceSeed{}, fmt.Errorf("unsupported descriptor source.kind %q", desc.Source.Kind)
	}
}
