package localrun

import "fmt"

type Phase string

const (
	PhaseLoadManifestFailed     Phase = "load_manifest_failed"
	PhaseValidateManifestFailed Phase = "validate_manifest_failed"
	PhaseUnsupportedMode        Phase = "unsupported_mode"
	PhaseProjectFakePlanFailed  Phase = "project_fake_plan_failed"
	PhaseFakeComparisonFailed   Phase = "fake_comparison_failed"
	PhaseScoreEvidenceFailed    Phase = "score_evidence_failed"
	PhaseScorePKLFailed         Phase = "score_pkl_failed"
	PhaseObjectiveFailed        Phase = "objective_failed"
	PhaseRenderReportFailed     Phase = "render_report_failed"
	PhaseBundleWriteFailed      Phase = "bundle_write_failed"
)

// Error is the phase-tagged local manifest-run failure.
type Error struct {
	Phase Phase
	Err   error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err == nil {
		return string(e.Phase)
	}
	return fmt.Sprintf("%s: %v", e.Phase, e.Err)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
