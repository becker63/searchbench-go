// Package bundle persists optimizer artifacts to disk (staging directory,
// atomic rename, COMPLETE marker on success). All os.File/os.Mkdir effects live here.
package bundle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

const (
	bundleSchemaVersion   = "searchbench.optimizer_bundle.v1"
	completeMarkerContent = "complete\n"
)

// Request describes one optimizer bundle to persist.
type Request struct {
	BundleCollection string
	BundleID         string
	CreatedAt        time.Time
	ParentBundle     string
	OutputArtifact   string
	Resolved         ResolvedDocument
	Result           pureoptimizer.NextChallengerRecord
}

// ResolvedDocument is the manifest-derived plan summary the bundle persists
// alongside the optimizer result.
type ResolvedDocument struct {
	ManifestPath     string
	RoundName        string
	Mode             string
	ParentRound      pureoptimizer.ParentRoundRef
	Target           pureoptimizer.NextChallengerTarget
	Agent            pureoptimizer.AgentConfig
	IncludedEvidence []string
	DeniedEvidence   []string
	InputPolicyPath  string
	OutputBundlePath string
}

// WriteBundle stages the optimizer bundle in a sibling directory and
// atomically renames it into place. It writes the COMPLETE marker only on a
// successful proposal.
func WriteBundle(ctx context.Context, req Request) (string, error) {
	runsRoot := req.BundleCollection
	finalDir := filepath.Join(runsRoot, req.BundleID)
	stageDir := filepath.Join(runsRoot, "."+req.BundleID+".staging")

	if err := ctx.Err(); err != nil {
		return "", err
	}
	if _, err := os.Stat(finalDir); err == nil {
		return "", fmt.Errorf("optimizer bundle already exists: %s", finalDir)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}

	if err := os.MkdirAll(runsRoot, 0o755); err != nil {
		return "", err
	}
	_ = os.RemoveAll(stageDir)
	if err := os.Mkdir(stageDir, 0o755); err != nil {
		return "", err
	}
	defer os.RemoveAll(stageDir)

	files := make([]bundleFile, 0, 8)
	writeJSON := func(name string, value any) error {
		data, err := marshalDeterministic(value)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(stageDir, name), data, 0o644); err != nil {
			return err
		}
		files = append(files, bundleFile{
			Kind:      strings.TrimSuffix(filepath.Base(name), filepath.Ext(name)),
			Path:      filepath.ToSlash(name),
			MediaType: "application/json",
			SHA256:    sha256String(data),
		})
		return nil
	}

	if err := writeJSON("resolved-next-challenger.json", resolvedDocumentJSON(req.Resolved)); err != nil {
		return "", err
	}
	if req.Result.RenderedPrompt != "" {
		data := []byte(req.Result.RenderedPrompt)
		if err := os.WriteFile(filepath.Join(stageDir, "optimizer_prompt.txt"), data, 0o644); err != nil {
			return "", err
		}
		files = append(files, bundleFile{
			Kind:      "optimizer_prompt",
			Path:      "optimizer_prompt.txt",
			MediaType: "text/plain",
			SHA256:    sha256String(data),
		})
	}
	if err := writeJSON("optimizer_result.json", nextChallengerRecordDocument(req.Result)); err != nil {
		return "", err
	}
	for _, attempt := range req.Result.Attempts {
		if attempt.RenderedPrompt != "" {
			name := fmt.Sprintf("attempts/attempt-%03d-prompt.txt", attempt.Number)
			data := []byte(attempt.RenderedPrompt)
			if err := os.MkdirAll(filepath.Dir(filepath.Join(stageDir, name)), 0o755); err != nil {
				return "", err
			}
			if err := os.WriteFile(filepath.Join(stageDir, name), data, 0o644); err != nil {
				return "", err
			}
			files = append(files, bundleFile{
				Kind:      "attempt_prompt",
				Path:      filepath.ToSlash(name),
				MediaType: "text/plain",
				SHA256:    sha256String(data),
			})
		}
		name := fmt.Sprintf("attempts/attempt-%03d-result.json", attempt.Number)
		if err := os.MkdirAll(filepath.Dir(filepath.Join(stageDir, name)), 0o755); err != nil {
			return "", err
		}
		if err := writeAttemptJSON(stageDir, name, attempt, &files); err != nil {
			return "", err
		}
	}
	if req.Result.Proposal != nil && req.Result.Success {
		data := []byte(req.Result.Proposal.Code)
		if err := os.WriteFile(filepath.Join(stageDir, req.Result.Proposal.ArtifactName), data, 0o644); err != nil {
			return "", err
		}
		files = append(files, bundleFile{
			Kind:      "next_challenger",
			Path:      filepath.ToSlash(req.Result.Proposal.ArtifactName),
			MediaType: "text/x-python",
			SHA256:    sha256String(data),
		})
	}

	metadata := bundleMetadata{
		SchemaVersion:  bundleSchemaVersion,
		BundleID:       req.BundleID,
		CreatedAt:      req.CreatedAt.UTC(),
		FinalStatus:    bundleStatus(req.Result),
		ParentBundle:   req.ParentBundle,
		OutputArtifact: req.OutputArtifact,
		Files:          append([]bundleFile(nil), files...),
	}
	metadataBytes, err := marshalDeterministic(metadata)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(stageDir, "metadata.json"), metadataBytes, 0o644); err != nil {
		return "", err
	}
	files = append(files, bundleFile{
		Kind:      "metadata",
		Path:      "metadata.json",
		MediaType: "application/json",
		SHA256:    sha256String(metadataBytes),
	})
	if req.Result.Success {
		completeBytes := []byte(completeMarkerContent)
		if err := os.WriteFile(filepath.Join(stageDir, "COMPLETE"), completeBytes, 0o644); err != nil {
			return "", err
		}
		files = append(files, bundleFile{
			Kind:      "complete",
			Path:      "COMPLETE",
			MediaType: "text/plain",
			SHA256:    sha256String(completeBytes),
		})
	}

	if err := os.Rename(stageDir, finalDir); err != nil {
		return "", err
	}
	return finalDir, nil
}

