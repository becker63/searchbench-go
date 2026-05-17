package report

import "time"

// Freshness describes whether a report reflects a fresh live run or replayed artifacts.
type Freshness string

const (
	FreshnessFresh        Freshness = "fresh"
	FreshnessFreshLiveRun Freshness = "fresh_live_run"
	FreshnessArchive      Freshness = "archive"
)

// Mode names the evaluation mode that produced this report.
type Mode string

const (
	ModeValidateBundle Mode = "validate_bundle"
	ModeLiveSmoke      Mode = "live_smoke"
	ModeEvaluateN      Mode = "evaluate_n"
	ModeStabilityProbe Mode = "stability_probe"
	ModeRoundRun       Mode = "round_run"
)

// Verdict summarizes stability_probe outcomes (#91).
type Verdict string

const (
	VerdictStable   Verdict = "STABLE"
	VerdictUnstable Verdict = "UNSTABLE"
	VerdictFail     Verdict = "FAIL"
)

// PromotionVerdict records whether evaluate_n supports promotion (#91).
type PromotionVerdict string

const (
	PromotionVerdictPromote      PromotionVerdict = "PROMOTE"
	PromotionVerdictHold         PromotionVerdict = "HOLD"
	PromotionVerdictInsufficient PromotionVerdict = "INSUFFICIENT"
)

// CanonicalReport is the primary human/agent-facing bundle summary (#84).
type CanonicalReport struct {
	SchemaVersion string    `json:"schema_version"`
	Mode          Mode      `json:"mode"`
	Freshness     Freshness `json:"freshness"`
	Passed        bool      `json:"passed"`
	CreatedAt     time.Time `json:"created_at"`

	RoundID    string `json:"round_id,omitempty"`
	BundlePath string `json:"bundle_path,omitempty"`
	ReportID   string `json:"report_id,omitempty"`

	Decision string  `json:"decision,omitempty"`
	Final    string  `json:"final,omitempty"`
	FinalVal float64 `json:"final_value,omitempty"`

	FailureCounts map[string]int `json:"failure_counts,omitempty"`

	Attempts *AttemptSummary `json:"attempts,omitempty"`

	InputFingerprint string         `json:"input_fingerprint,omitempty"`
	ModelSeed        string         `json:"model_seed,omitempty"`
	GenerationConfig map[string]any `json:"generation_config,omitempty"`
	RequestHashes    []string       `json:"request_hashes,omitempty"`
	ResponseHashes   []string       `json:"response_hashes,omitempty"`

	Verdict          Verdict          `json:"verdict,omitempty"`
	PromotionVerdict PromotionVerdict `json:"promotion_verdict,omitempty"`
}

// AttemptSummary consolidates multi-attempt live evaluation (#88, #89, #91).
type AttemptSummary struct {
	RequestedCount int `json:"requested_count,omitempty"`
	CompletedCount int `json:"completed_count,omitempty"`
	Count          int `json:"count"`

	PassRate           float64 `json:"pass_rate"`
	MedianFinalValue   float64 `json:"median_final_value,omitempty"`
	InfraFailureRate   float64 `json:"infra_failure_rate"`
	ModelFailureRate   float64 `json:"model_failure_rate,omitempty"`
	PredictionPassRate float64 `json:"prediction_pass_rate,omitempty"`

	MedianTotalTokens int `json:"median_total_tokens,omitempty"`
	TokenMin          int `json:"token_min,omitempty"`
	TokenMax          int `json:"token_max,omitempty"`

	SamePredictionRate float64 `json:"same_prediction_rate,omitempty"`
	TokenSpread        int     `json:"token_spread,omitempty"`
	ToolCallSpread     int     `json:"tool_call_spread,omitempty"`
	ToolCallMin        int     `json:"tool_call_min,omitempty"`
	ToolCallMax        int     `json:"tool_call_max,omitempty"`

	PromotionGatePassed bool   `json:"promotion_gate_passed,omitempty"`
	Verdict             string `json:"verdict,omitempty"`
}

// DefaultCanonicalReport builds a minimal canonical report for one completed round.
func DefaultCanonicalReport(mode Mode, freshness Freshness, passed bool, roundID, bundlePath, reportID, decision, final string, finalVal float64) CanonicalReport {
	return CanonicalReport{
		SchemaVersion: "searchbench.canonical_report.v1",
		Mode:          mode,
		Freshness:     freshness,
		Passed:        passed,
		CreatedAt:     time.Now().UTC(),
		RoundID:       roundID,
		BundlePath:    bundlePath,
		ReportID:      reportID,
		Decision:      decision,
		Final:         final,
		FinalVal:      finalVal,
	}
}
