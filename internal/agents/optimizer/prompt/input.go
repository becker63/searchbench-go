package prompt

import (
	"encoding/json"
	"sort"

	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// OutputSchemaJSON is the strict JSON contract for optimizer proposals.
const OutputSchemaJSON = `{"type":"object","additionalProperties":false,"properties":{"artifact_id":{"type":"string"},"artifact_name":{"type":"string"},"interface_id":{"type":"string"},"code":{"type":"string"},"summary":{"type":"string"},"risk_notes":{"type":"array","items":{"type":"string"}}},"required":["artifact_id","artifact_name","interface_id","code"]}`

// DefaultConstraints are the stable optimizer prompt constraints.
var DefaultConstraints = []string{
	"Use only the provided parent-run evidence and current policy source.",
	"Do not include gold labels, oracle files, or raw dataset answers.",
	"Return strict JSON only. Do not wrap the final answer in markdown.",
	"Emit exactly one replacement policy proposal.",
}

// Input is the typed optimizer prompt contract.
type Input struct {
	SystemPrompt        string
	TargetArtifactID    string
	TargetArtifactName  string
	TargetInterfaceID   string
	ParentBundleID      string
	IncludedEvidence    []string
	DeniedEvidence      []string
	InputPolicyID       string
	InputPolicySource   string
	ReportSummaryJSON   string
	RoundEvidenceJSON   string
	ObjectiveResultJSON string
	Constraints         []string
	RetryFeedback       []string
	OutputSchemaJSON    string
}

// InputFromSpec projects one optimizer spec into its prompt-safe input.
func InputFromSpec(spec pureoptimizer.Spec) (Input, error) {
	include := append([]string(nil), spec.Evidence.IncludedKinds...)
	deny := append([]string(nil), spec.Evidence.DeniedKinds...)
	sort.Strings(include)
	sort.Strings(deny)

	input := Input{
		SystemPrompt:       spec.Agent.SystemPrompt,
		TargetArtifactID:   spec.Target.OutputArtifactID.String(),
		TargetArtifactName: spec.Target.OutputName,
		TargetInterfaceID:  spec.Target.InterfaceID,
		ParentBundleID:     spec.Evidence.ParentRound.BundleID,
		IncludedEvidence:   include,
		DeniedEvidence:     deny,
		InputPolicyID:      spec.Evidence.InputPolicy.ArtifactID.String(),
		InputPolicySource:  spec.Evidence.InputPolicy.Source,
		Constraints:        append([]string(nil), DefaultConstraints...),
		OutputSchemaJSON:   OutputSchemaJSON,
	}

	if spec.Evidence.ReportSummary != nil {
		data, err := marshalPretty(spec.Evidence.ReportSummary)
		if err != nil {
			return Input{}, err
		}
		input.ReportSummaryJSON = data
	}
	if spec.Evidence.RoundEvidence != nil {
		data, err := marshalPretty(spec.Evidence.RoundEvidence)
		if err != nil {
			return Input{}, err
		}
		input.RoundEvidenceJSON = data
	}
	if spec.Evidence.ObjectiveResult != nil {
		data, err := marshalPretty(spec.Evidence.ObjectiveResult)
		if err != nil {
			return Input{}, err
		}
		input.ObjectiveResultJSON = data
	}

	return input, nil
}

func marshalPretty(value any) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