type resolvedPlanDocument struct {
	ManifestPath     string                             `json:"manifest_path"`
	RoundName        string                             `json:"round_name"`
	Mode             string                             `json:"mode"`
	ParentRound      pureoptimizer.ParentRoundRef       `json:"parent_round"`
	Target           pureoptimizer.NextChallengerTarget `json:"target"`
	Agent            pureoptimizer.AgentConfig          `json:"agent"`
	IncludedEvidence []string                           `json:"included_evidence,omitempty"`
	DeniedEvidence   []string                           `json:"denied_evidence,omitempty"`
	InputPolicyPath  string                             `json:"input_policy_path,omitempty"`
	OutputBundlePath string                             `json:"output_bundle_path,omitempty"`
}

type attemptDocument struct {
	Number                 int                                   `json:"number"`
	State                  pureoptimizer.AttemptState            `json:"state"`
	RawOutput              string                                `json:"raw_output,omitempty"`
	Proposal               *pureoptimizer.NextChallengerProposal `json:"proposal,omitempty"`
	Failure                *pureoptimizer.Failure                `json:"failure,omitempty"`
	PipelineResults        []stepResultDocument                  `json:"pipeline_results,omitempty"`
	PipelineClassification *classificationDocument               `json:"pipeline_classification,omitempty"`
	RetryFeedback          string                                `json:"retry_feedback,omitempty"`
}

type resultDocument struct {
	Success  bool                                  `json:"success"`
	Proposal *pureoptimizer.NextChallengerProposal `json:"proposal,omitempty"`
	Failure  *pureoptimizer.Failure                `json:"failure,omitempty"`
	Attempts []attemptDocument                     `json:"attempts,omitempty"`
	Phases   []pureoptimizer.Phase                 `json:"phases,omitempty"`
}

type bundleMetadata struct {
	SchemaVersion  string       `json:"schema_version"`
	BundleID       string       `json:"bundle_id"`
	CreatedAt      time.Time    `json:"created_at"`
	FinalStatus    string       `json:"final_status"`
	ParentBundle   string       `json:"parent_bundle,omitempty"`
	OutputArtifact string       `json:"output_artifact,omitempty"`
	Files          []bundleFile `json:"files"`
}

type bundleFile struct {
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	MediaType string `json:"media_type"`
	SHA256    string `json:"sha256,omitempty"`
}

type stepResultDocument struct {
	Name                string   `json:"name"`
	Command             []string `json:"command,omitempty"`
	CWD                 string   `json:"cwd,omitempty"`
	Passed              bool     `json:"passed,omitempty"`
	ExitCode            int      `json:"exit_code"`
	Stdout              string   `json:"stdout,omitempty"`
	Stderr              string   `json:"stderr,omitempty"`
	Skipped             bool     `json:"skipped,omitempty"`
	TimedOut            bool     `json:"timed_out,omitempty"`
	InfrastructureError string   `json:"infrastructure_error,omitempty"`
}

type classificationDocument struct {
	GenerationFailures     []stepResultDocument `json:"generation_failures,omitempty"`
	FormatErrors           []stepResultDocument `json:"format_errors,omitempty"`
	TypeErrors             []stepResultDocument `json:"type_errors,omitempty"`
	LintErrors             []stepResultDocument `json:"lint_errors,omitempty"`
	TestFailures           []stepResultDocument `json:"test_failures,omitempty"`
	InfrastructureFailures []stepResultDocument `json:"infrastructure_failures,omitempty"`
	PassedSteps            []stepResultDocument `json:"passed_steps,omitempty"`
}

