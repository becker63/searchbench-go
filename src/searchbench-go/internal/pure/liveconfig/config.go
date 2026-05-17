// Package liveconfig holds repo-owned defaults for the live IC vs jCodeMunch round.
// Buck targets and tests should prefer these values over ad-hoc environment variables.
package liveconfig

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	// RoundName is the configs/rounds directory name for the live smoke round.
	RoundName = "live-ic-vs-jcodemunch"
	// RoundID is the bundle id written for live smoke runs.
	RoundID = "live-ic-vs-jcodemunch-001"
	// GameSlug is the code-localization game path segment under artifacts/.
	GameSlug = "code-localization"

	// DefaultLCAConfig is the Hugging Face dataset config for live LCA slices.
	DefaultLCAConfig = "py"
	// DefaultLCASplit is the Hugging Face dataset split for live LCA slices.
	DefaultLCASplit = "dev"
	// DefaultLCAMaxItems is the default row count exported for live runs.
	DefaultLCAMaxItems = 1
	// DefaultLCASkip is rows skipped when streaming from Hugging Face.
	DefaultLCASkip = 50

	// DefaultLiveTimeout is the default wall-clock budget for a live round.
	DefaultLiveTimeout = "45m"

	// ICBackendBuckTarget is the optimizable backend descriptor for IC.
	ICBackendBuckTarget = "//src/iterative-context:optimizable_backend"

	envCerebrasAPIKey          = "CEREBRAS_API_KEY"
	envHFToken                 = "HF_TOKEN"
	envJCodeMunchCommand       = "SEARCHBENCH_JCODEMUNCH_COMMAND"
	envIterativeContextCommand = "SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND"
	envMaterializeCacheDir     = "SEARCHBENCH_MATERIALIZE_CACHE_DIR"
	envRunLiveE2E              = "SEARCHBENCH_RUN_LIVE_E2E"
	envLiveMode                = "SEARCHBENCH_LIVE_MODE"
	envEvaluateAttempts        = "SEARCHBENCH_EVALUATE_ATTEMPTS"
	envStabilityProbeAttempts  = "SEARCHBENCH_STABILITY_PROBE_ATTEMPTS"
)

// Mode names for Buck-owned live evaluation targets.
const (
	ModeValidateBundle = "validate_bundle"
	ModeLiveSmoke      = "live_smoke"
	ModeEvaluateN      = "evaluate_n"
	ModeStabilityProbe = "stability_probe"
)

// Config is the non-secret configuration for repo-owned live SearchBench runs.
type Config struct {
	RepoRoot       string
	ManifestDir    string
	ManifestPath   string
	ArtifactRoot   string
	BundleDest     string
	MaterializeDir string
	LCAConfig      string
	LCASplit       string
	LCAMaxItems    int
	LCASkip        int
	LiveTimeout    string
}

// Default returns live round paths relative to repoRoot.
func Default(repoRoot string) Config {
	manifestDir := filepath.Join(repoRoot, "configs", "rounds", RoundName)
	return Config{
		RepoRoot:       repoRoot,
		ManifestDir:    manifestDir,
		ManifestPath:   filepath.Join(manifestDir, "round.pkl"),
		ArtifactRoot:   filepath.Join(manifestDir, "artifacts"),
		BundleDest:     filepath.Join(manifestDir, "artifacts", "games", GameSlug, "rounds", RoundID),
		MaterializeDir: filepath.Join(repoRoot, ".cache", "searchbench", "materialized-repos"),
		LCAConfig:      DefaultLCAConfig,
		LCASplit:       DefaultLCASplit,
		LCAMaxItems:    DefaultLCAMaxItems,
		LCASkip:        DefaultLCASkip,
		LiveTimeout:    DefaultLiveTimeout,
	}
}

// DefaultJCodeMunchCommand is the default MCP launcher when not overridden in .env.
func DefaultJCodeMunchCommand(repoRoot string) string {
	return "uvx jcodemunch-mcp"
}

// DefaultIterativeContextCommand is the default IC MCP launcher when not overridden in .env.
func DefaultIterativeContextCommand(repoRoot string) string {
	return "uv run --directory " + filepath.Join(repoRoot, "src", "iterative-context") + " python -m iterative_context.server"
}

// SecretEnvKeys are loaded from root .env only (never from Buck attrs).
var SecretEnvKeys = []string{envCerebrasAPIKey, envHFToken}

// OptionalEnvKeys may be set in .env to override repo defaults.
var OptionalEnvKeys = []string{
	envJCodeMunchCommand,
	envIterativeContextCommand,
	envMaterializeCacheDir,
}

// JCodeMunchCommand returns the MCP launcher, preferring env after runtime defaults.
func (c Config) JCodeMunchCommand() string {
	if v := strings.TrimSpace(os.Getenv(envJCodeMunchCommand)); v != "" {
		return v
	}
	return DefaultJCodeMunchCommand(c.RepoRoot)
}

// IterativeContextCommand returns the IC MCP launcher.
func (c Config) IterativeContextCommand() string {
	if v := strings.TrimSpace(os.Getenv(envIterativeContextCommand)); v != "" {
		return v
	}
	return DefaultIterativeContextCommand(c.RepoRoot)
}

// MaterializeCacheDir returns the dataset materialize cache directory.
func (c Config) MaterializeCacheDir() string {
	if v := strings.TrimSpace(os.Getenv(envMaterializeCacheDir)); v != "" {
		return v
	}
	return c.MaterializeDir
}

// RunLiveE2EEnvKey gates opt-in live tests (Buck/live_e2e only).
func RunLiveE2EEnvKey() string { return envRunLiveE2E }

// LiveModeFromEnv returns the Buck/test live mode selector (empty → caller default).
func LiveModeFromEnv() string {
	return strings.TrimSpace(os.Getenv(envLiveMode))
}
