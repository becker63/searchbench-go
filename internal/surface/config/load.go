package config

import (
	"context"
	"fmt"

	"github.com/becker63/searchbench-go/internal/surface/config/generated"
)

// LoadFromPath resolves a Pkl experiment manifest into its typed Go config and
// then applies SearchBench-specific Go validation.
func LoadFromPath(ctx context.Context, path string) (Experiment, error) {
	experiment, err := generated.LoadFromPath(ctx, path)
	if err != nil {
		return Experiment{}, fmt.Errorf("config: load experiment: %w", err)
	}
	if err := Validate(experiment); err != nil {
		return Experiment{}, err
	}
	return experiment, nil
}
