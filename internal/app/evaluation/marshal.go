package evaluation

import "encoding/json"

func (p Plan) MarshalJSON() ([]byte, error) {
	type bundleView struct {
		ManifestPath   string          `json:"manifest_path,omitempty"`
		ExperimentName string          `json:"experiment_name,omitempty"`
		Mode           string          `json:"mode,omitempty"`
		Dataset        DatasetConfig   `json:"dataset,omitempty"`
		Systems        any             `json:"systems"`
		Tasks          any             `json:"tasks"`
		Parallelism    any             `json:"parallelism,omitempty"`
		Evaluator      EvaluatorConfig `json:"evaluator,omitempty"`
		Scoring        ScoringConfig   `json:"scoring,omitempty"`
		Output         OutputConfig    `json:"output,omitempty"`
		Report         ReportConfig    `json:"report_options,omitempty"`
		Bundle         BundleConfig    `json:"bundle,omitempty"`
		ReportID       string          `json:"report_id,omitempty"`
		CreatedAt      string          `json:"created_at,omitempty"`
	}

	return json.Marshal(bundleView{
		ManifestPath:   p.ManifestPath,
		ExperimentName: p.ExperimentName,
		Mode:           p.Mode,
		Dataset:        p.Dataset,
		Systems:        p.BundleSystems(),
		Tasks:          p.Tasks,
		Parallelism:    p.Parallelism,
		Evaluator:      p.Evaluator,
		Scoring:        p.Scoring,
		Output:         p.Output,
		Report:         p.Report,
		Bundle:         p.Bundle,
		ReportID:       p.ReportID.String(),
		CreatedAt:      p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}
