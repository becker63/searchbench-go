package text

import (
	"fmt"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/report"
)

// RenderCanonical formats the primary human-facing bundle summary (#84).
func RenderCanonical(c report.CanonicalReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "SearchBench canonical report\n")
	fmt.Fprintf(&b, "mode=%s freshness=%s passed=%v\n", c.Mode, c.Freshness, c.Passed)
	if c.RoundID != "" {
		fmt.Fprintf(&b, "round_id=%s\n", c.RoundID)
	}
	if c.BundlePath != "" {
		fmt.Fprintf(&b, "bundle_path=%s\n", c.BundlePath)
	}
	if c.Decision != "" {
		fmt.Fprintf(&b, "decision=%s final=%s value=%0.6f\n", c.Decision, c.Final, c.FinalVal)
	}
	if len(c.FailureCounts) > 0 {
		fmt.Fprintf(&b, "failure_counts:\n")
		for k, v := range c.FailureCounts {
			fmt.Fprintf(&b, "  %s: %d\n", k, v)
		}
	}
	if c.Attempts != nil {
		a := c.Attempts
		fmt.Fprintf(&b, "attempts: count=%d pass_rate=%0.3f infra_failure_rate=%0.3f\n",
			a.Count, a.PassRate, a.InfraFailureRate)
		if a.SamePredictionRate > 0 {
			fmt.Fprintf(&b, "stability: same_prediction_rate=%0.3f token_spread=%d tool_call_spread=%d\n",
				a.SamePredictionRate, a.TokenSpread, a.ToolCallSpread)
		}
	}
	if c.InputFingerprint != "" {
		fmt.Fprintf(&b, "input_fingerprint=%s\n", c.InputFingerprint)
	}
	if c.ModelSeed != "" {
		fmt.Fprintf(&b, "model_seed=%s\n", c.ModelSeed)
	}
	return b.String()
}
