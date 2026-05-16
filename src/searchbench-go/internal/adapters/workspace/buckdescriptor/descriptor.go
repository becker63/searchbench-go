package buckdescriptor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const KindOptimizableBackendV1 = "searchbench.optimizable_backend.v1"

// Descriptor is the candidate-facing SearchBench ↔ IC backend contract (not repo checks).
type Descriptor struct {
	Kind               string             `json:"kind"`
	Source             Source             `json:"source"`
	Launcher           Launcher           `json:"launcher"`
	CandidateValidator CandidateValidator `json:"candidate_validator"`
	RuntimeAdmin       RuntimeAdmin       `json:"runtime_admin"`
	RepoChecks         map[string]string  `json:"repo_checks,omitempty"`
}

type Source struct {
	Kind       string `json:"kind"`
	Path       string `json:"path"`
	DeclaredBy string `json:"declared_by,omitempty"`
}

type Launcher struct {
	Kind    string            `json:"kind"`
	CWDMode string            `json:"cwd_mode"`
	Argv    []string          `json:"argv"`
	Env     map[string]string `json:"env"`
}

type CandidateValidator struct {
	Kind  string   `json:"kind"`
	Steps []string `json:"steps"`
}

type RuntimeAdmin struct {
	InstallTool         string `json:"install_tool"`
	VerifyTool          string `json:"verify_tool"`
	HiddenFromEvaluator bool   `json:"hidden_from_evaluator"`
}

// ParseDescriptorJSON loads and validates a backend descriptor.
func ParseDescriptorJSON(data []byte) (Descriptor, error) {
	var d Descriptor
	if err := json.Unmarshal(data, &d); err != nil {
		return Descriptor{}, fmt.Errorf("parse descriptor json: %w", err)
	}
	return d, ValidateDescriptor(d)
}

// LoadDescriptorFile reads descriptor JSON from path.
func LoadDescriptorFile(path string) (Descriptor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Descriptor{}, err
	}
	return ParseDescriptorJSON(data)
}

// ValidateDescriptor rejects unknown kinds, malformed fields, and repo_checks.
func ValidateDescriptor(d Descriptor) error {
	if strings.TrimSpace(d.Kind) != KindOptimizableBackendV1 {
		return fmt.Errorf("unsupported descriptor kind %q", d.Kind)
	}
	if d.RepoChecks != nil && len(d.RepoChecks) > 0 {
		return fmt.Errorf("descriptor must not include repo_checks")
	}
	sk := strings.TrimSpace(d.Source.Kind)
	if sk == "" {
		return fmt.Errorf("source.kind is required")
	}
	if sk != "local_path" && sk != "archive" && sk != "git" {
		return fmt.Errorf("unsupported source.kind %q", sk)
	}
	if sk == "local_path" && strings.TrimSpace(d.Source.Path) == "" {
		return fmt.Errorf("source.path is required for local_path")
	}
	if strings.TrimSpace(d.Launcher.Kind) == "" {
		return fmt.Errorf("launcher.kind is required")
	}
	if len(d.Launcher.Argv) == 0 {
		return fmt.Errorf("launcher.argv is required")
	}
	if strings.TrimSpace(d.CandidateValidator.Kind) == "" {
		return fmt.Errorf("candidate_validator.kind is required")
	}
	return nil
}
