package score

import (
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestObjectiveResultValidateAcceptsFuturePklShapedResult(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	if err := result.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	finalValue, ok := result.FinalValue()
	if !ok {
		t.Fatal("FinalValue() = missing, want present")
	}
	if finalValue.Name != "final" || finalValue.Value != 0.77 {
		t.Fatalf("FinalValue() = %#v, want final=0.77", finalValue)
	}
}

func TestObjectiveResultValidateAcceptsCurrentAndParentEvidenceRefs(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	if got, want := len(result.EvidenceRefs), 2; got != want {
		t.Fatalf("len(EvidenceRefs) = %d, want %d", got, want)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestObjectiveResultPreservesIntermediateAndPenaltyValues(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()

	current := findObjectiveValue(t, result.Values, "currentLocalizationQuality")
	if current.Kind != ObjectiveValueIntermediate {
		t.Fatalf("current.Kind = %q, want %q", current.Kind, ObjectiveValueIntermediate)
	}

	penalty := findObjectiveValue(t, result.Values, "regressionPenalty")
	if penalty.Kind != ObjectiveValuePenalty {
		t.Fatalf("penalty.Kind = %q, want %q", penalty.Kind, ObjectiveValuePenalty)
	}
	if penalty.Value != 1.0 {
		t.Fatalf("penalty.Value = %f, want 1.0", penalty.Value)
	}
}

func TestObjectiveResultRejectsDuplicateValueNames(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	result.Values = append(result.Values, ObjectiveValue{
		Name:  "base",
		Value: 0.5,
		Kind:  ObjectiveValueIntermediate,
	})

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrDuplicateObjectiveValue.Error()) {
		t.Fatalf("Validate() error = %v, want duplicate objective value error", err)
	}
}

func TestObjectiveResultRejectsMissingSchemaVersion(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	result.SchemaVersion = ""

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrMissingSchemaVersion.Error()) {
		t.Fatalf("Validate() error = %v, want missing schema version error", err)
	}
}

func TestObjectiveResultRejectsMissingObjectiveID(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	result.ObjectiveID = ""

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrMissingObjectiveID.Error()) {
		t.Fatalf("Validate() error = %v, want missing objective id error", err)
	}
}

func TestObjectiveResultRejectsEmptyObjectiveValueName(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	result.Values[0].Name = ""

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), "objective value name is required") {
		t.Fatalf("Validate() error = %v, want empty objective value name error", err)
	}
}

func TestObjectiveResultRejectsMissingFinalValue(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	result.Final = ""

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrMissingFinalValue.Error()) {
		t.Fatalf("Validate() error = %v, want missing final value error", err)
	}
}

func TestObjectiveResultRejectsFinalValueNotPresentInValues(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	result.Final = "not_present"

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrFinalValueNotFound.Error()) {
		t.Fatalf("Validate() error = %v, want final value not found error", err)
	}
}

func TestObjectiveResultRejectsNaN(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	setObjectiveValue(t, &result, "base", math.NaN())

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrInvalidNumericValue.Error()) {
		t.Fatalf("Validate() error = %v, want invalid numeric value error", err)
	}
}

func TestObjectiveResultRejectsPositiveInfinity(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	setObjectiveValue(t, &result, "base", math.Inf(1))

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrInvalidNumericValue.Error()) {
		t.Fatalf("Validate() error = %v, want invalid numeric value error", err)
	}
}

func TestObjectiveResultRejectsNegativeInfinity(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	setObjectiveValue(t, &result, "base", math.Inf(-1))

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrInvalidNumericValue.Error()) {
		t.Fatalf("Validate() error = %v, want invalid numeric value error", err)
	}
}

func TestObjectiveResultEnforcesMinBound(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	min := 0.80
	result.Bounds = &ObjectiveBounds{Min: &min}

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrFinalValueOutOfBounds.Error()) {
		t.Fatalf("Validate() error = %v, want out-of-bounds error", err)
	}
}

