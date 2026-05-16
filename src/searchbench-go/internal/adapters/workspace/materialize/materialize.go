package materialize

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// Options control candidate workspace materialization.
type Options struct {
	// Keep when true skips cleanup (debug); Cleanup becomes a no-op.
	Keep bool
}

// CandidateMaterializer copies a WorkspaceSeed into a mutable IC candidate workspace.
type CandidateMaterializer struct {
	TempPrefix string
	Opts       Options
}

// Materialize copies seed-backed source into a fresh candidate workspace.
func (m *CandidateMaterializer) Materialize(seed optimizer.WorkspaceSeed) (optimizer.ICCandidateWorkspace, func() error, error) {
	if err := seed.Validate(); err != nil {
		return optimizer.ICCandidateWorkspace{}, nil, fmt.Errorf("seed: %w", err)
	}
	switch seed.Kind {
	case optimizer.SeedKindLocalPath, optimizer.SeedKindBuckDescriptor:
		// buck_descriptor Level-2 seeds still point at local_path roots.
	default:
		return optimizer.ICCandidateWorkspace{}, nil, fmt.Errorf("unsupported seed kind %q", seed.Kind)
	}
	src := strings.TrimSpace(seed.Root)
	if src == "" {
		return optimizer.ICCandidateWorkspace{}, nil, fmt.Errorf("seed root path is required")
	}
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return optimizer.ICCandidateWorkspace{}, nil, err
	}
	prefix := m.TempPrefix
	if prefix == "" {
		prefix = "searchbench-ic-candidate-"
	}
	dest, err := os.MkdirTemp("", prefix)
	if err != nil {
		return optimizer.ICCandidateWorkspace{}, nil, fmt.Errorf("mkdir candidate workspace: %w", err)
	}
	if err := copyTree(srcAbs, dest); err != nil {
		_ = os.RemoveAll(dest)
		return optimizer.ICCandidateWorkspace{}, nil, err
	}
	ws := optimizer.ICCandidateWorkspace{
		ID:   seed.ID,
		Root: dest,
		Seed: seed.Identity,
	}
	cleanup := func() error {
		if m.Opts.Keep {
			return nil
		}
		return os.RemoveAll(dest)
	}
	return ws, cleanup, nil
}

func copyTree(srcRoot, destRoot string) error {
	return filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		relSlash := filepath.ToSlash(rel)
		if ShouldExcludePath(relSlash) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		dest := filepath.Join(destRoot, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return copyFile(path, dest)
	})
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
