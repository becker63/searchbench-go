package eino

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/becker63/searchbench-go/internal/domain"
)

type finalPrediction struct {
	PredictedFiles []string `json:"predicted_files"`
	Files          []string `json:"files"`
	Reasoning      string   `json:"reasoning"`
}

// FinalizePrediction parses and normalizes the evaluator's final JSON output.
func FinalizePrediction(raw string) (domain.Prediction, FailureKind, error) {
	var payload finalPrediction
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &payload); err != nil {
		return domain.Prediction{}, FailureKindFinalizationFailed, fmt.Errorf("parse final prediction JSON: %w", err)
	}

	files := payload.PredictedFiles
	if len(files) == 0 {
		files = payload.Files
	}

	normalized := domain.CanonicalizePaths(files)
	if len(normalized) == 0 {
		return domain.Prediction{}, FailureKindInvalidPrediction, errors.New("predicted files are required")
	}

	return domain.Prediction{
		Files:     normalized,
		Reasoning: strings.TrimSpace(payload.Reasoning),
	}, "", nil
}