func TestObjectiveResultEnforcesMaxBound(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	max := 0.70
	result.Bounds = &ObjectiveBounds{Max: &max}

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrFinalValueOutOfBounds.Error()) {
		t.Fatalf("Validate() error = %v, want out-of-bounds error", err)
	}
}

func TestObjectiveResultRejectsDuplicateEvidenceRefNames(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	result.EvidenceRefs = append(result.EvidenceRefs, ObjectiveEvidenceRef{
		Name:      "current",
		ScorePath: "artifacts/runs/current/score.json",
	})

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrDuplicateEvidenceRef.Error()) {
		t.Fatalf("Validate() error = %v, want duplicate evidence ref error", err)
	}
}

func TestObjectiveResultRejectsEmptyEvidenceRefNames(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	result.EvidenceRefs[0].Name = ""

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrMalformedEvidenceRef.Error()) {
		t.Fatalf("Validate() error = %v, want malformed evidence ref error", err)
	}
}

func TestObjectiveResultRejectsMalformedEvidenceRefs(t *testing.T) {
	t.Parallel()

	result := sampleObjectiveResult()
	result.EvidenceRefs[0] = ObjectiveEvidenceRef{Name: "current"}

	if err := result.Validate(); err == nil || !strings.Contains(err.Error(), ErrMalformedEvidenceRef.Error()) {
		t.Fatalf("Validate() error = %v, want malformed evidence ref error", err)
	}
}

func TestObjectivePackageAvoidsForbiddenImports(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	dir := filepath.Dir(currentFile)
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, dir, func(info os.FileInfo) bool {
		return info.Name() == "objective.go"
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
		"internal/artifact",
		"materialization",
		"tree-sitter",
		"treesitter",
		"openai",
		"anthropic",
		"cerebras",
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

func sampleObjectiveResult() ObjectiveResult {
	min := 0.0
	max := 1.0

	return ObjectiveResult{
		SchemaVersion: ObjectiveSchemaVersion,
		ObjectiveID:   "candidate_vs_parent_v1",
		EvidenceRefs: []ObjectiveEvidenceRef{
			{
				Name:       "current",
				BundlePath: "artifacts/runs/current",
				ScorePath:  "artifacts/runs/current/score.json",
				SHA256:     "abc123",
			},
			{
				Name:       "parent",
				BundlePath: "artifacts/runs/parent",
				ScorePath:  "artifacts/runs/parent/score.json",
				ReportPath: "artifacts/runs/parent/report.json",
			},
		},
		Values: []ObjectiveValue{
			{Name: "currentLocalizationQuality", Value: 0.82, Kind: ObjectiveValueIntermediate},
			{Name: "parentLocalizationQuality", Value: 0.74, Kind: ObjectiveValueIntermediate},
			{Name: "improvementVsParent", Value: 0.08, Kind: ObjectiveValueIntermediate},
			{Name: "tokenEfficiency", Value: 0.91, Kind: ObjectiveValueIntermediate},
			{Name: "base", Value: 0.77, Kind: ObjectiveValueIntermediate},
			{Name: "regressionPenalty", Value: 1.0, Kind: ObjectiveValuePenalty},
			{Name: "invalidPredictionPenalty", Value: 1.0, Kind: ObjectiveValuePenalty},
			{Name: "final", Value: 0.77, Kind: ObjectiveValueFinal},
		},
		Final:  "final",
		Bounds: &ObjectiveBounds{Min: &min, Max: &max},
	}
}

func findObjectiveValue(t *testing.T, values []ObjectiveValue, name string) ObjectiveValue {
	t.Helper()

	for _, value := range values {
		if value.Name == name {
			return value
		}
	}
	t.Fatalf("objective value %q not found", name)
	return ObjectiveValue{}
}

func setObjectiveValue(t *testing.T, result *ObjectiveResult, name string, value float64) {
	t.Helper()

	for i := range result.Values {
		if result.Values[i].Name == name {
			result.Values[i].Value = value
			return
		}
	}
	t.Fatalf("objective value %q not found", name)
}