func resolvedDocumentJSON(doc ResolvedDocument) resolvedPlanDocument {
	mode := doc.Mode
	if mode == "" {
		mode = "optimization"
	}
	return resolvedPlanDocument{
		ManifestPath:     doc.ManifestPath,
		RoundName:        doc.RoundName,
		Mode:             mode,
		ParentRound:      doc.ParentRound,
		Target:           doc.Target,
		Agent:            doc.Agent,
		IncludedEvidence: append([]string(nil), doc.IncludedEvidence...),
		DeniedEvidence:   append([]string(nil), doc.DeniedEvidence...),
		InputPolicyPath:  filepath.ToSlash(doc.InputPolicyPath),
		OutputBundlePath: filepath.ToSlash(doc.OutputBundlePath),
	}
}

func nextChallengerRecordDocument(result pureoptimizer.NextChallengerRecord) resultDocument {
	out := resultDocument{
		Success:  result.Success,
		Proposal: result.Proposal,
		Failure:  result.Failure,
		Phases:   append([]pureoptimizer.Phase(nil), result.Phases...),
		Attempts: make([]attemptDocument, 0, len(result.Attempts)),
	}
	for _, attempt := range result.Attempts {
		doc := attemptDocument{
			Number:        attempt.Number,
			State:         attempt.State,
			RawOutput:     attempt.RawOutput,
			Proposal:      attempt.Proposal,
			Failure:       attempt.Failure,
			RetryFeedback: attempt.RetryFeedback,
		}
		for _, step := range attempt.PipelineResults {
			doc.PipelineResults = append(doc.PipelineResults, serializeStepResult(step))
		}
		if attempt.PipelineClassification != nil {
			doc.PipelineClassification = serializeClassification(*attempt.PipelineClassification)
		}
		out.Attempts = append(out.Attempts, doc)
	}
	return out
}

func writeAttemptJSON(stageDir string, name string, attempt pureoptimizer.Attempt, files *[]bundleFile) error {
	data, err := marshalDeterministic(attemptDocument{
		Number:        attempt.Number,
		State:         attempt.State,
		RawOutput:     attempt.RawOutput,
		Proposal:      attempt.Proposal,
		Failure:       attempt.Failure,
		RetryFeedback: attempt.RetryFeedback,
		PipelineResults: func() []stepResultDocument {
			out := make([]stepResultDocument, 0, len(attempt.PipelineResults))
			for _, step := range attempt.PipelineResults {
				out = append(out, serializeStepResult(step))
			}
			return out
		}(),
		PipelineClassification: func() *classificationDocument {
			if attempt.PipelineClassification == nil {
				return nil
			}
			return serializeClassification(*attempt.PipelineClassification)
		}(),
	})
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stageDir, name), data, 0o644); err != nil {
		return err
	}
	*files = append(*files, bundleFile{
		Kind:      "attempt_result",
		Path:      filepath.ToSlash(name),
		MediaType: "application/json",
		SHA256:    sha256String(data),
	})
	return nil
}

func bundleStatus(result pureoptimizer.NextChallengerRecord) string {
	if result.Success {
		return "complete"
	}
	return "failed"
}

func serializeStepResult(step pipeline.StepResult) stepResultDocument {
	doc := stepResultDocument{
		Name:     step.Name,
		Command:  append([]string(nil), step.Command...),
		CWD:      filepath.ToSlash(step.CWD),
		Passed:   step.Passed,
		ExitCode: step.ExitCode,
		Stdout:   step.Stdout,
		Stderr:   step.Stderr,
		Skipped:  step.Skipped,
		TimedOut: step.TimedOut,
	}
	if step.InfrastructureError != nil {
		doc.InfrastructureError = step.InfrastructureError.Error()
	}
	return doc
}

func serializeClassification(classification pipeline.Classification) *classificationDocument {
	doc := &classificationDocument{}
	for _, step := range classification.GenerationFailures {
		doc.GenerationFailures = append(doc.GenerationFailures, serializeStepResult(step))
	}
	for _, step := range classification.FormatErrors {
		doc.FormatErrors = append(doc.FormatErrors, serializeStepResult(step))
	}
	for _, step := range classification.TypeErrors {
		doc.TypeErrors = append(doc.TypeErrors, serializeStepResult(step))
	}
	for _, step := range classification.LintErrors {
		doc.LintErrors = append(doc.LintErrors, serializeStepResult(step))
	}
	for _, step := range classification.TestFailures {
		doc.TestFailures = append(doc.TestFailures, serializeStepResult(step))
	}
	for _, step := range classification.InfrastructureFailures {
		doc.InfrastructureFailures = append(doc.InfrastructureFailures, serializeStepResult(step))
	}
	for _, step := range classification.PassedSteps {
		doc.PassedSteps = append(doc.PassedSteps, serializeStepResult(step))
	}
	return doc
}

func marshalDeterministic(value any) ([]byte, error) {
	return json.MarshalIndent(value, "", "  ")
}

func sha256String(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
