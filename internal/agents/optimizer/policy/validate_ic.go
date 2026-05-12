package policy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	execpipeline "github.com/becker63/searchbench-go/internal/adapters/pipeline/exec"
	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

func validateIterativeContextProposal(ctx context.Context, proposal pureoptimizer.NextChallengerProposal) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	if strings.TrimSpace(proposal.Code) == "" {
		return pipelineFailureResult(stagePolicyFailure("policy code is empty"))
	}

	if strings.Contains(proposal.Code, "```") {
		return pipelineFailureResult(stagePolicyFailure("policy code contains markdown code fences; emit raw Python only"))
	}

	icRoot, err := ResolveIterativeContextRoot()
	if err != nil {
		return infraFailure(proposal, fmt.Errorf("resolve iterative-context root: %w", err))
	}

	uvBin, err := execpipeline.LookupUvBinary()
	if err != nil {
		return infraFailure(proposal, err)
	}

	workspace, err := os.MkdirTemp("", "searchbench-policy-candidate-*")
	if err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "create staged policy workspace",
			Cause:   err,
		}
	}
	defer func() { _ = os.RemoveAll(workspace) }()

	validationDir := filepath.Join(workspace, "validation")
	if err := os.MkdirAll(validationDir, 0o755); err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "create validation directory",
			Cause:   err,
		}
	}

	policyPath := filepath.Join(workspace, proposal.ArtifactName)
	if err := os.WriteFile(policyPath, []byte(proposal.Code), 0o644); err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "write staged policy proposal",
			Cause:   err,
		}
	}

	sum := sha256.Sum256([]byte(proposal.Code))
	meta := map[string]any{
		"artifact_id":   proposal.ArtifactID.String(),
		"artifact_name": proposal.ArtifactName,
		"interface_id":  proposal.InterfaceID,
		"symbol":        CanonicalICPolicySymbol,
		"sha256":        hex.EncodeToString(sum[:]),
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "encode staged policy metadata",
			Cause:   err,
		}
	}
	if err := os.WriteFile(filepath.Join(workspace, "policy.json"), metaBytes, 0o644); err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "write policy.json metadata",
			Cause:   err,
		}
	}

	stageStep := pipeline.StepResult{
		Name:    "stage_policy",
		Command: []string{},
		CWD:     workspace,
		Passed:  true,
	}
	if err := os.WriteFile(filepath.Join(validationDir, "stage-policy.json"), metaBytes, 0o644); err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "write stage-policy.json",
			Cause:   err,
		}
	}

	results := []pipeline.StepResult{stageStep}

	policyAbs, err := filepath.Abs(policyPath)
	if err != nil {
		return infraFailure(proposal, fmt.Errorf("abs policy path: %w", err))
	}
	wsAbs, err := filepath.Abs(workspace)
	if err != nil {
		return infraFailure(proposal, fmt.Errorf("abs workspace path: %w", err))
	}

	pythonBin, err := resolvePython()
	if err != nil {
		step := pipeline.StepResult{
			Name:                "policy_static_precheck",
			Command:             []string{"python3", "-m", "py_compile", proposal.ArtifactName},
			CWD:                 workspace,
			ExitCode:            -1,
			InfrastructureError: err,
		}
		results = append(results, step)
		classification := pipeline.Classify(results)
		return failureFromClassification(results, &classification, proposal)
	}

	preSpec := pipeline.CommandSpec{
		Name:    "policy_static_precheck",
		Command: []string{pythonBin, "-m", "py_compile", proposal.ArtifactName},
		CWD:     workspace,
		Timeout: 30 * time.Second,
	}
	preResult := execpipeline.ExecCommandRunner{}.Run(ctx, preSpec)
	preResult.Name = preSpec.Name
	preResult.Command = append([]string(nil), preSpec.Command...)
	preResult.CWD = preSpec.CWD
	if preResult.InfrastructureError == nil && preResult.ExitCode == 0 {
		preResult.Passed = true
	}
	results = append(results, preResult)
	if preResult.Failed() {
		classification := pipeline.Classify(results)
		return failureFromClassification(results, &classification, proposal)
	}

	policyID := proposal.ArtifactID.String()
	icArgv := []string{
		uvBin, "run", "python", "-m", "iterative_context.validate_policy",
		"--policy-path", policyAbs,
		"--policy-id", policyID,
		"--symbol", CanonicalICPolicySymbol,
		"--json",
	}

	allow := execpipeline.ICOptimizerAllowlist{
		UvBinary:         uvBin,
		WorkspaceRootAbs: wsAbs,
		PolicyFileAbs:    policyAbs,
		PolicyID:         policyID,
		Symbol:           CanonicalICPolicySymbol,
	}

	execSteps := []pipeline.CommandSpec{
		{
			Name:    "ic_validate_policy",
			Command: icArgv,
			CWD:     icRoot,
			Timeout: 3 * time.Minute,
		},
		{
			Name:    "basedpyright",
			Command: []string{uvBin, "run", "basedpyright"},
			CWD:     icRoot,
			Timeout: 6 * time.Minute,
		},
		{
			Name:    "ruff_check",
			Command: []string{uvBin, "run", "ruff", "check"},
			CWD:     icRoot,
			Timeout: 3 * time.Minute,
		},
		{
			Name:    "pytest",
			Command: []string{uvBin, "run", "pytest"},
			CWD:     icRoot,
			Timeout: 12 * time.Minute,
		},
	}

	runner := execpipeline.Runner{
		CommandRunner: execpipeline.ExecCommandRunner{},
		Allowlist:     allow,
	}
	execResults := runner.Run(ctx, execSteps)
	results = append(results, execResults...)

	classification := pipeline.Classify(results)
	_ = writePipelineArtifacts(validationDir, results, classification)

	if classification.HasFailures() {
		return failureFromClassification(results, &classification, proposal)
	}

	return pureoptimizer.ProposalValidationResult{
		Results:        results,
		Classification: &classification,
	}, nil
}

