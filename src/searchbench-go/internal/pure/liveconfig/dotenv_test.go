package liveconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSecretsOnlyDoesNotLoadNonSecrets(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "CEREBRAS_API_KEY=secret-value\n" +
		"HF_TOKEN=hf-secret\n" +
		"SEARCHBENCH_JCODEMUNCH_COMMAND=should-not-load\n" +
		"RANDOM_CONFIG=also-skip\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CEREBRAS_API_KEY", "")
	t.Setenv("HF_TOKEN", "")
	t.Setenv("SEARCHBENCH_JCODEMUNCH_COMMAND", "")
	t.Setenv("RANDOM_CONFIG", "")

	LoadSecretsOnly(dir)

	if got := os.Getenv("CEREBRAS_API_KEY"); got != "secret-value" {
		t.Fatalf("CEREBRAS_API_KEY = %q", got)
	}
	if got := os.Getenv("HF_TOKEN"); got != "hf-secret" {
		t.Fatalf("HF_TOKEN = %q", got)
	}
	if got := os.Getenv("SEARCHBENCH_JCODEMUNCH_COMMAND"); got != "" {
		t.Fatalf("non-secret leaked into env: %q", got)
	}
	if got := os.Getenv("RANDOM_CONFIG"); got != "" {
		t.Fatalf("unlisted key leaked: %q", got)
	}
}

func TestLoadDevOverridesRequiresMarker(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "SEARCHBENCH_JCODEMUNCH_COMMAND=blocked\n" +
		"SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND=allowed # dev\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SEARCHBENCH_JCODEMUNCH_COMMAND", "")
	t.Setenv("SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND", "")

	LoadDevOverrides(dir)

	if got := os.Getenv("SEARCHBENCH_JCODEMUNCH_COMMAND"); got != "" {
		t.Fatalf("unmarked override loaded: %q", got)
	}
	if got := os.Getenv("SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND"); got != "allowed" {
		t.Fatalf("dev override = %q, want allowed", got)
	}
}

func TestRedactEnvMasksSecrets(t *testing.T) {
	t.Parallel()
	redacted := RedactEnv([]string{
		"CEREBRAS_API_KEY=abc",
		"HF_TOKEN=def",
		"PATH=/usr/bin",
	})
	if redacted[0] != "CEREBRAS_API_KEY=***" || redacted[1] != "HF_TOKEN=***" {
		t.Fatalf("redacted=%v", redacted)
	}
	if redacted[2] != "PATH=/usr/bin" {
		t.Fatalf("non-secret changed: %q", redacted[2])
	}
}
