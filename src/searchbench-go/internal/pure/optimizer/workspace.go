package optimizer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Workspace seed kinds supported by the IC candidate lifecycle.
const (
	SeedKindLocalPath      = "local_path"
	SeedKindBuckDescriptor = "buck_descriptor"
	SeedKindBuckManifest   = "buck_manifest"
	SeedKindBuckArchive    = "buck_archive"
)

// WorkspaceSeedProvider identity values recorded on seeds and evidence.
const (
	SeedProviderLocalPath      = "local_path"
	SeedProviderBuckDescriptor = "buck_descriptor"
)

// WorkspaceSeed is an immutable description of backend source material used to
// materialize a candidate workspace. It does not hardcode Buck or repo paths.
type WorkspaceSeed struct {
	ID       string                `json:"id"`
	Kind     string                `json:"kind"`
	Root     string                `json:"root,omitempty"`
	Manifest string                `json:"manifest,omitempty"`
	Archive  string                `json:"archive,omitempty"`
	Identity WorkspaceSeedIdentity `json:"identity"`
}

// WorkspaceSeedIdentity is stable evidence metadata for comparing seed providers.
type WorkspaceSeedIdentity struct {
	Provider string `json:"provider"`
	Source   string `json:"source"`
	Sha256   string `json:"sha256"`
}

// Validate checks required identity fields for a prepared seed.
func (s WorkspaceSeed) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("workspace seed id is required")
	}
	if strings.TrimSpace(s.Kind) == "" {
		return fmt.Errorf("workspace seed kind is required")
	}
	return s.Identity.Validate()
}

// WorkspaceSeedIdentity validates provider/source/sha256 presence.
func (i WorkspaceSeedIdentity) Validate() error {
	if strings.TrimSpace(i.Provider) == "" {
		return fmt.Errorf("workspace seed identity provider is required")
	}
	if strings.TrimSpace(i.Source) == "" {
		return fmt.Errorf("workspace seed identity source is required")
	}
	if strings.TrimSpace(i.Sha256) == "" {
		return fmt.Errorf("workspace seed identity sha256 is required")
	}
	return nil
}

// ICCandidateWorkspace is a mutable IC checkout used for validation and runtime.
// The workspace that passes validation is the workspace whose MCP server launches.
type ICCandidateWorkspace struct {
	ID   string                `json:"id"`
	Root string                `json:"root"`
	Seed WorkspaceSeedIdentity `json:"seed"`
}

// Validate checks candidate workspace identity.
func (w ICCandidateWorkspace) Validate() error {
	if strings.TrimSpace(w.ID) == "" {
		return fmt.Errorf("ic candidate workspace id is required")
	}
	if strings.TrimSpace(w.Root) == "" {
		return fmt.Errorf("ic candidate workspace root is required")
	}
	return w.Seed.Validate()
}

// ICPolicyArtifact is staged optimizer policy metadata for the IC boundary.
// It is distinct from domain.PolicyArtifact (round config vocabulary).
type ICPolicyArtifact struct {
	Path      string `json:"path"`
	PolicyID  string `json:"policy_id"`
	Symbol    string `json:"symbol"`
	Sha256    string `json:"sha256"`
	Interface string `json:"interface"`
}

// Validate checks required staged policy fields.
func (p ICPolicyArtifact) Validate() error {
	if strings.TrimSpace(p.Path) == "" {
		return fmt.Errorf("ic policy artifact path is required")
	}
	if strings.TrimSpace(p.PolicyID) == "" {
		return fmt.Errorf("ic policy artifact policy_id is required")
	}
	if strings.TrimSpace(p.Sha256) == "" {
		return fmt.Errorf("ic policy artifact sha256 is required")
	}
	return nil
}

// PipelineValidationResult records candidate validation against ICCandidateWorkspace.
type PipelineValidationResult struct {
	OK    bool                 `json:"ok"`
	Steps []PipelineStepResult `json:"steps,omitempty"`
}

// PipelineStepResult is one IC policy pipeline step outcome.
type PipelineStepResult struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	CWD      string `json:"cwd,omitempty"`
	Command  string `json:"command,omitempty"`
	ExitCode int    `json:"exit_code,omitempty"`
}

// ICLaunchSpec describes how to start an IC MCP server for an accepted candidate.
type ICLaunchSpec struct {
	CWD  string            `json:"cwd"`
	Argv []string          `json:"argv"`
	Env  map[string]string `json:"env,omitempty"`
}

// Validate ensures launch cwd and argv are present.
func (s ICLaunchSpec) Validate() error {
	if strings.TrimSpace(s.CWD) == "" {
		return fmt.Errorf("ic launch spec cwd is required")
	}
	if len(s.Argv) == 0 {
		return fmt.Errorf("ic launch spec argv is required")
	}
	return nil
}

// ICPolicyInstallRequest is harness-owned score installation input.
type ICPolicyInstallRequest struct {
	WorkspaceRoot string `json:"workspace_root"`
	PolicyPath    string `json:"policy_path"`
	PolicyID      string `json:"policy_id"`
	Symbol        string `json:"symbol,omitempty"`
	Sha256        string `json:"sha256"`
	Interface     string `json:"interface,omitempty"`
}

// ICRuntimeIdentity records which candidate workspace and policy are active.
type ICRuntimeIdentity struct {
	WorkspaceRoot string                `json:"workspace_root"`
	SeedIdentity  WorkspaceSeedIdentity `json:"seed_identity"`
	PolicyID      string                `json:"policy_id"`
	PolicySha256  string                `json:"policy_sha256"`
	ActiveScoreID string                `json:"active_score_id,omitempty"`
	Verified      bool                  `json:"verified"`
}

// AcceptedICCandidate binds a validated workspace, policy, pipeline, and launch spec.
type AcceptedICCandidate struct {
	Workspace  ICCandidateWorkspace     `json:"workspace"`
	Policy     ICPolicyArtifact         `json:"policy"`
	Validation PipelineValidationResult `json:"validation"`
	Launch     ICLaunchSpec             `json:"launch"`
}

// Validate checks the accepted candidate lifecycle bundle.
func (a AcceptedICCandidate) Validate() error {
	if err := a.Workspace.Validate(); err != nil {
		return fmt.Errorf("workspace: %w", err)
	}
	if err := a.Policy.Validate(); err != nil {
		return fmt.Errorf("policy: %w", err)
	}
	if err := a.Launch.Validate(); err != nil {
		return fmt.Errorf("launch: %w", err)
	}
	if strings.TrimSpace(a.Launch.CWD) != strings.TrimSpace(a.Workspace.Root) {
		return fmt.Errorf("launch cwd must equal accepted candidate workspace root")
	}
	if !a.Validation.OK {
		return fmt.Errorf("validation must be ok for accepted candidate")
	}
	return nil
}

// ICPolicyArtifactFromStagedMeta maps staged policy.json-style metadata into ICPolicyArtifact.
func ICPolicyArtifactFromStagedMeta(path, policyID, symbol, interfaceID, code string) ICPolicyArtifact {
	sum := sha256.Sum256([]byte(code))
	return ICPolicyArtifact{
		Path:      path,
		PolicyID:  policyID,
		Symbol:    symbol,
		Sha256:    hex.EncodeToString(sum[:]),
		Interface: interfaceID,
	}
}
