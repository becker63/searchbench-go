package score

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const ObjectiveSchemaVersion = "searchbench.objective_result.v1"

var (
	ErrObjectiveValidationFailed = errors.New("score: objective validation failed")
	ErrMissingObjectiveID        = errors.New("score: missing objective id")
	ErrMissingSchemaVersion      = errors.New("score: missing schema version")
	ErrMissingFinalValue         = errors.New("score: missing final value")
	ErrFinalValueNotFound        = errors.New("score: final value not found")
	ErrDuplicateObjectiveValue   = errors.New("score: duplicate objective value name")
	ErrDuplicateEvidenceRef      = errors.New("score: duplicate objective evidence ref name")
	ErrInvalidNumericValue       = errors.New("score: invalid numeric value")
	ErrMalformedEvidenceRef      = errors.New("score: malformed objective evidence ref")
	ErrFinalValueOutOfBounds     = errors.New("score: final value outside declared bounds")
)

// ObjectiveResult is the typed, reviewable result of a future visible
// objective calculation.
//
// It stores named intermediate values and identifies the final value by name.
// The model intentionally does not encode how those values were produced.
type ObjectiveResult struct {
	SchemaVersion string                 `json:"schema_version" pkl:"schemaVersion"`
	ObjectiveID   string                 `json:"objective_id" pkl:"objectiveId"`
	EvidenceRefs  []ObjectiveEvidenceRef `json:"evidence_refs,omitempty" pkl:"evidenceRefs"`
	Values        []ObjectiveValue       `json:"values" pkl:"values"`
	Final         string                 `json:"final" pkl:"final"`
	Bounds        *ObjectiveBounds       `json:"bounds,omitempty" pkl:"bounds"`
}

// ObjectiveValue is one named output from a future objective calculation.
type ObjectiveValue struct {
	Name        string             `json:"name" pkl:"name"`
	Value       float64            `json:"value" pkl:"value"`
	Kind        ObjectiveValueKind `json:"kind" pkl:"kind"`
	Unit        string             `json:"unit,omitempty" pkl:"unit"`
	Description string             `json:"description,omitempty" pkl:"description"`
}

// ObjectiveValueKind classifies one objective value.
type ObjectiveValueKind string

const (
	ObjectiveValueIntermediate ObjectiveValueKind = "intermediate"
	ObjectiveValuePenalty      ObjectiveValueKind = "penalty"
	ObjectiveValueFinal        ObjectiveValueKind = "final"
)

// ObjectiveEvidenceRef is a typed reference to evidence used by a future
// objective calculation.
type ObjectiveEvidenceRef struct {
	Name       string `json:"name" pkl:"name"`
	BundlePath string `json:"bundle_path,omitempty" pkl:"bundlePath"`
	ScorePath  string `json:"score_path,omitempty" pkl:"scorePath"`
	ReportPath string `json:"report_path,omitempty" pkl:"reportPath"`
	SHA256     string `json:"sha256,omitempty" pkl:"sha256"`
}

// ObjectiveBounds optionally constrains the final objective value.
type ObjectiveBounds struct {
	Min *float64 `json:"min,omitempty" pkl:"min"`
	Max *float64 `json:"max,omitempty" pkl:"max"`
}

// Validate checks that the objective result is structurally meaningful and
// numerically safe to persist later.
func (r ObjectiveResult) Validate() error {
	if strings.TrimSpace(r.SchemaVersion) == "" {
		return fmt.Errorf("%w: %w", ErrObjectiveValidationFailed, ErrMissingSchemaVersion)
	}
	if strings.TrimSpace(r.ObjectiveID) == "" {
		return fmt.Errorf("%w: %w", ErrObjectiveValidationFailed, ErrMissingObjectiveID)
	}
	if err := validateObjectiveEvidenceRefs(r.EvidenceRefs); err != nil {
		return fmt.Errorf("%w: %w", ErrObjectiveValidationFailed, err)
	}
	if len(r.Values) == 0 {
		return fmt.Errorf("%w: %w", ErrObjectiveValidationFailed, ErrMissingFinalValue)
	}

	valuesByName := make(map[string]ObjectiveValue, len(r.Values))
	for _, value := range r.Values {
		name := strings.TrimSpace(value.Name)
		if name == "" {
			return fmt.Errorf("%w: objective value name is required", ErrObjectiveValidationFailed)
		}
		if _, exists := valuesByName[name]; exists {
			return fmt.Errorf("%w: %w: %s", ErrObjectiveValidationFailed, ErrDuplicateObjectiveValue, name)
		}
		if !isFinite(value.Value) {
			return fmt.Errorf("%w: %w: %s", ErrObjectiveValidationFailed, ErrInvalidNumericValue, name)
		}
		valuesByName[name] = value
	}

	finalName := strings.TrimSpace(r.Final)
	if finalName == "" {
		return fmt.Errorf("%w: %w", ErrObjectiveValidationFailed, ErrMissingFinalValue)
	}
	finalValue, ok := valuesByName[finalName]
	if !ok {
		return fmt.Errorf("%w: %w: %s", ErrObjectiveValidationFailed, ErrFinalValueNotFound, finalName)
	}
	if !isFinite(finalValue.Value) {
		return fmt.Errorf("%w: %w: %s", ErrObjectiveValidationFailed, ErrInvalidNumericValue, finalName)
	}
	if err := validateObjectiveBounds(r.Bounds, finalValue.Value); err != nil {
		return fmt.Errorf("%w: %w", ErrObjectiveValidationFailed, err)
	}
	return nil
}

// FinalValue returns the named final value when present.
func (r ObjectiveResult) FinalValue() (ObjectiveValue, bool) {
	finalName := strings.TrimSpace(r.Final)
	if finalName == "" {
		return ObjectiveValue{}, false
	}
	for _, value := range r.Values {
		if strings.TrimSpace(value.Name) == finalName {
			return value, true
		}
	}
	return ObjectiveValue{}, false
}

func validateObjectiveEvidenceRefs(refs []ObjectiveEvidenceRef) error {
	seen := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		name := strings.TrimSpace(ref.Name)
		if name == "" {
			return fmt.Errorf("%w: evidence ref name is required", ErrMalformedEvidenceRef)
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("%w: %w: %s", ErrMalformedEvidenceRef, ErrDuplicateEvidenceRef, name)
		}
		seen[name] = struct{}{}

		hasLocator := false
		for _, locator := range []string{ref.BundlePath, ref.ScorePath, ref.ReportPath, ref.SHA256} {
			if strings.TrimSpace(locator) != "" {
				hasLocator = true
				break
			}
		}
		if !hasLocator {
			return fmt.Errorf("%w: evidence ref %q requires at least one locator", ErrMalformedEvidenceRef, name)
		}
	}
	return nil
}

func validateObjectiveBounds(bounds *ObjectiveBounds, finalValue float64) error {
	if bounds == nil {
		return nil
	}
	if bounds.Min != nil {
		if !isFinite(*bounds.Min) {
			return ErrInvalidNumericValue
		}
		if finalValue < *bounds.Min {
			return ErrFinalValueOutOfBounds
		}
	}
	if bounds.Max != nil {
		if !isFinite(*bounds.Max) {
			return ErrInvalidNumericValue
		}
		if finalValue > *bounds.Max {
			return ErrFinalValueOutOfBounds
		}
	}
	if bounds.Min != nil && bounds.Max != nil && *bounds.Min > *bounds.Max {
		return ErrFinalValueOutOfBounds
	}
	return nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
