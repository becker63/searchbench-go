package execution

import (
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

type FailureStage string

const (
	// FailurePrepare identifies prepare/session setup failures.
	FailurePrepare FailureStage = "prepare"
	// FailureExecute identifies execution-time failures.
	FailureExecute FailureStage = "execute"
	// FailureScore identifies scoring-time failures.
	FailureScore FailureStage = "score"
	// FailureReport identifies report construction failures.
	FailureReport FailureStage = "report"
)

// RunFailure records the report-facing failed path for one match/system run.
//
// ExecutedRun means execution succeeded. score.ScoredRun means scoring
// succeeded. RunFailure is the separate artifact used when one of those steps
// could not complete.
// FailureCategory classifies report-facing failures for consolidated live reports.
type FailureCategory string

const (
	FailureCategoryInfrastructure FailureCategory = "infrastructure"
	FailureCategoryModel          FailureCategory = "model"
	FailureCategoryTool           FailureCategory = "tool"
	FailureCategoryScoring        FailureCategory = "scoring"
	FailureCategoryPrediction     FailureCategory = "prediction"
	FailureCategoryArtifact       FailureCategory = "artifact"
	FailureCategoryPromotionGate  FailureCategory = "promotion_gate"
	FailureCategoryUnknown        FailureCategory = "unknown"
)

type RunFailure struct {
	RunID    domain.RunID    `json:"run_id"`
	MatchID  domain.MatchID  `json:"match_id"`
	System   domain.SystemID `json:"system"`
	Stage    FailureStage    `json:"stage"`
	Category FailureCategory `json:"category"`
	Message  string          `json:"message"`
}

// NewFailure constructs a report-facing failure record from a run spec.
func NewFailure(spec Spec, stage FailureStage, message string) RunFailure {
	return RunFailure{
		RunID:    spec.ID,
		MatchID:  spec.Match.ID,
		System:   spec.System.ID,
		Stage:    stage,
		Category: classifyFailure(stage, message),
		Message:  message,
	}
}

func classifyFailure(stage FailureStage, message string) FailureCategory {
	switch stage {
	case FailureScore:
		return FailureCategoryScoring
	case FailureReport:
		return FailureCategoryArtifact
	case FailurePrepare:
		if containsFold(message, "context", "overflow", "token") {
			return FailureCategoryModel
		}
		return FailureCategoryInfrastructure
	case FailureExecute:
		if containsFold(message, "mcp", "tool", "backend") {
			return FailureCategoryTool
		}
		if containsFold(message, "cerebras", "model", "provider", "rate limit", "api") {
			return FailureCategoryModel
		}
		if containsFold(message, "prediction", "invalid", "empty predicted") {
			return FailureCategoryPrediction
		}
		return FailureCategoryInfrastructure
	default:
		return FailureCategoryUnknown
	}
}

func containsFold(haystack string, needles ...string) bool {
	lower := strings.ToLower(haystack)
	for _, n := range needles {
		if strings.Contains(lower, strings.ToLower(n)) {
			return true
		}
	}
	return false
}
