package round

import "encoding/json"

func (p Plan) MarshalJSON() ([]byte, error) {
	type bundleView struct {
		ManifestPath              string                          `json:"manifest_path,omitempty"`
		RoundName                 string                          `json:"round_name,omitempty"`
		Mode                      string                          `json:"mode,omitempty"`
		Game                      GameConfig                      `json:"game,omitempty"`
		Round                     RoundConfig                     `json:"round,omitempty"`
		Dataset                   DatasetConfig                   `json:"dataset,omitempty"`
		Policies                  any                             `json:"policies"`
		Matches                   any                             `json:"matches"`
		Parallelism               any                             `json:"parallelism,omitempty"`
		CandidateInterfaceID      string                          `json:"candidate_interface_id,omitempty"`
		Evaluator                 EvaluatorConfig                 `json:"evaluator,omitempty"`
		Optimizer                 *OptimizerConfig                `json:"optimizer,omitempty"`
		Scoring                   ScoringConfig                   `json:"scoring,omitempty"`
		Output                    OutputConfig                    `json:"output,omitempty"`
		Report                    ReportConfig                    `json:"report_options,omitempty"`
		Bundle                    BundleConfig                    `json:"bundle,omitempty"`
		Lineage                   LineageConfig                   `json:"lineage,omitempty"`
		ChallengerMaterialization ChallengerMaterializationConfig `json:"challenger_materialization,omitempty"`
		ReportID                  string                          `json:"report_id,omitempty"`
		CreatedAt                 string                          `json:"created_at,omitempty"`
	}

	return json.Marshal(bundleView{
		ManifestPath:              p.ManifestPath,
		RoundName:                 p.RoundName,
		Mode:                      p.Mode,
		Game:                      p.Game,
		Round:                     p.Round,
		Dataset:                   p.Dataset,
		Policies:                  p.BundleSystems(),
		Matches:                   p.Matches,
		Parallelism:               p.Parallelism,
		CandidateInterfaceID:      p.CandidateInterfaceID,
		Evaluator:                 p.Evaluator,
		Optimizer:                 p.Optimizer,
		Scoring:                   p.Scoring,
		Output:                    p.Output,
		Report:                    p.Report,
		Bundle:                    p.Bundle,
		Lineage:                   p.Lineage,
		ChallengerMaterialization: p.ChallengerMaterialization,
		ReportID:                  p.ReportID.String(),
		CreatedAt:                 p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}
