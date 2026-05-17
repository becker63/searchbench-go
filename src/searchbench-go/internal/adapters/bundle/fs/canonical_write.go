package bundlefs

import (
	"fmt"

	canonicaltext "github.com/becker63/searchbench-go/internal/adapters/report/text"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// CanonicalArtifacts builds report.json and report.txt bundle files (#84).
func CanonicalArtifacts(
	mode report.Mode,
	freshness report.Freshness,
	passed bool,
	roundID string,
	bundlePath string,
	roundReport report.RoundReport,
	objective *score.ObjectiveResult,
	extra report.CanonicalReport,
) ([]BundleArtifact, error) {
	canonical := report.DefaultCanonicalReport(
		mode,
		freshness,
		passed,
		roundID,
		bundlePath,
		string(roundReport.ID),
		string(roundReport.Decision.Decision),
		finalLabel(objective),
		finalValue(objective),
	)
	canonical.FailureCounts = failureCategoryCounts(roundReport.Failures)
	if extra.Attempts != nil {
		canonical.Attempts = extra.Attempts
	}
	if extra.InputFingerprint != "" {
		canonical.InputFingerprint = extra.InputFingerprint
	}
	if extra.ModelSeed != "" {
		canonical.ModelSeed = extra.ModelSeed
	}
	if extra.GenerationConfig != nil {
		canonical.GenerationConfig = extra.GenerationConfig
	}
	if len(extra.RequestHashes) > 0 {
		canonical.RequestHashes = extra.RequestHashes
	}
	if len(extra.ResponseHashes) > 0 {
		canonical.ResponseHashes = extra.ResponseHashes
	}
	if extra.Verdict != "" {
		canonical.Verdict = extra.Verdict
	}
	if extra.PromotionVerdict != "" {
		canonical.PromotionVerdict = extra.PromotionVerdict
	}

	jsonBytes, err := marshalDeterministic(canonical)
	if err != nil {
		return nil, fmt.Errorf("canonical report json: %w", err)
	}
	txt := canonicaltext.RenderCanonical(canonical)
	return []BundleArtifact{
		{Kind: "canonical_report", Path: "report.json", MediaType: "application/json", Content: jsonBytes},
		{Kind: "canonical_report_txt", Path: "report.txt", MediaType: "text/plain", Content: []byte(txt)},
	}, nil
}

func finalLabel(objective *score.ObjectiveResult) string {
	if objective == nil {
		return ""
	}
	return objective.Final
}

func finalValue(objective *score.ObjectiveResult) float64 {
	if objective == nil {
		return 0
	}
	v, ok := objective.FinalValue()
	if !ok {
		return 0
	}
	return v.Value
}
