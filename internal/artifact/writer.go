package artifact

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
)

const completeMarkerContent = "complete\n"

// WriteBundle writes one immutable run bundle and returns its completed local
// reference.
func WriteBundle(ctx context.Context, request BundleRequest) (BundleRef, error) {
	return newWriter().WriteBundle(ctx, request)
}

type writer struct {
	now         func() time.Time
	marshalJSON func(any) ([]byte, error)
	rename      func(string, string) error
	writeFile   func(string, []byte) error
	mkdirAll    func(string) error
	removeAll   func(string) error
	afterWrite  func(string)
}

func newWriter() writer {
	return writer{
		now:         func() time.Time { return time.Now().UTC() },
		marshalJSON: marshalDeterministic,
		rename:      os.Rename,
		writeFile: func(path string, data []byte) error {
			return os.WriteFile(path, data, 0o644)
		},
		mkdirAll: func(path string) error {
			return os.MkdirAll(path, 0o755)
		},
		removeAll: os.RemoveAll,
	}
}

func (w writer) WriteBundle(ctx context.Context, request BundleRequest) (BundleRef, error) {
	const (
		phaseValidate = "validate_bundle_request"
		phasePrepare  = "prepare_bundle_directory"
		phaseResolved = "serialize_resolved_input"
		phaseReport   = "serialize_report"
		phaseScore    = "serialize_score_evidence"
		phaseMetadata = "serialize_metadata"
		phaseFinalize = "finalize_bundle"
	)

	if err := ctx.Err(); err != nil {
		return BundleRef{}, &Error{Phase: phaseValidate, Kind: FailureKindValidationFailed, Err: err}
	}

	request = normalizeRequest(w, request)
	if err := validateRequest(request); err != nil {
		return BundleRef{}, &Error{Phase: phaseValidate, Kind: FailureKindValidationFailed, Err: err}
	}

	runsRoot := filepath.Join(string(request.RootPath), "runs")
	finalDir := filepath.Join(runsRoot, request.BundleID)
	stageDir := filepath.Join(runsRoot, "."+request.BundleID+".staging")
	completePath := filepath.Join(finalDir, completeMarkerName)

	if hasCompleteMarker(finalDir) {
		return BundleRef{}, &Error{
			Phase: phasePrepare,
			Kind:  FailureKindAlreadyExists,
			Path:  finalDir,
			Err:   errors.New("completed bundle already exists"),
		}
	}
	if _, err := os.Stat(finalDir); err == nil {
		return BundleRef{}, &Error{
			Phase: phasePrepare,
			Kind:  FailureKindFilesystemFailed,
			Path:  finalDir,
			Err:   errors.New("bundle directory already exists without completion marker"),
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return BundleRef{}, &Error{Phase: phasePrepare, Kind: FailureKindFilesystemFailed, Path: finalDir, Err: err}
	}

	if err := w.mkdirAll(runsRoot); err != nil {
		return BundleRef{}, &Error{Phase: phasePrepare, Kind: FailureKindFilesystemFailed, Path: runsRoot, Err: err}
	}
	_ = w.removeAll(stageDir)
	if err := os.Mkdir(stageDir, 0o755); err != nil {
		return BundleRef{}, &Error{Phase: phasePrepare, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	defer func() {
		_ = w.removeAll(stageDir)
	}()

	files := make([]BundleFile, 0, 6)

	resolvedBytes, err := w.marshalJSON(request.ResolvedInput)
	if err != nil {
		return BundleRef{}, &Error{Phase: phaseResolved, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
	}
	if err := w.writeArtifact(stageDir, "resolved.json", resolvedBytes); err != nil {
		return BundleRef{}, &Error{Phase: phaseResolved, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	files = append(files, fileRecord("resolved", "resolved.json", "application/json", sha256Bytes(resolvedBytes)))

	reportBytes, err := w.marshalJSON(request.CandidateReport)
	if err != nil {
		return BundleRef{}, &Error{Phase: phaseReport, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
	}
	if err := w.writeArtifact(stageDir, "report.json", reportBytes); err != nil {
		return BundleRef{}, &Error{Phase: phaseReport, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	files = append(files, fileRecord("report", "report.json", "application/json", sha256Bytes(reportBytes)))

	if request.RenderedReport != nil {
		rendered := normalizedRenderedReport(*request.RenderedReport)
		renderedBytes := []byte(rendered.Content)
		if err := w.writeArtifact(stageDir, rendered.FileName, renderedBytes); err != nil {
			return BundleRef{}, &Error{Phase: phaseReport, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
		}
		files = append(files, fileRecord("rendered_report", rendered.FileName, rendered.MediaType, sha256Bytes(renderedBytes)))
	}

	scoreEvidence, err := report.ProjectScoreEvidence(request.CandidateReport)
	if err != nil {
		return BundleRef{}, &Error{Phase: phaseScore, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
	}
	scoreBytes, err := w.marshalJSON(scoreEvidence)
	if err != nil {
		return BundleRef{}, &Error{Phase: phaseScore, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
	}
	if err := w.writeArtifact(stageDir, "score.json", scoreBytes); err != nil {
		return BundleRef{}, &Error{Phase: phaseScore, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	files = append(files, fileRecord("score", "score.json", "application/json", sha256Bytes(scoreBytes)))

	completeBytes := []byte(completeMarkerContent)
	metadataFiles := append([]BundleFile(nil), files...)
	metadataFiles = append(metadataFiles,
		fileRecord("metadata", "metadata.json", "application/json", nil),
		fileRecord("complete", completeMarkerName, "text/plain", sha256Bytes(completeBytes)),
	)
	metadata := buildMetadata(request.BundleID, request.CreatedAt, metadataFiles)
	metadataBytes, err := w.marshalJSON(metadata)
	if err != nil {
		return BundleRef{}, &Error{Phase: phaseMetadata, Kind: FailureKindSerializationFailed, Path: stageDir, Err: err}
	}
	if err := w.writeArtifact(stageDir, "metadata.json", metadataBytes); err != nil {
		return BundleRef{}, &Error{Phase: phaseMetadata, Kind: FailureKindFilesystemFailed, Path: stageDir, Err: err}
	}
	files = append(files,
		fileRecord("metadata", "metadata.json", "application/json", nil),
		fileRecord("complete", completeMarkerName, "text/plain", sha256Bytes(completeBytes)),
	)

	if err := ctx.Err(); err != nil {
		return BundleRef{}, &Error{Phase: phaseFinalize, Kind: FailureKindFinalizeFailed, Path: stageDir, Err: err}
	}
	if err := w.rename(stageDir, finalDir); err != nil {
		return BundleRef{}, &Error{Phase: phaseFinalize, Kind: FailureKindFinalizeFailed, Path: finalDir, Err: err}
	}
	if err := w.writeArtifact(finalDir, completeMarkerName, completeBytes); err != nil {
		return BundleRef{}, &Error{Phase: phaseFinalize, Kind: FailureKindFinalizeFailed, Path: completePath, Err: err}
	}

	return BundleRef{
		BundleID:  request.BundleID,
		Path:      domain.HostPath(finalDir),
		Files:     append([]BundleFile(nil), files...),
		CreatedAt: request.CreatedAt.UTC(),
	}, nil
}

func (w writer) writeArtifact(dir string, name string, data []byte) error {
	path := filepath.Join(dir, name)
	if err := w.writeFile(path, data); err != nil {
		return err
	}
	if w.afterWrite != nil {
		w.afterWrite(filepath.ToSlash(name))
	}
	return nil
}

func normalizeRequest(w writer, request BundleRequest) BundleRequest {
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

func validateRequest(request BundleRequest) error {
	if strings.TrimSpace(string(request.RootPath)) == "" {
		return errors.New("bundle root path is required")
	}
	if !safeBundleName(request.BundleID) {
		return fmt.Errorf("bundle id %q is invalid", request.BundleID)
	}
	if err := request.ResolvedInput.Validate(); err != nil {
		return fmt.Errorf("resolved input: %w", err)
	}
	if err := request.CandidateReport.Spec.Validate(); err != nil {
		return fmt.Errorf("candidate report spec: %w", err)
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
		case "report.md", "report.txt":
		default:
			return fmt.Errorf("rendered report file name %q is unsupported", rendered.FileName)
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

func (r ResolvedComparisonInput) Validate() error {
	return report.NewComparisonSpecFromRefs(r.Systems, r.Tasks).Validate()
}
