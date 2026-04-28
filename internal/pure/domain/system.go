package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

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

// PolicyRef is the report-safe identity for a policy artifact.
//
// It intentionally omits policy source so reports and comparisons can refer to
// the policy without embedding executable code.
type PolicyRef struct {
	ID         PolicyID       `json:"id"`
	Language   PolicyLanguage `json:"language"`
	SHA256     PolicyHash     `json:"sha256"`
	Entrypoint string         `json:"entrypoint"`
}

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

func NewPythonPolicy(id PolicyID, source string, entrypoint string) PolicyArtifact {
	return PolicyArtifact{
		ID:         id,
		Language:   PolicyLanguagePython,
		SHA256:     hashPolicySource(source),
		Entrypoint: entrypoint,
		Source:     source,
	}
}

// Validate checks that the policy is executable and self-consistent.
//
// In particular, Validate ensures the stored content hash matches Source so the
// policy cannot silently drift away from its identity.
func (p PolicyArtifact) Validate() error {
	if p.ID.Empty() {
		return errors.New("policy id is required")
	}
	switch p.Language {
	case PolicyLanguagePython:
	default:
		return fmt.Errorf("unsupported policy language: %q", p.Language)
	}
	if p.Entrypoint == "" {
		return errors.New("policy entrypoint is required")
	}
	if p.Source == "" {
		return errors.New("policy source is required")
	}
	if p.SHA256 == "" {
		return errors.New("policy sha256 is required")
	}
	expected := hashPolicySource(p.Source)
	if p.SHA256 != expected {
		return fmt.Errorf("policy sha256 does not match source: got %q want %q", p.SHA256, expected)
	}
	return nil
}

// Ref returns the report-safe identity for the policy without source text.
func (p PolicyArtifact) Ref() PolicyRef {
	return PolicyRef{
		ID:         p.ID,
		Language:   p.Language,
		SHA256:     p.SHA256,
		Entrypoint: p.Entrypoint,
	}
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

// SystemFingerprint is the stable hash of a system's behavior-affecting
// configuration.
type SystemFingerprint string

// SystemRef is the report-safe identity for an executable system.
//
// It contains all fields needed to explain what was compared, but it omits
// executable-only data such as policy source.
type SystemRef struct {
	ID           SystemID          `json:"id"`
	Name         string            `json:"name"`
	Backend      BackendKind       `json:"backend"`
	Model        ModelSpec         `json:"model"`
	PromptBundle PromptBundleRef   `json:"prompt_bundle"`
	Policy       *PolicyRef        `json:"policy,omitempty"`
	Runtime      RuntimeConfig     `json:"runtime,omitempty"`
	Fingerprint  SystemFingerprint `json:"fingerprint"`
}

// Validate checks that the executable system recipe is usable.
func (s SystemSpec) Validate() error {
	if s.ID.Empty() {
		return errors.New("system id is required")
	}
	switch s.Backend {
	case BackendIterativeContext, BackendJCodeMunch, BackendFake:
	default:
		return fmt.Errorf("unsupported backend: %q", s.Backend)
	}
	if s.Model.Provider == "" {
		return errors.New("model provider is required")
	}
	if s.Model.Name == "" {
		return errors.New("model name is required")
	}
	if s.PromptBundle.Name == "" {
		return errors.New("prompt bundle name is required")
	}
	if s.Policy != nil {
		if err := s.Policy.Validate(); err != nil {
			return fmt.Errorf("policy: %w", err)
		}
	}
	return nil
}

// Fingerprint returns a deterministic hash over the system's
// behavior-affecting configuration.
//
// Cosmetic display fields such as Name are intentionally excluded, while
// PolicyRef is included so language, entrypoint, and hash all affect identity.
func (s SystemSpec) Fingerprint() SystemFingerprint {
	var policy *PolicyRef
	if s.Policy != nil {
		ref := s.Policy.Ref()
		policy = &ref
	}

	// Name is intentionally excluded because it is cosmetic; prompt/model/policy
	// selection and runtime limits are the fields that affect behavior.
	canonical := struct {
		Backend      BackendKind     `json:"backend"`
		Model        ModelSpec       `json:"model"`
		PromptBundle PromptBundleRef `json:"prompt_bundle"`
		Policy       *PolicyRef      `json:"policy,omitempty"`
		Runtime      RuntimeConfig   `json:"runtime,omitempty"`
	}{
		Backend:      s.Backend,
		Model:        s.Model,
		PromptBundle: s.PromptBundle,
		Policy:       policy,
		Runtime:      s.Runtime,
	}

	data, err := json.Marshal(canonical)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return SystemFingerprint(hex.EncodeToString(sum[:]))
}

// Ref converts the executable system recipe into its report-safe identity.
//
// Reports should prefer SystemRef so policy source does not leak by default.
func (s SystemSpec) Ref() SystemRef {
	var policy *PolicyRef
	if s.Policy != nil {
		ref := s.Policy.Ref()
		policy = &ref
	}

	return SystemRef{
		ID:           s.ID,
		Name:         s.Name,
		Backend:      s.Backend,
		Model:        s.Model,
		PromptBundle: s.PromptBundle,
		Policy:       policy,
		Runtime:      s.Runtime,
		Fingerprint:  s.Fingerprint(),
	}
}

func hashPolicySource(source string) PolicyHash {
	sum := sha256.Sum256([]byte(source))
	return PolicyHash(hex.EncodeToString(sum[:]))
}
