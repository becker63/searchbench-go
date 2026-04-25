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

type SystemFingerprint string

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

func (s SystemSpec) Fingerprint() SystemFingerprint {
	var policy *PolicyRef
	if s.Policy != nil {
		ref := s.Policy.Ref()
		policy = &ref
	}

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
