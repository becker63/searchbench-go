package domain

// BackendKind identifies the execution backend used by a system.
//
// A backend is not the whole system. It is only the runtime/tool adapter.
// The full system also includes model, prompt bundle, policy, and runtime config.
type BackendKind string

const (
	BackendIterativeContext BackendKind = "iterative-context"
	BackendJCodeMunch       BackendKind = "jcodemunch"
	BackendFake             BackendKind = "fake"
)

// ModelSpec identifies the model used by a system.
type ModelSpec struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
}

// PromptBundleRef identifies the prompt bundle used by a system.
//
// Keep this as a reference rather than inline prompt text. Prompt text can live
// in files/artifacts later.
type PromptBundleRef struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// PolicyLanguage identifies the language/runtime for a policy artifact.
type PolicyLanguage string

const (
	PolicyLanguagePython PolicyLanguage = "python"
)

// PolicyHash is the stable content hash for a policy source.
type PolicyHash string

// PolicyArtifact is a replaceable policy used by a system.
//
// In the Go harness, policy source is data. For iterative-context, this can be
// passed into the Python backend through the MCP/session API instead of being
// written into a worktree.
type PolicyArtifact struct {
	ID         PolicyID       `json:"id"`
	Language   PolicyLanguage `json:"language"`
	SHA256     PolicyHash     `json:"sha256"`
	Entrypoint string         `json:"entrypoint"`
	Source     string         `json:"source,omitempty"`
}

// RuntimeConfig captures behavior-affecting runtime limits.
//
// If changing a field could change the system's output, it belongs somewhere
// under SystemSpec.
type RuntimeConfig struct {
	MaxSteps           int `json:"max_steps,omitempty"`
	MaxToolCalls       int `json:"max_tool_calls,omitempty"`
	MaxInputTokens     int `json:"max_input_tokens,omitempty"`
	MaxOutputTokens    int `json:"max_output_tokens,omitempty"`
	MaxContextTokens   int `json:"max_context_tokens,omitempty"`
	ToolResultMaxBytes int `json:"tool_result_max_bytes,omitempty"`
}

// SystemSpec is the complete recipe for an agentic code-search system.
//
// Baseline and candidate are not different types. They are roles that a
// SystemSpec can occupy inside a comparison.
type SystemSpec struct {
	ID           SystemID        `json:"id"`
	Name         string          `json:"name"`
	Backend      BackendKind     `json:"backend"`
	Model        ModelSpec       `json:"model"`
	PromptBundle PromptBundleRef `json:"prompt_bundle"`
	Policy       *PolicyArtifact `json:"policy,omitempty"`
	Runtime      RuntimeConfig   `json:"runtime,omitempty"`
}