func writePipelineArtifacts(dir string, results []pipeline.StepResult, classification pipeline.Classification) error {
	simplified := make([]map[string]any, 0, len(results))
	for _, r := range results {
		row := map[string]any{
			"name":        r.Name,
			"passed":      r.Passed,
			"exit_code":   r.ExitCode,
			"timed_out":   r.TimedOut,
			"skipped":     r.Skipped,
			"duration_ms": r.Duration.Milliseconds(),
		}
		if r.InfrastructureError != nil {
			row["infrastructure_error"] = r.InfrastructureError.Error()
		}
		simplified = append(simplified, row)
	}
	payload := map[string]any{
		"steps":        simplified,
		"has_failures": classification.HasFailures(),
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "pipeline-result.json"), data, 0o644)
}

func pipelineFailureResult(step pipeline.StepResult) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	results := []pipeline.StepResult{step}
	classification := pipeline.Classify(results)
	return failureFromClassification(results, &classification, pureoptimizer.NextChallengerProposal{})
}

func stagePolicyFailure(message string) pipeline.StepResult {
	return pipeline.StepResult{
		Name:     "stage_policy",
		Command:  []string{},
		ExitCode: 1,
		Stderr:   message,
	}
}

func infraFailure(proposal pureoptimizer.NextChallengerProposal, cause error) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	step := pipeline.StepResult{
		Name:                "policy_workspace",
		Command:             []string{},
		ExitCode:            -1,
		InfrastructureError: cause,
	}
	results := []pipeline.StepResult{step}
	classification := pipeline.Classify(results)
	res, fail := failureFromClassification(results, &classification, proposal)
	fail.Retryable = false
	fail.Kind = pureoptimizer.FailureKindPolicyPipelineInfrastructure
	return res, fail
}

func failureFromClassification(results []pipeline.StepResult, classification *pipeline.Classification, proposal pureoptimizer.NextChallengerProposal) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	kind := pureoptimizer.FailureKindPolicyPipelineFailed
	retryable := true
	category := "validation"
	if len(classification.InfrastructureFailures) > 0 {
		kind = pureoptimizer.FailureKindPolicyPipelineInfrastructure
		retryable = false
		category = "infrastructure"
	}
	msg := "policy validation pipeline failed"
	if proposal.ArtifactID.String() != "" {
		msg = fmt.Sprintf("policy validation pipeline failed for %s", proposal.ArtifactID.String())
	}
	return pureoptimizer.ProposalValidationResult{
			Results:        results,
			Classification: classification,
		}, &pureoptimizer.Failure{
			Phase:            pureoptimizer.PhaseRunPolicyPipeline,
			Kind:             kind,
			Message:          msg,
			Retryable:        retryable,
			PipelineCategory: category,
			PipelineFeedback: pipeline.FormatPipelineFeedback(*classification, 2400),
			Cause:            errors.New(msg),
		}
}
