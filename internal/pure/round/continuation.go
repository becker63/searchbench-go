package round

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

const ContinuationSchemaVersion = "searchbench.continuation.v1"

var (
	ErrInvalidContinuation      = errors.New("round: invalid continuation")
	ErrMissingContinuationGame  = errors.New("round: missing continuation game id")
	ErrMissingContinuationRound = errors.New("round: missing continuation round id")
	ErrMissingContinuationRole  = errors.New("round: missing surviving candidate role")
	ErrMissingContinuationPath  = errors.New("round: missing surviving candidate artifact path")
	ErrMissingContinuationIface = errors.New("round: missing candidate interface id")
)

// Continuation is the durable lineage payload written into a completed round
// bundle. It carries the resolved context required to author the next round
// without repeating the inner evaluation shape in Pkl.
type Continuation struct {
	SchemaVersion       string                            `json:"schema_version"`
	BundleID            string                            `json:"bundle_id"`
	Game                ContinuationGame                  `json:"game"`
	Round               ContinuationRound                 `json:"round"`
	CandidateInterface  ContinuationInterface             `json:"candidate_interface"`
	SurvivingCandidate  ContinuationCandidate             `json:"surviving_candidate"`
	DefaultContinuation ContinuationDefaults              `json:"default_continuation"`
	Matches             domain.NonEmpty[domain.MatchSpec] `json:"matches"`
	Evaluator           ContinuationEvaluator             `json:"evaluator"`
	Scoring             ContinuationScoring               `json:"scoring"`
	Evidence            ContinuationEvidence              `json:"evidence"`
}

type ContinuationGame struct {
	ID   string `json:"id"`
	Kind string `json:"kind,omitempty"`
}

type ContinuationRound struct {
	ID string `json:"id"`
}

type ContinuationInterface struct {
	ID string `json:"id"`
}

type ContinuationCandidate struct {
	Role         domain.Role       `json:"role"`
	System       domain.SystemSpec `json:"system"`
	ArtifactPath string            `json:"artifact_path"`
}

type ContinuationDefaults struct {
	IncumbentFrom string `json:"incumbent_from"`
	MatchesFrom   string `json:"matches_from"`
	ObjectiveFrom string `json:"objective_from"`
	EvaluatorFrom string `json:"evaluator_from"`
}

type ContinuationEvaluator struct {
	Model        ContinuationModel  `json:"model"`
	Bounds       ContinuationBounds `json:"bounds"`
	Retry        ContinuationRetry  `json:"retry"`
	AllowedTools []string           `json:"allowed_tools,omitempty"`
	DeniedTools  []string           `json:"denied_tools,omitempty"`
	SystemPrompt string             `json:"system_prompt,omitempty"`
	PolicySHA256 string             `json:"policy_sha256,omitempty"`
}

type ContinuationModel struct {
	Provider        string `json:"provider"`
	Name            string `json:"name"`
	MaxOutputTokens int    `json:"max_output_tokens,omitempty"`
}

type ContinuationBounds struct {
	MaxModelTurns  int `json:"max_model_turns,omitempty"`
	MaxToolCalls   int `json:"max_tool_calls,omitempty"`
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
}

type ContinuationRetry struct {
	MaxAttempts                int  `json:"max_attempts,omitempty"`
	RetryOnModelError          bool `json:"retry_on_model_error,omitempty"`
	RetryOnToolFailure         bool `json:"retry_on_tool_failure,omitempty"`
	RetryOnFinalizationFailure bool `json:"retry_on_finalization_failure,omitempty"`
	RetryOnInvalidPrediction   bool `json:"retry_on_invalid_prediction,omitempty"`
}

type ContinuationScoring struct {
	ObjectivePath string   `json:"objective_path"`
	ReportFormats []string `json:"report_formats,omitempty"`
}

type ContinuationEvidence struct {
	RoundEvidencePath string `json:"round_evidence_path"`
	ObjectivePath     string `json:"objective_path,omitempty"`
	ReportPath        string `json:"report_path"`
	DecisionPath      string `json:"decision_path"`
}

func (c Continuation) Validate() error {
	if strings.TrimSpace(c.SchemaVersion) == "" {
		return fmt.Errorf("%w: schema version is required", ErrInvalidContinuation)
	}
	if c.SchemaVersion != ContinuationSchemaVersion {
		return fmt.Errorf("%w: unsupported schema version %q", ErrInvalidContinuation, c.SchemaVersion)
	}
	if strings.TrimSpace(c.Game.ID) == "" {
		return fmt.Errorf("%w: %w", ErrInvalidContinuation, ErrMissingContinuationGame)
	}
	if strings.TrimSpace(c.Round.ID) == "" {
		return fmt.Errorf("%w: %w", ErrInvalidContinuation, ErrMissingContinuationRound)
	}
	if strings.TrimSpace(c.CandidateInterface.ID) == "" {
		return fmt.Errorf("%w: %w", ErrInvalidContinuation, ErrMissingContinuationIface)
	}
	if c.SurvivingCandidate.Role == "" {
		return fmt.Errorf("%w: %w", ErrInvalidContinuation, ErrMissingContinuationRole)
	}
	if err := c.SurvivingCandidate.System.Validate(); err != nil {
		return fmt.Errorf("%w: surviving candidate system: %w", ErrInvalidContinuation, err)
	}
	if strings.TrimSpace(c.SurvivingCandidate.ArtifactPath) == "" {
		return fmt.Errorf("%w: %w", ErrInvalidContinuation, ErrMissingContinuationPath)
	}
	if strings.TrimSpace(c.Scoring.ObjectivePath) == "" {
		return fmt.Errorf("%w: objective path is required", ErrInvalidContinuation)
	}
	if strings.TrimSpace(c.Evidence.RoundEvidencePath) == "" || strings.TrimSpace(c.Evidence.ReportPath) == "" || strings.TrimSpace(c.Evidence.DecisionPath) == "" {
		return fmt.Errorf("%w: evidence paths are required", ErrInvalidContinuation)
	}
	if err := c.Matches.Validate(); err != nil {
		return fmt.Errorf("%w: matches: %w", ErrInvalidContinuation, err)
	}
	return nil
}

// ResolveArtifactPath expands a bundle-relative artifact path against the
// completed round bundle path.
func (c Continuation) ResolveArtifactPath(bundlePath domain.HostPath) string {
	if strings.TrimSpace(c.SurvivingCandidate.ArtifactPath) == "" {
		return ""
	}
	return filepath.Join(string(bundlePath), filepath.FromSlash(c.SurvivingCandidate.ArtifactPath))
}
