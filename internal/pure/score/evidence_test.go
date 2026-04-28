package score

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewMetricEvidencePreservesDirectionAndFlags(t *testing.T) {
	t.Parallel()

	got, err := NewMetricEvidence(MetricCost, 0.6, 0.1)
	if err != nil {
		t.Fatalf("NewMetricEvidence() error = %v", err)
	}
	if got.Direction != LowerIsBetter {
		t.Fatalf("Direction = %q, want %q", got.Direction, LowerIsBetter)
	}
	if !got.Improved || got.Regressed {
		t.Fatalf("flags = improved:%v regressed:%v, want true/false", got.Improved, got.Regressed)
	}
}

func TestAggregateUsageIsHonestAboutAvailability(t *testing.T) {
	t.Parallel()

	got := AggregateUsage(nil)
	if got.Available {
		t.Fatalf("Available = true, want false")
	}
	if got.MeasuredRuns != 0 {
		t.Fatalf("MeasuredRuns = %d, want 0", got.MeasuredRuns)
	}
}

func TestScoreEvidenceValidateRejectsMissingIdentity(t *testing.T) {
	t.Parallel()

	doc := ScoreEvidenceDocument{}
	if err := doc.Validate(); err == nil || !strings.Contains(err.Error(), ErrMissingEvidenceSchemaVersion.Error()) {
		t.Fatalf("Validate() error = %v, want missing schema version error", err)
	}

	doc.SchemaVersion = EvidenceSchemaVersion
	if err := doc.Validate(); err == nil || !strings.Contains(err.Error(), ErrMissingEvidenceReportID.Error()) {
		t.Fatalf("Validate() error = %v, want missing report id error", err)
	}
}

func TestExtractLocalizationDistanceProjectsNamedMetrics(t *testing.T) {
	t.Parallel()

	metrics := []MetricEvidence{
		{Metric: MetricGoldHop},
		{Metric: MetricIssueHop},
		{Metric: MetricComposite},
	}
	got := ExtractLocalizationDistance(metrics)
	if got.GoldHop == nil || got.IssueHop == nil {
		t.Fatalf("localization distance = %#v, want gold_hop and issue_hop", got)
	}
}

func TestScoreEvidencePackageAvoidsForbiddenImports(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	dir := filepath.Dir(currentFile)
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, dir, func(info os.FileInfo) bool {
		name := info.Name()
		return name == "evidence.go"
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parser.ParseDir() error = %v", err)
	}

	forbiddenSubstrings := []string{
		"pkl",
		"cloudwego/eino",
		"mcp",
		"langsmith",
		"langfuse",
		"materialization",
		"tree-sitter",
		"treesitter",
		"internal/adapters/artifact",
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				for _, forbidden := range forbiddenSubstrings {
					if strings.Contains(strings.ToLower(path), forbidden) {
						t.Fatalf("forbidden import %q contains %q", path, forbidden)
					}
				}
			}
		}
	}
}
