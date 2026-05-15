package scoring

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/apple/pkl-go/pkl"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

var (
	ErrInvalidRequest = errors.New("scoring: invalid request")
	ErrSetupEvaluator = errors.New("scoring: evaluator setup failed")
	ErrEvaluate       = errors.New("scoring: objective evaluation failed")
	ErrValidate       = errors.New("scoring: objective validation failed")
)

// Request is one executable Pkl scoring evaluation.
type Request struct {
	ScoringPath string
	CurrentRef  score.ObjectiveEvidenceRef
	// CurrentEvidencePath overrides the on-disk Pkl module imported for current
	// evidence while preserving CurrentRef as the durable evidence identity.
	CurrentEvidencePath string
	ParentRef           *score.ObjectiveEvidenceRef
	// ParentEvidencePath overrides the on-disk Pkl module imported for parent
	// evidence while preserving ParentRef as the durable evidence identity.
	ParentEvidencePath string
	PklCommand         []string
}

// Evaluate executes one Pkl scoring file against round evidence and returns
// the validated objective result.
func Evaluate(ctx context.Context, request Request) (score.ObjectiveResult, error) {
	normalized, err := normalizeRequest(request)
	if err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	if err := validateRequest(normalized); err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	currentEvidencePath, err := resolveEvidencePath(firstNonEmpty(normalized.CurrentEvidencePath, normalized.CurrentRef.EvidencePath))
	if err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: current evidence path: %w", ErrInvalidRequest, err)
	}
	var parentEvidencePath string
	if normalized.ParentRef != nil {
		parentEvidencePath, err = resolveEvidencePath(firstNonEmpty(normalized.ParentEvidencePath, normalized.ParentRef.EvidencePath))
		if err != nil {
			return score.ObjectiveResult{}, fmt.Errorf("%w: parent evidence path: %w", ErrInvalidRequest, err)
		}
	}

	evaluator, err := newEvaluator(ctx, normalized.PklCommand)
	if err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: %w", ErrSetupEvaluator, err)
	}
	defer func() { _ = evaluator.Close() }()

	var result score.ObjectiveResult
	if err := evaluator.EvaluateModule(ctx, pkl.TextSource(wrapperModuleSource(normalized, currentEvidencePath, parentEvidencePath)), &result); err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: %w", ErrEvaluate, err)
	}
	fillObjectiveValueKinds(&result)
	if err := result.Validate(); err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: %w", ErrValidate, err)
	}
	return result, nil
}

// fillObjectiveValueKinds sets ObjectiveValueKind when Pkl decoding leaves it empty
// (some evaluator paths omit the kind field in JSON even when Pkl populated it).
func fillObjectiveValueKinds(r *score.ObjectiveResult) {
	if r == nil {
		return
	}
	finalName := strings.TrimSpace(r.Final)
	for i := range r.Values {
		if r.Values[i].Kind != "" {
			continue
		}
		name := strings.TrimSpace(r.Values[i].Name)
		switch {
		case finalName != "" && name == finalName:
			r.Values[i].Kind = score.ObjectiveValueFinal
		case strings.HasSuffix(name, "Penalty"):
			r.Values[i].Kind = score.ObjectiveValuePenalty
		default:
			r.Values[i].Kind = score.ObjectiveValueIntermediate
		}
	}
}

func normalizeRequest(request Request) (Request, error) {
	if strings.TrimSpace(request.ScoringPath) == "" {
		return request, errors.New("scoring path is required")
	}
	if filepath.IsAbs(request.ScoringPath) {
		return request, nil
	}
	abs, err := filepath.Abs(request.ScoringPath)
	if err != nil {
		return request, fmt.Errorf("resolve scoring path: %w", err)
	}
	request.ScoringPath = abs
	return request, nil
}

func newEvaluator(ctx context.Context, pklCommand []string) (pkl.Evaluator, error) {
	options := []func(*pkl.EvaluatorOptions){pkl.PreconfiguredOptions}
	if len(pklCommand) == 0 {
		return pkl.NewEvaluator(ctx, options...)
	}
	return pkl.NewEvaluatorWithCommand(ctx, append([]string(nil), pklCommand...), options...)
}

func validateRequest(request Request) error {
	if _, err := os.Stat(request.ScoringPath); err != nil {
		return fmt.Errorf("scoring path: %w", err)
	}
	if err := validateEvidenceRef("current", request.CurrentRef); err != nil {
		return err
	}
	if request.ParentRef != nil {
		if err := validateEvidenceRef("parent", *request.ParentRef); err != nil {
			return err
		}
	}
	return nil
}

func resolveEvidencePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("evidence path is required")
	}
	if filepath.IsAbs(path) {
		if _, err := os.Stat(path); err != nil {
			return "", err
		}
		return path, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		return "", err
	}
	return abs, nil
}

func validateEvidenceRef(role string, ref score.ObjectiveEvidenceRef) error {
	if strings.TrimSpace(ref.Name) == "" {
		return fmt.Errorf("%s evidence ref name is required", role)
	}
	if strings.TrimSpace(ref.EvidencePath) == "" {
		return fmt.Errorf("%s evidence ref evidence path is required", role)
	}
	return nil
}

func wrapperModuleSource(request Request, currentEvidencePath string, parentEvidencePath string) string {
	currentRef := renderEvidenceRef("currentRef", request.CurrentRef)
	currentImport := fmt.Sprintf("import %q as currentEvidence\n", fileURI(currentEvidencePath))

	parentImport := ""
	parentAssignment := "parent = null\nparentRef = null\n"
	if request.ParentRef != nil {
		parentImport = fmt.Sprintf("import %q as parentEvidence\n", fileURI(parentEvidencePath))
		parentAssignment = "parent = parentEvidence.toDynamic()\n" + renderEvidenceRef("parentRef", *request.ParentRef)
	}

	return fmt.Sprintf(`amends %q
%s%s
current = currentEvidence.toDynamic()
%s
%s`, fileURI(request.ScoringPath), currentImport, parentImport, currentRef, parentAssignment)
}

func renderEvidenceRef(name string, ref score.ObjectiveEvidenceRef) string {
	lines := []string{
		fmt.Sprintf("%s = new ObjectiveEvidenceRef {", name),
		fmt.Sprintf("  name = %q", ref.Name),
	}
	if ref.BundlePath != "" {
		lines = append(lines, fmt.Sprintf("  bundlePath = %q", ref.BundlePath))
	}
	if ref.EvidencePath != "" {
		lines = append(lines, fmt.Sprintf("  evidencePath = %q", ref.EvidencePath))
	}
	if ref.ReportPath != "" {
		lines = append(lines, fmt.Sprintf("  reportPath = %q", ref.ReportPath))
	}
	if ref.SHA256 != "" {
		lines = append(lines, fmt.Sprintf("  sha256 = %q", ref.SHA256))
	}
	lines = append(lines, "}\n")
	return strings.Join(lines, "\n")
}

func fileURI(path string) string {
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
