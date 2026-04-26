package domain

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
)

// SchemaError is a typed validation/parsing error for dataset and localization
// domain shapes.
type SchemaError struct {
	Field   string
	Message string
}

func (e *SchemaError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// LCATaskIdentity is the deterministic identity for one LCA dataset task.
type LCATaskIdentity struct {
	DatasetName   string `json:"dataset_name"`
	DatasetConfig string `json:"dataset_config"`
	DatasetSplit  string `json:"dataset_split"`
	RepoOwner     string `json:"repo_owner"`
	RepoName      string `json:"repo_name"`
	BaseSHA       string `json:"base_sha"`
	IssueURL      string `json:"issue_url,omitempty"`
	PullURL       string `json:"pull_url,omitempty"`
}

// Validate rejects incomplete task identities.
func (i LCATaskIdentity) Validate() error {
	switch {
	case strings.TrimSpace(i.DatasetName) == "":
		return &SchemaError{Field: "dataset_name", Message: "value is required"}
	case strings.TrimSpace(i.DatasetConfig) == "":
		return &SchemaError{Field: "dataset_config", Message: "value is required"}
	case strings.TrimSpace(i.DatasetSplit) == "":
		return &SchemaError{Field: "dataset_split", Message: "value is required"}
	case strings.TrimSpace(i.RepoOwner) == "":
		return &SchemaError{Field: "repo_owner", Message: "value is required"}
	case strings.TrimSpace(i.RepoName) == "":
		return &SchemaError{Field: "repo_name", Message: "value is required"}
	case strings.TrimSpace(i.BaseSHA) == "":
		return &SchemaError{Field: "base_sha", Message: "value is required"}
	default:
		return nil
	}
}

// RepoFullName returns the normalized repository owner/name pair.
func (i LCATaskIdentity) RepoFullName() string {
	return normalizeIdentityToken(i.RepoOwner) + "/" + normalizeIdentityToken(i.RepoName)
}

// IssueOrPullKey returns the deterministic issue key used in the task ID.
func (i LCATaskIdentity) IssueOrPullKey() string {
	if value := strings.TrimSpace(i.IssueURL); value != "" {
		return value
	}
	if value := strings.TrimSpace(i.PullURL); value != "" {
		return value
	}
	return "unknown-issue"
}

// TaskID returns the stable benchmark task identifier.
func (i LCATaskIdentity) TaskID() TaskID {
	return TaskID(fmt.Sprintf(
		"%s:%s:%s:%s@%s:%s",
		strings.TrimSpace(i.DatasetName),
		strings.TrimSpace(i.DatasetConfig),
		strings.TrimSpace(i.DatasetSplit),
		i.RepoFullName(),
		normalizeIdentityToken(i.BaseSHA),
		i.IssueOrPullKey(),
	))
}

// LCAContext is the model-visible issue and repo metadata for one LCA task.
type LCAContext struct {
	IssueTitle    string   `json:"issue_title"`
	IssueBody     string   `json:"issue_body"`
	DiffURL       string   `json:"diff_url,omitempty"`
	Diff          string   `json:"diff,omitempty"`
	HeadSHA       string   `json:"head_sha,omitempty"`
	RepoLanguage  string   `json:"repo_language,omitempty"`
	RepoLanguages []string `json:"repo_languages,omitempty"`
	RepoLicense   string   `json:"repo_license,omitempty"`
	RepoStars     int      `json:"repo_stars,omitempty"`
}

// Validate rejects incomplete agent-visible context.
func (c LCAContext) Validate() error {
	switch {
	case strings.TrimSpace(c.IssueTitle) == "":
		return &SchemaError{Field: "issue_title", Message: "value is required"}
	case strings.TrimSpace(c.IssueBody) == "":
		return &SchemaError{Field: "issue_body", Message: "value is required"}
	default:
		return nil
	}
}

// LCAGold captures the canonical changed_files gold label and small derived
// metadata.
type LCAGold struct {
	ChangedFiles      NonEmpty[RepoRelPath] `json:"changed_files"`
	ChangedFilesCount int                   `json:"changed_files_count,omitempty"`
	ChangedFilesExts  []string              `json:"changed_files_exts,omitempty"`
}

// NewLCAGold constructs normalized LCA gold labels.
func NewLCAGold(paths []string) (LCAGold, error) {
	normalized := CanonicalizePaths(paths)
	changedFiles, err := NonEmptyFromSlice(normalized)
	if err != nil {
		return LCAGold{}, &SchemaError{
			Field:   "changed_files",
			Message: "at least one changed file is required",
		}
	}

	return LCAGold{
		ChangedFiles:      changedFiles,
		ChangedFilesCount: len(normalized),
		ChangedFilesExts:  changedFileExts(normalized),
	}, nil
}

// LCATask is the canonical pure LCA task model.
type LCATask struct {
	Identity LCATaskIdentity `json:"identity"`
	Context  LCAContext      `json:"context"`
	Gold     LCAGold         `json:"gold"`
	Repo     *RepoSnapshot   `json:"repo,omitempty"`
}

// Validate rejects incomplete localization tasks.
func (t LCATask) Validate() error {
	if err := t.Identity.Validate(); err != nil {
		return err
	}
	if err := t.Context.Validate(); err != nil {
		return err
	}
	if err := t.Gold.ChangedFiles.Validate(); err != nil {
		return &SchemaError{Field: "changed_files", Message: "at least one changed file is required"}
	}
	return nil
}

// TaskID delegates to the identity-based deterministic task ID.
func (t LCATask) TaskID() TaskID {
	return t.Identity.TaskID()
}

// ChangedFiles returns the normalized canonical gold changed files.
func (t LCATask) ChangedFiles() []RepoRelPath {
	return t.Gold.ChangedFiles.All()
}

// WithRepo returns a copy annotated with a concrete repository path.
func (t LCATask) WithRepo(path HostPath) LCATask {
	task := t
	task.Repo = &RepoSnapshot{
		Name: RepoName(t.Identity.RepoFullName()),
		SHA:  RepoSHA(normalizeIdentityToken(t.Identity.BaseSHA)),
		Path: path,
	}
	return task
}

// TaskSpec projects the LCA task into the existing generic Searchbench task
// shape used by compare/report code.
func (t LCATask) TaskSpec() TaskSpec {
	repo := RepoSnapshot{
		Name: RepoName(t.Identity.RepoFullName()),
		SHA:  RepoSHA(normalizeIdentityToken(t.Identity.BaseSHA)),
	}
	if t.Repo != nil {
		repo = *t.Repo
	}

	return TaskSpec{
		ID:        t.TaskID(),
		Benchmark: BenchmarkLCA,
		Repo:      repo,
		Input: TaskInput{
			Title: strings.TrimSpace(t.Context.IssueTitle),
			Body:  strings.TrimSpace(t.Context.IssueBody),
		},
		Oracle: TaskOracle{
			GoldFiles: t.ChangedFiles(),
		},
	}
}

// LCAHFRow is the typed transport-edge shape for one Hugging Face dataset row.
type LCAHFRow struct {
	RepoOwner     string   `json:"repo_owner"`
	RepoName      string   `json:"repo_name"`
	BaseSHA       string   `json:"base_sha"`
	IssueTitle    string   `json:"issue_title"`
	IssueBody     string   `json:"issue_body"`
	ChangedFiles  []string `json:"changed_files"`
	IssueURL      string   `json:"issue_url,omitempty"`
	PullURL       string   `json:"pull_url,omitempty"`
	DiffURL       string   `json:"diff_url,omitempty"`
	Diff          string   `json:"diff,omitempty"`
	HeadSHA       string   `json:"head_sha,omitempty"`
	RepoLanguage  string   `json:"repo_language,omitempty"`
	RepoLanguages []string `json:"repo_languages,omitempty"`
	RepoLicense   string   `json:"repo_license,omitempty"`
	RepoStars     int      `json:"repo_stars,omitempty"`
}

// UnmarshalJSON preserves the Python loader behavior where changed_files may be
// either a JSON array or a stringified Python/JSON list.
func (r *LCAHFRow) UnmarshalJSON(data []byte) error {
	type wire struct {
		RepoOwner     string          `json:"repo_owner"`
		RepoName      string          `json:"repo_name"`
		BaseSHA       string          `json:"base_sha"`
		IssueTitle    string          `json:"issue_title"`
		IssueBody     string          `json:"issue_body"`
		ChangedFiles  json.RawMessage `json:"changed_files"`
		IssueURL      string          `json:"issue_url,omitempty"`
		PullURL       string          `json:"pull_url,omitempty"`
		DiffURL       string          `json:"diff_url,omitempty"`
		Diff          string          `json:"diff,omitempty"`
		HeadSHA       string          `json:"head_sha,omitempty"`
		RepoLanguage  string          `json:"repo_language,omitempty"`
		RepoLanguages []string        `json:"repo_languages,omitempty"`
		RepoLicense   string          `json:"repo_license,omitempty"`
		RepoStars     int             `json:"repo_stars,omitempty"`
	}

	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	changedFiles, err := parseChangedFilesField(value.ChangedFiles)
	if err != nil {
		return err
	}

	*r = LCAHFRow{
		RepoOwner:     value.RepoOwner,
		RepoName:      value.RepoName,
		BaseSHA:       value.BaseSHA,
		IssueTitle:    value.IssueTitle,
		IssueBody:     value.IssueBody,
		ChangedFiles:  changedFiles,
		IssueURL:      value.IssueURL,
		PullURL:       value.PullURL,
		DiffURL:       value.DiffURL,
		Diff:          value.Diff,
		HeadSHA:       value.HeadSHA,
		RepoLanguage:  value.RepoLanguage,
		RepoLanguages: append([]string(nil), value.RepoLanguages...),
		RepoLicense:   value.RepoLicense,
		RepoStars:     value.RepoStars,
	}
	return nil
}

// ToTask converts one typed dataset row into the canonical LCA task model.
func (r LCAHFRow) ToTask(datasetName string, datasetConfig string, datasetSplit string) (LCATask, error) {
	identity := LCATaskIdentity{
		DatasetName:   datasetName,
		DatasetConfig: datasetConfig,
		DatasetSplit:  datasetSplit,
		RepoOwner:     r.RepoOwner,
		RepoName:      r.RepoName,
		BaseSHA:       r.BaseSHA,
		IssueURL:      r.IssueURL,
		PullURL:       r.PullURL,
	}
	if err := identity.Validate(); err != nil {
		return LCATask{}, err
	}

	context := LCAContext{
		IssueTitle:    r.IssueTitle,
		IssueBody:     r.IssueBody,
		DiffURL:       r.DiffURL,
		Diff:          r.Diff,
		HeadSHA:       r.HeadSHA,
		RepoLanguage:  r.RepoLanguage,
		RepoLanguages: append([]string(nil), r.RepoLanguages...),
		RepoLicense:   r.RepoLicense,
		RepoStars:     r.RepoStars,
	}
	if err := context.Validate(); err != nil {
		return LCATask{}, err
	}

	gold, err := NewLCAGold(r.ChangedFiles)
	if err != nil {
		return LCATask{}, err
	}

	task := LCATask{
		Identity: identity,
		Context:  context,
		Gold:     gold,
	}
	if err := task.Validate(); err != nil {
		return LCATask{}, err
	}
	return task, nil
}

// CanonicalizePath normalizes one repo-relative path into a stable comparison
// key.
func CanonicalizePath(value string) RepoRelPath {
	canonical := strings.TrimSpace(value)
	if canonical == "" {
		return ""
	}

	canonical = strings.ReplaceAll(canonical, "\\", "/")
	canonical = strings.TrimPrefix(canonical, "./")
	canonical = strings.TrimLeft(canonical, "/")
	canonical = path.Clean(canonical)
	for canonical == "." || strings.HasPrefix(canonical, "../") {
		canonical = strings.TrimPrefix(canonical, "./")
		canonical = strings.TrimPrefix(canonical, "../")
		canonical = path.Clean(canonical)
		if canonical == "." {
			return ""
		}
	}
	canonical = strings.TrimLeft(canonical, "/")
	if canonical == "." || canonical == "" {
		return ""
	}
	return RepoRelPath(strings.ToLower(canonical))
}

// CanonicalizePaths normalizes, deduplicates, and sorts repo-relative paths.
func CanonicalizePaths(values []string) []RepoRelPath {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[RepoRelPath]struct{}, len(values))
	normalized := make([]RepoRelPath, 0, len(values))
	for _, value := range values {
		path := CanonicalizePath(value)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		normalized = append(normalized, path)
	}

	sort.Slice(normalized, func(i int, j int) bool {
		return normalized[i] < normalized[j]
	})
	return normalized
}

func normalizeIdentityToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func changedFileExts(paths []RepoRelPath) []string {
	if len(paths) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(paths))
	exts := make([]string, 0, len(paths))
	for _, file := range paths {
		ext := strings.ToLower(path.Ext(string(file)))
		if ext == "" {
			continue
		}
		if _, ok := seen[ext]; ok {
			continue
		}
		seen[ext] = struct{}{}
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	return exts
}

func parseChangedFilesField(data json.RawMessage) ([]string, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, &SchemaError{Field: "changed_files", Message: "value is required"}
	}

	var asSlice []string
	if err := json.Unmarshal(data, &asSlice); err == nil {
		return append([]string(nil), asSlice...), nil
	}

	var asString string
	if err := json.Unmarshal(data, &asString); err != nil {
		return nil, &SchemaError{
			Field:   "changed_files",
			Message: "must be a string array or a stringified list",
		}
	}

	trimmed := strings.TrimSpace(asString)
	if trimmed == "" {
		return nil, &SchemaError{Field: "changed_files", Message: "value is required"}
	}

	jsonish := strings.ReplaceAll(trimmed, "'", "\"")
	var fromStringified []string
	if err := json.Unmarshal([]byte(jsonish), &fromStringified); err != nil {
		return nil, &SchemaError{
			Field:   "changed_files",
			Message: "stringified list could not be parsed",
		}
	}
	return fromStringified, nil
}
