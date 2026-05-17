package bundlefs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
)

// ValidateCompletedBundle checks a config-owned completed bundle tree (#83).
func ValidateCompletedBundle(bundlePath string) (report.CanonicalReport, error) {
	bundlePath = strings.TrimSpace(bundlePath)
	if bundlePath == "" {
		return report.CanonicalReport{}, fmt.Errorf("bundle path is required")
	}
	complete := filepath.Join(bundlePath, completeMarkerName)
	if _, err := os.Stat(complete); err != nil {
		return report.CanonicalReport{}, fmt.Errorf("bundle missing %s: %w", completeMarkerName, err)
	}

	reportPath := filepath.Join(bundlePath, "round-report.json")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return report.CanonicalReport{}, fmt.Errorf("read round report: %w", err)
	}

	var roundReport report.RoundReport
	if err := json.Unmarshal(data, &roundReport); err != nil {
		return report.CanonicalReport{}, fmt.Errorf("decode round report: %w", err)
	}

	counts := failureCategoryCounts(roundReport.Failures)
	passed := totalFailures(counts) == 0 &&
		len(roundReport.Failures.Incumbent) == 0 &&
		len(roundReport.Failures.Challenger) == 0

	decision := ""
	if roundReport.Decision.Decision != "" {
		decision = string(roundReport.Decision.Decision)
	}

	canonical := report.DefaultCanonicalReport(
		report.ModeValidateBundle,
		report.FreshnessArchive,
		passed,
		"",
		bundlePath,
		string(roundReport.ID),
		decision,
		"",
		0,
	)
	canonical.FailureCounts = counts
	if !passed {
		canonical.Passed = false
	}
	return canonical, nil
}

func failureCategoryCounts(failures domain.Pair[[]execution.RunFailure]) map[string]int {
	counts := make(map[string]int)
	for _, f := range failures.Incumbent {
		counts[string(f.Category)]++
	}
	for _, f := range failures.Challenger {
		counts[string(f.Category)]++
	}
	return counts
}

func totalFailures(counts map[string]int) int {
	n := 0
	for _, v := range counts {
		n += v
	}
	return n
}
