package round

import (
	"math"
	"sort"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
)

const (
	minCompletionFraction        = 2.0 / 3.0
	minPassRateForPromotion      = 0.67
	maxInfraRateForPromotion     = 0.34
	stabilitySamePredictionMin   = 0.8
	stabilityMaxTokenSpreadRatio = 0.25
)

// AttemptOutcome records one live evaluation attempt (#88, #91).
type AttemptOutcome struct {
	AttemptID int
	Completed bool
	Passed    bool
	FinalVal  float64
	Failures  map[string]int

	TotalTokens           int
	ToolCalls             int
	Decision              string
	PredictionFingerprint string
}

// ConsolidationInput configures multi-attempt aggregation.
type ConsolidationInput struct {
	Requested int
	Stability bool
}

// ConsolidationResult is the consolidated live evaluation summary (#91).
type ConsolidationResult struct {
	Summary             *report.AttemptSummary
	Passed              bool
	Decision            string
	PromotionGatePassed bool
}

// ConsolidateAttempts builds a summary for evaluate_n and stability_probe (#88, #89, #91).
func ConsolidateAttempts(outcomes []AttemptOutcome, input ConsolidationInput) ConsolidationResult {
	requested := input.Requested
	if requested <= 0 {
		requested = len(outcomes)
	}
	if len(outcomes) == 0 {
		return ConsolidationResult{
			Summary: &report.AttemptSummary{
				RequestedCount: requested,
				Verdict:        verdictFail,
			},
			Passed:   false,
			Decision: string(report.DecisionRejectChallenger),
		}
	}

	completed := 0
	passed := 0
	infraFails := 0
	modelFails := 0
	vals := make([]float64, 0, len(outcomes))
	tokens := make([]int, 0, len(outcomes))
	toolCalls := make([]int, 0, len(outcomes))
	fingerprints := make([]string, 0, len(outcomes))

	for _, o := range outcomes {
		if o.Completed {
			completed++
		}
		if o.Passed {
			passed++
		}
		vals = append(vals, o.FinalVal)
		infraFails += o.Failures[string(execution.FailureCategoryInfrastructure)]
		modelFails += o.Failures[string(execution.FailureCategoryModel)]
		if o.Completed && o.TotalTokens > 0 {
			tokens = append(tokens, o.TotalTokens)
		}
		if o.Completed {
			toolCalls = append(toolCalls, o.ToolCalls)
		}
		if o.Completed && strings.TrimSpace(o.PredictionFingerprint) != "" {
			fingerprints = append(fingerprints, o.PredictionFingerprint)
		}
	}

	sort.Float64s(vals)
	medianVal := medianFloat(vals)
	medianTokens := medianInt(tokens)
	tokenMin, tokenMax, tokenSpread := intSpread(tokens)
	toolMin, toolMax, toolSpread := intSpread(toolCalls)
	samePrediction := samePredictionRate(fingerprints)

	count := len(outcomes)
	summary := &report.AttemptSummary{
		RequestedCount:     requested,
		CompletedCount:     completed,
		Count:              count,
		PassRate:           float64(passed) / float64(count),
		MedianFinalValue:   medianVal,
		InfraFailureRate:   float64(infraFails) / float64(count),
		ModelFailureRate:   float64(modelFails) / float64(count),
		PredictionPassRate: float64(passed) / float64(count),
		MedianTotalTokens:  medianTokens,
		TokenMin:           tokenMin,
		TokenMax:           tokenMax,
		TokenSpread:        tokenSpread,
		ToolCallMin:        toolMin,
		ToolCallMax:        toolMax,
		ToolCallSpread:     toolSpread,
		SamePredictionRate: samePrediction,
	}

	minCompleted := minRequiredCompletions(requested)
	if input.Stability {
		summary.Verdict = string(computeStabilityVerdict(summary, completed, minCompleted))
		return ConsolidationResult{
			Summary:  summary,
			Passed:   summary.Verdict == string(stabilityStable),
			Decision: "NO_DECISION",
		}
	}

	gatePassed := promotionGatePassed(summary, completed, minCompleted)
	summary.PromotionGatePassed = gatePassed
	summary.Verdict = evaluateVerdict(summary, gatePassed, completed, minCompleted)

	decision := string(report.DecisionRejectChallenger)
	if summary.Verdict == verdictPromote {
		decision = string(report.DecisionPromoteChallenger)
	} else if summary.Verdict == verdictReview {
		decision = string(report.DecisionReview)
	}

	return ConsolidationResult{
		Summary:             summary,
		Passed:              summary.Verdict == verdictPass || summary.Verdict == verdictPromote,
		Decision:            decision,
		PromotionGatePassed: gatePassed,
	}
}

