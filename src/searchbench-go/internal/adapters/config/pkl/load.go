package config

import (
	"context"
	"fmt"

	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated"
)

// ResolveFromPath resolves a Pkl round manifest into its typed Go config
// without applying SearchBench-specific Go validation yet.
func ResolveFromPath(ctx context.Context, path string) (RoundSpec, error) {
	round, err := generated.LoadFromPath(ctx, path)
	if err != nil {
		return RoundSpec{}, fmt.Errorf("config: load round manifest: %w", err)
	}
	return round, nil
}

// LoadFromPath resolves a Pkl round manifest into its typed Go config and
// then applies SearchBench-specific Go validation.
func LoadFromPath(ctx context.Context, path string) (RoundSpec, error) {
	round, err := ResolveFromPath(ctx, path)
	if err != nil {
		return RoundSpec{}, err
	}
	if err := Validate(round); err != nil {
		return RoundSpec{}, err
	}
	return round, nil
}
