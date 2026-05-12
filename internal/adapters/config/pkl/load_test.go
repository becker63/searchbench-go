package config

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLoadLocalManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	path := filepath.Join("..", "..", "..", "..", "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")
	roundSpec, err := LoadFromPath(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	if roundSpec.Round == nil {
		t.Fatal("roundSpec.Round is nil")
	}
	if got, want := roundSpec.Round.Id, "round-001"; got != want {
		t.Fatalf("roundSpec.Round.Id = %q, want %q", got, want)
	}
	if roundSpec.Round.Continues != nil {
		t.Fatalf("roundSpec.Round.Continues = %#v, want nil for from-scratch round", roundSpec.Round.Continues)
	}
}

func TestLoadContinuationManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	path := filepath.Join("..", "..", "..", "..", "configs", "rounds", "continue-ic-from-local", "round.pkl")
	roundSpec, err := LoadFromPath(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	if roundSpec.Round == nil {
		t.Fatal("roundSpec.Round is nil")
	}
	if roundSpec.Round.Continues == nil {
		t.Fatal("roundSpec.Round.Continues is nil")
	}
	if got, want := *roundSpec.Round.Continues, "../local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001"; got != want {
		t.Fatalf("roundSpec.Round.Continues = %q, want %q", got, want)
	}
}

func TestLoadGeneratedContinuationManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	path := filepath.Join("..", "..", "..", "..", "configs", "rounds", "generate-ic-from-local", "round.pkl")
	roundSpec, err := LoadFromPath(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	if roundSpec.Round == nil || roundSpec.Round.Challenger.Generate == nil {
		t.Fatalf("roundSpec.Round = %#v, want generated challenger config", roundSpec.Round)
	}
	if got, want := roundSpec.Round.Challenger.Generate.ArtifactName, "next_challenger_policy.round-002.py"; got != want {
		t.Fatalf("artifactName = %q, want %q", got, want)
	}
}

func requirePkl(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pkl"); err != nil {
		t.Skip("pkl CLI not available on PATH")
	}
}
