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
	"github.com/becker63/searchbench-go/internal/score"
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
	ParentRef   *score.ObjectiveEvidenceRef
	PklCommand  []string
}

// Evaluate executes one Pkl scoring file against score evidence and returns
// the validated objective result.
func Evaluate(ctx context.Context, request Request) (score.ObjectiveResult, error) {
	normalized, err := normalizeRequest(request)
	if err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	if err := validateRequest(normalized); err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	currentScorePath, err := resolveScorePath(normalized.CurrentRef.ScorePath)
	if err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: current score path: %w", ErrInvalidRequest, err)
	}
	var parentScorePath string
	if normalized.ParentRef != nil {
		parentScorePath, err = resolveScorePath(normalized.ParentRef.ScorePath)
		if err != nil {
			return score.ObjectiveResult{}, fmt.Errorf("%w: parent score path: %w", ErrInvalidRequest, err)
		}
	}

	evaluator, err := newEvaluator(ctx, normalized.PklCommand)
	if err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: %w", ErrSetupEvaluator, err)
	}
	defer func() { _ = evaluator.Close() }()

	var result score.ObjectiveResult
	if err := evaluator.EvaluateModule(ctx, pkl.TextSource(wrapperModuleSource(normalized, currentScorePath, parentScorePath)), &result); err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: %w", ErrEvaluate, err)
	}
	if err := result.Validate(); err != nil {
		return score.ObjectiveResult{}, fmt.Errorf("%w: %w", ErrValidate, err)
	}
	return result, nil
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
	if err := validateScoreRef("current", request.CurrentRef); err != nil {
		return err
	}
	if request.ParentRef != nil {
		if err := validateScoreRef("parent", *request.ParentRef); err != nil {
			return err
		}
	}
	return nil
}

func resolveScorePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("score path is required")
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

func validateScoreRef(role string, ref score.ObjectiveEvidenceRef) error {
	if strings.TrimSpace(ref.Name) == "" {
		return fmt.Errorf("%s evidence ref name is required", role)
	}
	if strings.TrimSpace(ref.ScorePath) == "" {
		return fmt.Errorf("%s evidence ref score path is required", role)
	}
	return nil
}

func wrapperModuleSource(request Request, currentScorePath string, parentScorePath string) string {
	currentRef := renderEvidenceRef("currentRef", request.CurrentRef)
	currentImport := fmt.Sprintf("import %q as currentScore\n", fileURI(currentScorePath))

	parentImport := ""
	parentAssignment := "parent = null\nparentRef = null\n"
	if request.ParentRef != nil {
		parentImport = fmt.Sprintf("import %q as parentScore\n", fileURI(parentScorePath))
		parentAssignment = "parent = parentScore.toDynamic()\n" + renderEvidenceRef("parentRef", *request.ParentRef)
	}

	return fmt.Sprintf(`amends %q
%s%s
current = currentScore.toDynamic()
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
	if ref.ScorePath != "" {
		lines = append(lines, fmt.Sprintf("  scorePath = %q", ref.ScorePath))
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
