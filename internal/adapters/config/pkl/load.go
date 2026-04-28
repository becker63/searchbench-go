package config

import (
	"context"
	"fmt"

	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated"
)

// ResolveFromPath resolves a Pkl experiment manifest into its typed Go config
// without applying SearchBench-specific Go validation yet.
func ResolveFromPath(ctx context.Context, path string) (Experiment, error) {
	experiment, err := generated.LoadFromPath(ctx, path)
	if err != nil {
		return Experiment{}, fmt.Errorf("config: load experiment: %w", err)
	}
	return experiment, nil
}

// LoadFromPath resolves a Pkl experiment manifest into its typed Go config and
// then applies SearchBench-specific Go validation.
func LoadFromPath(ctx context.Context, path string) (Experiment, error) {
	experiment, err := ResolveFromPath(ctx, path)
	if err != nil {
		return Experiment{}, err
	}
	if err := Validate(experiment); err != nil {
		return Experiment{}, err
	}
	return experiment, nil
}
