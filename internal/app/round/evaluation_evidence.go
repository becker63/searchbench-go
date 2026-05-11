package round

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

type materializedEvidence struct {
	CurrentRef          score.ObjectiveEvidenceRef
	CurrentEvidencePath string
	ParentRef           *score.ObjectiveEvidenceRef
	ParentEvidencePath  string
	cleanup             func()
}

func materializeRoundEvidence(plan Plan, current score.RoundEvidenceDocument) (materializedEvidence, error) {
	data, err := bundlefs.MarshalRoundEvidencePKL(current)
	if err != nil {
		return materializedEvidence{}, err
	}

	dir, err := os.MkdirTemp("", "searchbench-round-evidence-*")
	if err != nil {
		return materializedEvidence{}, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }

	currentEvidencePath := filepath.Join(dir, "evidence.pkl")
	if err := os.WriteFile(currentEvidencePath, data, 0o644); err != nil {
		cleanup()
		return materializedEvidence{}, err
	}

	out := materializedEvidence{
		CurrentRef:          plan.Scoring.CurrentEvidence,
		CurrentEvidencePath: currentEvidencePath,
		ParentRef:           cloneEvidenceRef(plan.Scoring.ParentEvidence),
		ParentEvidencePath:  strings.TrimSpace(plan.Scoring.ParentEvidencePath),
		cleanup:             cleanup,
	}
	if out.ParentRef != nil {
		parentEvidencePath, err := resolveParentEvidencePath(out.ParentRef, out.ParentEvidencePath)
		if err != nil {
			cleanup()
			return materializedEvidence{}, err
		}
		out.ParentEvidencePath = parentEvidencePath
	}
	return out, nil
}

func (m materializedEvidence) Cleanup() {
	if m.cleanup != nil {
		m.cleanup()
	}
}

func resolveParentEvidencePath(ref *score.ObjectiveEvidenceRef, override string) (string, error) {
	if ref == nil {
		return "", nil
	}
	path := strings.TrimSpace(override)
	if path == "" {
		path = strings.TrimSpace(ref.EvidencePath)
	}
	if path == "" {
		return "", errors.New("parent evidence path is required when parent evidence is supplied")
	}
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("parent evidence path: %w", err)
	}
	return path, nil
}
