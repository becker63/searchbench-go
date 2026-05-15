package domain

// Prediction is the normalized answer produced by a system run.
//
// For now Searchbench is file-localization first. Function/symbol prediction
// can be added later without putting raw model output into this type.
type Prediction struct {
	Files     []RepoRelPath `json:"files"`
	Reasoning string        `json:"reasoning,omitempty"`
}
