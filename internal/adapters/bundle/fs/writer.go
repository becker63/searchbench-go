package bundlefs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

const completeMarkerContent = "complete\n"

// WriteBundle writes one immutable round bundle and returns its completed local
// reference.
func WriteBundle(ctx context.Context, request RoundBundleInput) (RoundBundleRef, error) {
	return newWriter().WriteBundle(ctx, request)
}

type writer struct {
	now                func() time.Time
	marshalJSON        func(any) ([]byte, error)
	marshalEvidencePKL func(score.RoundEvidenceDocument) ([]byte, error)
	rename             func(string, string) error
	writeFile          func(string, []byte) error
	mkdirAll           func(string) error
	removeAll          func(string) error
	afterWrite         func(string)
}

func newWriter() writer {
	return writer{
		now:                func() time.Time { return time.Now().UTC() },
		marshalJSON:        marshalDeterministic,
		marshalEvidencePKL: marshalRoundEvidencePkl,
		rename:             os.Rename,
		writeFile: func(path string, data []byte) error {
			return os.WriteFile(path, data, 0o644)
		},
		mkdirAll: func(path string) error {
			return os.MkdirAll(path, 0o755)
		},
		removeAll: os.RemoveAll,
	}
}

func (w writer) WriteBundle(ctx context.Context, request RoundBundleInput) (RoundBundleRef, error) {
	const (
		phaseValidate        = "validate_bundle_request"
		phasePrepare         = "prepare_bundle_directory"
		phaseResolved        = "serialize_resolved_round"
		phaseReport          = "serialize_report"
		phaseEvidence        = "serialize_round_evidence"
		phaseObjective       = "serialize_objective_result"
		phaseDecision        = "serialize_decision"
		phaseArtifacts       = "serialize_bundle_artifacts"
		phaseContinuation    = "serialize_continuation"
		phaseContinuationPKL = "serialize_continuation_pkl"
		phaseMetadata        = "serialize_metadata"
		phaseFinalize        = "finalize_bundle"
	)

	if err := ctx.Err(); err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseValidate, Kind: FailureKindValidationFailed, Err: err}
	}

	request = normalizeRequest(w, request)
	if err := validateRequest(request); err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseValidate, Kind: FailureKindValidationFailed, Err: err}
	}

	roundsRoot := roundBundleCollectionRoot(request)
	finalDir := filepath.Join(roundsRoot, request.BundleID)
	stageDir := filepath.Join(roundsRoot, "."+request.BundleID+".staging")
	completePath := filepath.Join(finalDir, completeMarkerName)

	if hasCompleteMarker(finalDir) {
		return RoundBundleRef{}, &Error{
			Phase: phasePrepare,
			Kind:  FailureKindAlreadyExists,
			Path:  finalDir,
			Err:   errors.New("completed bundle already exists"),
		}
	}
	if _, err := os.Stat(finalDir); err == nil {
		return RoundBundleRef{}, &Error{
			Phase: phasePrepare,
			Kind:  FailureKindFilesystemFailed,
			Path:  finalDir,
			Err:   errors.New("bundle directory already exists without completion marker"),
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return RoundBundleRef{}, &Error{Phase: phasePrepare, Kind: FailureKindFilesystemFailed, Path: finalDir, Err: err}
	}

	if err := w.mkdirAll(roundsRoot); err != nil {
		return RoundBundleRef{}, &Error{Phase: phasePrepare, Kind: FailureKindFilesystemFailed, Path: roundsRoot, Err: err}
	}
	_ = w.removeAll(stageDir)
	if err := os.Mkdir(stageDir, 0o755); err != nil {
		return RoundBundleRef{}, &Error{Phase: phasePrepare, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	defer func() {
		_ = w.removeAll(stageDir)
	}()

	files := make([]BundleFile, 0, 10+len(request.AdditionalFiles))

	resolvedBytes, err := w.marshalJSON(request.ResolvedInput)
	if err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseResolved, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
	}
	if err := w.writeArtifact(stageDir, "resolved-round.json", resolvedBytes); err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseResolved, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	files = append(files, fileRecord("resolved_round", "resolved-round.json", "application/json", sha256Bytes(resolvedBytes)))

	reportBytes, err := w.marshalJSON(request.RoundReport)
	if err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseReport, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
	}
	if err := w.writeArtifact(stageDir, "round-report.json", reportBytes); err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseReport, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	files = append(files, fileRecord("round_report", "round-report.json", "application/json", sha256Bytes(reportBytes)))

	if request.RenderedReport != nil {
		rendered := normalizedRenderedReport(*request.RenderedReport)
		renderedBytes := []byte(rendered.Content)
		if err := w.writeArtifact(stageDir, rendered.FileName, renderedBytes); err != nil {
			return RoundBundleRef{}, &Error{Phase: phaseReport, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
		}
		files = append(files, fileRecord("rendered_report", rendered.FileName, rendered.MediaType, sha256Bytes(renderedBytes)))
	}

	evidenceBytes, err := w.marshalEvidencePKL(request.RoundEvidence)
	if err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseEvidence, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
	}
	if err := w.writeArtifact(stageDir, "evidence.pkl", evidenceBytes); err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseEvidence, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	files = append(files, fileRecord("round_evidence", "evidence.pkl", "text/plain", sha256Bytes(evidenceBytes)))

	decisionBytes, err := w.marshalJSON(request.RoundReport.Decision)
	if err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseDecision, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
	}
	if err := w.writeArtifact(stageDir, "decision.json", decisionBytes); err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseDecision, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	files = append(files, fileRecord("decision", "decision.json", "application/json", sha256Bytes(decisionBytes)))

	if request.ObjectiveResult != nil {
		objectiveBytes, err := w.marshalJSON(*request.ObjectiveResult)
		if err != nil {
			return RoundBundleRef{}, &Error{Phase: phaseObjective, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
		}
		if err := w.writeArtifact(stageDir, "objective.json", objectiveBytes); err != nil {
			return RoundBundleRef{}, &Error{Phase: phaseObjective, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
		}
		files = append(files, fileRecord("objective", "objective.json", "application/json", sha256Bytes(objectiveBytes)))
	}

	for _, artifact := range request.AdditionalFiles {
		if err := w.writeArtifact(stageDir, artifact.Path, artifact.Content); err != nil {
			return RoundBundleRef{}, &Error{Phase: phaseArtifacts, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
		}
		files = append(files, fileRecord(artifact.Kind, artifact.Path, normalizedArtifactMediaType(artifact.MediaType), sha256Bytes(artifact.Content)))
	}

	if request.Continuation != nil {
		continuationBytes, err := w.marshalJSON(*request.Continuation)
		if err != nil {
			return RoundBundleRef{}, &Error{Phase: phaseContinuation, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
		}
		if err := w.writeArtifact(stageDir, "continuation.json", continuationBytes); err != nil {
			return RoundBundleRef{}, &Error{Phase: phaseContinuation, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
		}
		files = append(files, fileRecord("continuation", "continuation.json", "application/json", sha256Bytes(continuationBytes)))
	}
	if request.ContinuationPKL != nil {
		continuationPKLBytes, err := renderContinuationPKL(finalDir, *request.Continuation, *request.ContinuationPKL)
		if err != nil {
			return RoundBundleRef{}, &Error{Phase: phaseContinuationPKL, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
		}
		if err := w.writeArtifact(stageDir, continuationPKLFileName, continuationPKLBytes); err != nil {
			return RoundBundleRef{}, &Error{Phase: phaseContinuationPKL, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
		}
		files = append(files, fileRecord("continuation_pkl", continuationPKLFileName, "text/plain", sha256Bytes(continuationPKLBytes)))
	}

	completeBytes := []byte(completeMarkerContent)
	metadataFiles := append([]BundleFile(nil), files...)
	metadataFiles = append(metadataFiles,
		fileRecord("metadata", "metadata.json", "application/json", nil),
		fileRecord("complete", completeMarkerName, "text/plain", sha256Bytes(completeBytes)),
	)
	metadata := buildMetadata(request.BundleID, request.CreatedAt, metadataFiles)
	metadataBytes, err := w.marshalJSON(metadata)
	if err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseMetadata, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
	}
	if err := w.writeArtifact(stageDir, "metadata.json", metadataBytes); err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseMetadata, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	files = append(files,
		fileRecord("metadata", "metadata.json", "application/json", nil),
		fileRecord("complete", completeMarkerName, "text/plain", sha256Bytes(completeBytes)),
	)

	if err := ctx.Err(); err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseFinalize, Kind: FailureKindFinalizeFailed, Path: stageDir, Err: err}
	}
	if err := w.rename(stageDir, finalDir); err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseFinalize, Kind: FailureKindFinalizeFailed, Path: finalDir, Err: err}
	}
	if err := w.writeArtifact(finalDir, completeMarkerName, completeBytes); err != nil {
		return RoundBundleRef{}, &Error{Phase: phaseFinalize, Kind: FailureKindFinalizeFailed, Path: completePath, Err: err}
	}

	return RoundBundleRef{
		BundleID:  request.BundleID,
		Path:      domain.HostPath(finalDir),
		Files:     append([]BundleFile(nil), files...),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func (w writer) writeArtifact(dir string, name string, data []byte) error {
	path := filepath.Join(dir, name)
	if err := w.mkdirAll(filepath.Dir(path)); err != nil {
		return err
	}
	if err := w.writeFile(path, data); err != nil {
		return err
	}
	if w.afterWrite != nil {
		w.afterWrite(filepath.ToSlash(name))
	}
	return nil
}

func roundBundleCollectionRoot(request RoundBundleInput) string {
	gameID := strings.TrimSpace(request.RoundEvidence.GameID)
	if gameID == "" {
		gameID = "code-localization"
	}
	return filepath.Join(string(request.RootPath), "games", safePathElement(gameID), "rounds")
}

func safePathElement(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, string(filepath.Separator), "-")
	value = strings.ReplaceAll(value, "/", "-")
	if value == "" {
		return "unknown"
	}
	return value
}

func normalizeRequest(w writer, request RoundBundleInput) RoundBundleInput {
	if request.CreatedAt.IsZero() {
		request.CreatedAt = w.now()
	}
	request.CreatedAt = request.CreatedAt.UTC()
	if request.RenderedReport != nil {
		rendered := normalizedRenderedReport(*request.RenderedReport)
		request.RenderedReport = &rendered
	}
	return request
}

func validateRequest(request RoundBundleInput) error {
	if strings.TrimSpace(string(request.RootPath)) == "" {
		return errors.New("bundle root path is required")
	}
	if !safeBundleName(request.BundleID) {
		return fmt.Errorf("bundle id %q is invalid", request.BundleID)
	}
	if request.ResolvedInput == nil {
		return errors.New("resolved input is required")
	}
	if err := request.RoundReport.Spec.Validate(); err != nil {
		return fmt.Errorf("round report spec: %w", err)
	}
	if err := request.RoundEvidence.Validate(); err != nil {
		return fmt.Errorf("round evidence: %w", err)
	}
	if request.ObjectiveResult != nil {
		if err := request.ObjectiveResult.Validate(); err != nil {
			return fmt.Errorf("objective result: %w", err)
		}
	}
	if request.RenderedReport != nil {
		rendered := normalizedRenderedReport(*request.RenderedReport)
		if rendered.Content == "" {
			return errors.New("rendered report content is required when rendered report is provided")
		}
		if !safeBundleName(rendered.FileName) {
			return fmt.Errorf("rendered report file name %q is invalid", rendered.FileName)
		}
		switch rendered.FileName {
		case "round-report.md", "round-report.txt":
		default:
			return fmt.Errorf("rendered report file name %q is unsupported", rendered.FileName)
		}
	}
	if request.Continuation != nil {
		if err := request.Continuation.Validate(); err != nil {
			return fmt.Errorf("continuation: %w", err)
		}
	}
	if request.ContinuationPKL != nil {
		if request.Continuation == nil {
			return errors.New("continuation.pkl requires continuation.json")
		}
		if strings.TrimSpace(request.ContinuationPKL.SchemaPath) == "" {
			return errors.New("continuation.pkl schema path is required")
		}
		if strings.TrimSpace(request.ContinuationPKL.HelpersPath) == "" {
			return errors.New("continuation.pkl helpers path is required")
		}
	}
	for _, artifact := range request.AdditionalFiles {
		if strings.TrimSpace(artifact.Kind) == "" {
			return errors.New("bundle artifact kind is required")
		}
		if err := validateBundleArtifact(artifact); err != nil {
			return err
		}
	}
	return nil
}

func normalizedRenderedReport(rendered RenderedReport) RenderedReport {
	if rendered.FileName == "" {
		rendered.FileName = defaultRenderedReport
	}
	if rendered.MediaType == "" {
		switch rendered.FileName {
		case "report.md":
			rendered.MediaType = "text/markdown"
		default:
			rendered.MediaType = "text/plain"
		}
	}
	return rendered
}

func normalizedArtifactMediaType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "application/octet-stream"
	}
	return value
}

func validateBundleArtifact(artifact BundleArtifact) error {
	path := filepath.ToSlash(filepath.Clean(strings.TrimSpace(artifact.Path)))
	if path == "" || path == "." || path == ".." {
		return fmt.Errorf("bundle artifact path %q is invalid", artifact.Path)
	}
	if strings.HasPrefix(path, "../") || strings.Contains(path, "/../") {
		return fmt.Errorf("bundle artifact path %q is invalid", artifact.Path)
	}
	if filepath.IsAbs(artifact.Path) {
		return fmt.Errorf("bundle artifact path %q is invalid", artifact.Path)
	}
	if len(artifact.Content) == 0 {
		return fmt.Errorf("bundle artifact %q content is required", artifact.Path)
	}
	return nil
}

func safeBundleName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return false
	}
	return !strings.ContainsAny(name, `/\`)
}

func hasCompleteMarker(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, completeMarkerName))
	return err == nil && !info.IsDir()
}
