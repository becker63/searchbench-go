package scoring

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestEvaluateLocalObjectiveWithParent(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	request := sampleRequest(t)
	result, err := Evaluate(context.Background(), request)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if result.ObjectiveID != "localization-v1" {
		t.Fatalf("result.ObjectiveID = %q", result.ObjectiveID)
	}
	if result.Final != "final" {
		t.Fatalf("result.Final = %q, want final", result.Final)
	}
	if got, want := len(result.EvidenceRefs), 2; got != want {
		t.Fatalf("len(result.EvidenceRefs) = %d, want %d", got, want)
	}
	if _, ok := result.FinalValue(); !ok {
		t.Fatal("FinalValue() missing")
	}
	if !hasValue(result.Values, "currentLocalizationQuality") {
		t.Fatalf("values = %#v, want currentLocalizationQuality", result.Values)
	}
	if !hasValue(result.Values, "parentLocalizationQuality") {
		t.Fatalf("values = %#v, want parentLocalizationQuality", result.Values)
	}
	if !hasValue(result.Values, "regressionPenalty") {
		t.Fatalf("values = %#v, want regressionPenalty", result.Values)
	}
	finalValue, ok := result.FinalValue()
	if !ok {
		t.Fatal("FinalValue() missing")
	}
	if finalValue.Name != result.Final {
		t.Fatalf("FinalValue().Name = %q, want %q from Final selector", finalValue.Name, result.Final)
	}
	if !hasValue(result.Values, "final") {
		t.Fatalf("values = %#v, want final", result.Values)
	}
}

func TestEvaluateLocalObjectiveWithoutParent(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	request := sampleRequest(t)
	request.ParentRef = nil

	result, err := Evaluate(context.Background(), request)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if got, want := len(result.EvidenceRefs), 1; got != want {
		t.Fatalf("len(result.EvidenceRefs) = %d, want %d", got, want)
	}
	if result.Final != "final" {
		t.Fatalf("result.Final = %q, want final", result.Final)
	}
	if !hasValue(result.Values, "parentLocalizationQuality") {
		t.Fatalf("values = %#v, want parentLocalizationQuality present", result.Values)
	}
	if !hasValue(result.Values, "regressionPenalty") {
		t.Fatalf("values = %#v, want regressionPenalty present", result.Values)
	}
	if !hasValue(result.Values, "final") {
		t.Fatalf("values = %#v, want final present", result.Values)
	}
}

func TestEvaluateRejectsInvalidObjectiveOutput(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	request := sampleRequest(t)
	request.ScoringPath = writeTempModule(t, `amends `+quotePkl(fileURI(filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "scoring", "localization-objective.pkl")))+`

final = ""
`)

	_, err := Evaluate(context.Background(), request)
	if err == nil || !strings.Contains(err.Error(), ErrValidate.Error()) {
		t.Fatalf("Evaluate() error = %v, want validation error", err)
	}
}

func TestEvaluateRejectsMissingScoringFile(t *testing.T) {
	t.Parallel()

	request := sampleRequest(t)
	request.ScoringPath = filepath.Join(t.TempDir(), "missing.pkl")

	_, err := Evaluate(context.Background(), request)
	if err == nil || !strings.Contains(err.Error(), ErrInvalidRequest.Error()) {
		t.Fatalf("Evaluate() error = %v, want invalid request error", err)
	}
}

func TestVisibleObjectiveFileUsesSharedValueHelpers(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "scoring", "localization-objective.pkl")
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	content := string(contentBytes)

	if strings.Contains(content, "new ObjectiveValue") {
		t.Fatalf("visible objective file still contains inline ObjectiveValue construction:\n%s", content)
	}
	for _, want := range []string{
		`helpers.intermediate("currentLocalizationQuality", currentLocalizationQuality)`,
		`helpers.intermediate("parentLocalizationQuality", parentLocalizationQuality)`,
		`helpers.penalty("regressionPenalty", regressionPenalty)`,
		`helpers.finalValue("final", finalScore)`,
		`final = "final"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("visible objective file missing %q:\n%s", want, content)
		}
	}
}

func sampleRequest(t *testing.T) Request {
	t.Helper()

	root := repoRoot(t)
	currentPath := writeScoreModule(t, "current.pkl", sampleEvidence("current-report", 4, 1, 6))
	parentPath := writeScoreModule(t, "parent.pkl", sampleEvidence("parent-report", 6, 0, 9))
	return Request{
		ScoringPath: filepath.Join(root, "configs", "experiments", "local-ic-vs-jcodemunch", "scoring", "localization-objective.pkl"),
		CurrentRef: score.ObjectiveEvidenceRef{
			Name:      "current",
			ScorePath: currentPath,
		},
		ParentRef: &score.ObjectiveEvidenceRef{
			Name:      "parent",
			ScorePath: parentPath,
		},
	}
}

func sampleEvidence(reportID string, goldHop float64, severeCount int, totalTokens domain.TokenCount) score.ScoreEvidenceDocument {
	goldMetric, err := score.NewMetricEvidence(score.MetricGoldHop, goldHop+1, goldHop)
	if err != nil {
		panic(err)
	}
	return score.ScoreEvidenceDocument{
		SchemaVersion: score.EvidenceSchemaVersion,
		ReportID:      domain.ReportID(reportID),
		LocalizationDistance: score.LocalizationDistanceEvidence{
			GoldHop: &goldMetric,
		},
		Usage: score.UsageEvidence{
			Available:   true,
			TotalTokens: totalTokens,
		},
		Regressions: score.RegressionEvidenceSummary{
			SevereCount: severeCount,
		},
		InvalidPredictions: score.InvalidPredictionEvidence{
			Known: true,
			Count: 0,
		},
	}
}

func hasValue(values []score.ObjectiveValue, name string) bool {
	for _, value := range values {
		if value.Name == name {
			return true
		}
	}
	return false
}

func requirePkl(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pkl"); err != nil {
		t.Skip("pkl CLI not available on PATH")
	}
}

func writeTempModule(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "objective.pkl")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
	return path
}

func writeScoreModule(t *testing.T, name string, evidence score.ScoreEvidenceDocument) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	content := fmt.Sprintf(`schemaVersion = %q
reportId = %q

localizationDistance {
  goldHop {
    candidate = %v
  }
}

usage {
  totalTokens = %d
}

regressions {
  severeCount = %d
}

invalidPredictions {
  known = %t
  count = %d
}
`,
		evidence.SchemaVersion,
		evidence.ReportID.String(),
		evidence.LocalizationDistance.GoldHop.Candidate,
		evidence.Usage.TotalTokens,
		evidence.Regressions.SevereCount,
		evidence.InvalidPredictions.Known,
		evidence.InvalidPredictions.Count,
	)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
	return path
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatalf("filepath.Abs(repo root) error = %v", err)
	}
	return root
}

func quotePkl(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}
