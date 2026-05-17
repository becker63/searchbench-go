package liveconfig

import (
	"os"
	"path/filepath"
	"strings"
)

const devOverrideMarker = "# dev"

// LoadSecretsOnly loads CEREBRAS_API_KEY and HF_TOKEN from repo-root .env when unset.
func LoadSecretsOnly(repoRoot string) {
	loadDotEnv(repoRoot, secretKeys(), false)
}

// LoadDevOverrides loads optional MCP/cache overrides from .env only when the line
// is marked with "# dev" (developer-local overrides, not product config).
func LoadDevOverrides(repoRoot string) {
	loadDotEnv(repoRoot, OptionalEnvKeys, true)
}

// ApplyRuntimeDefaults sets non-secret runtime env from Go-owned cfg when unset.
// Secrets and dev-marked overrides must be loaded separately via LoadSecretsOnly
// and LoadDevOverrides.
func ApplyRuntimeDefaults(cfg Config) {
	if strings.TrimSpace(os.Getenv(envJCodeMunchCommand)) == "" {
		_ = os.Setenv(envJCodeMunchCommand, DefaultJCodeMunchCommand(cfg.RepoRoot))
	}
	if strings.TrimSpace(os.Getenv(envIterativeContextCommand)) == "" {
		_ = os.Setenv(envIterativeContextCommand, DefaultIterativeContextCommand(cfg.RepoRoot))
	}
}

// ApplyLiveRuntimeDefaults sets live-round runtime env including materialize cache.
// Call only from Buck-owned live modes, not from generic offline round runs.
func ApplyLiveRuntimeDefaults(cfg Config) {
	ApplyRuntimeDefaults(cfg)
	if strings.TrimSpace(os.Getenv(envMaterializeCacheDir)) == "" {
		_ = os.Setenv(envMaterializeCacheDir, cfg.MaterializeDir)
	}
}

// LoadRootDotEnv is deprecated; use LoadSecretsOnly, LoadDevOverrides, and ApplyRuntimeDefaults.
func LoadRootDotEnv(repoRoot string) {
	LoadSecretsOnly(repoRoot)
	LoadDevOverrides(repoRoot)
}

// ApplyDefaults is deprecated; use ApplyRuntimeDefaults.
func ApplyDefaults(cfg Config) {
	ApplyRuntimeDefaults(cfg)
}

func secretKeys() []string {
	return SecretEnvKeys
}

func loadDotEnv(repoRoot string, keys []string, requireDevMarker bool) {
	allowed := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		allowed[k] = struct{}{}
	}

	path := filepath.Join(repoRoot, ".env")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		raw := strings.TrimSpace(line)
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		if requireDevMarker && !strings.Contains(strings.ToLower(raw), devOverrideMarker) {
			continue
		}
		key, val, ok := strings.Cut(raw, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"`)
		if idx := strings.Index(val, devOverrideMarker); idx >= 0 {
			val = strings.TrimSpace(val[:idx])
		}
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		if _, ok := allowed[key]; !ok {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

// RedactEnv returns a copy of env suitable for logs: secret values replaced.
func RedactEnv(env []string) []string {
	secret := make(map[string]struct{}, len(SecretEnvKeys)+1)
	for _, k := range secretKeys() {
		secret[k] = struct{}{}
	}
	out := make([]string, len(env))
	for i, kv := range env {
		key, _, ok := strings.Cut(kv, "=")
		if !ok {
			out[i] = kv
			continue
		}
		if _, sensitive := secret[key]; sensitive {
			out[i] = key + "=***"
			continue
		}
		out[i] = kv
	}
	return out
}