const (
	verdictPass     = "PASS"
	verdictFail     = "FAIL"
	verdictUnstable = "UNSTABLE"
	verdictPromote  = "PROMOTE_CHALLENGER"
	verdictReview   = "REVIEW"
)

type liveStabilityVerdict string

const (
	stabilityStable   liveStabilityVerdict = "STABLE"
	stabilityUnstable liveStabilityVerdict = "UNSTABLE"
	stabilityFail     liveStabilityVerdict = "FAIL"
)

func promotionGatePassed(summary *report.AttemptSummary, completed, minCompleted int) bool {
	if summary == nil || completed < minCompleted {
		return false
	}
	if summary.InfraFailureRate > maxInfraRateForPromotion {
		return false
	}
	if summary.PassRate < minPassRateForPromotion {
		return false
	}
	if summary.PredictionPassRate < minPassRateForPromotion {
		return false
	}
	return summary.MedianFinalValue > 0
}

func evaluateVerdict(summary *report.AttemptSummary, gatePassed bool, completed, minCompleted int) string {
	if completed < minCompleted {
		return verdictFail
	}
	if !gatePassed {
		if summary.InfraFailureRate > maxInfraRateForPromotion || summary.PassRate < minPassRateForPromotion {
			return verdictFail
		}
		return verdictUnstable
	}
	if summary.PassRate >= minPassRateForPromotion && summary.MedianFinalValue > 0 {
		return verdictPromote
	}
	if summary.PassRate >= minPassRateForPromotion {
		return verdictPass
	}
	return verdictReview
}

func computeStabilityVerdict(summary *report.AttemptSummary, completed, minCompleted int) liveStabilityVerdict {
	if completed < minCompleted {
		return stabilityFail
	}
	if summary.InfraFailureRate > maxInfraRateForPromotion {
		return stabilityFail
	}
	if summary.SamePredictionRate >= stabilitySamePredictionMin {
		if summary.MedianTotalTokens == 0 || float64(summary.TokenSpread)/float64(summary.MedianTotalTokens) <= stabilityMaxTokenSpreadRatio {
			return stabilityStable
		}
	}
	return stabilityUnstable
}

func minRequiredCompletions(requested int) int {
	if requested <= 0 {
		return 1
	}
	return int(math.Ceil(float64(requested) * minCompletionFraction))
}

func medianFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sort.Float64s(vals)
	return vals[len(vals)/2]
}

func medianInt(vals []int) int {
	if len(vals) == 0 {
		return 0
	}
	sort.Ints(vals)
	return vals[len(vals)/2]
}

func intSpread(vals []int) (min, max, spread int) {
	if len(vals) == 0 {
		return 0, 0, 0
	}
	min, max = vals[0], vals[0]
	for _, v := range vals[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max, max - min
}

func samePredictionRate(fingerprints []string) float64 {
	if len(fingerprints) == 0 {
		return 0
	}
	counts := make(map[string]int, len(fingerprints))
	best := 0
	for _, fp := range fingerprints {
		counts[fp]++
		if counts[fp] > best {
			best = counts[fp]
		}
	}
	return float64(best) / float64(len(fingerprints))
}
