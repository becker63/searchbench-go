package localrun

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	artifact "github.com/becker63/searchbench-go/internal/adapters/artifact/fsbundle"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

type materializedEvidence struct {
	CurrentRef       score.ObjectiveEvidenceRef
	CurrentScorePath string
	ParentRef        *score.ObjectiveEvidenceRef
	ParentScorePath  string
	cleanup          func()
}

func materializeScoreEvidence(plan resolvedPlan, current score.ScoreEvidenceDocument) (materializedEvidence, error) {
	data, err := artifact.MarshalScoreEvidencePKL(current)
	if err != nil {
		return materializedEvidence{}, err
	}

	dir, err := os.MkdirTemp("", "searchbench-localrun-score-*")
	if err != nil {
		return materializedEvidence{}, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }

	currentScorePath := filepath.Join(dir, "score.pkl")
	if err := os.WriteFile(currentScorePath, data, 0o644); err != nil {
		cleanup()
		return materializedEvidence{}, err
	}

	out := materializedEvidence{
		CurrentRef:       plan.resolvedInput.Scoring.Evidence.Current,
		CurrentScorePath: currentScorePath,
		ParentRef:        cloneEvidenceRef(plan.parentRef),
		ParentScorePath:  strings.TrimSpace(plan.parentScorePath),
		cleanup:          cleanup,
	}
	if out.ParentRef != nil {
		parentScorePath, err := resolveParentScorePath(out.ParentRef, out.ParentScorePath)
		if err != nil {
			cleanup()
			return materializedEvidence{}, err
		}
		out.ParentScorePath = parentScorePath
	}
	return out, nil
}

func (m materializedEvidence) Cleanup() {
	if m.cleanup != nil {
		m.cleanup()
	}
}

func resolveParentScorePath(ref *score.ObjectiveEvidenceRef, override string) (string, error) {
	if ref == nil {
		return "", nil
	}
	path := strings.TrimSpace(override)
	if path == "" {
		path = strings.TrimSpace(ref.ScorePath)
	}
	if path == "" {
		return "", errors.New("parent score path is required when parent evidence is supplied")
	}
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("parent score path: %w", err)
	}
	return path, nil
}
