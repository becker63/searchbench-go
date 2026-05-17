package lca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

const hfRowsAPI = "https://datasets-server.huggingface.co/rows"

// ExportOptions configures Hugging Face LCA JSONL materialization (#79).
type ExportOptions struct {
	ManifestDir string
	Config      string
	Split       string
	MaxItems    int
	Skip        int
	HTTPClient  *http.Client
	// RowsAPIBase overrides the HF datasets server base URL (tests only).
	RowsAPIBase string
}

// ExportFromHuggingFace streams rows from the public HF datasets server into manifest-local JSONL.
func ExportFromHuggingFace(ctx context.Context, opts ExportOptions) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if opts.MaxItems <= 0 {
		return "", fmt.Errorf("lca export: maxItems must be positive")
	}
	config := strings.TrimSpace(opts.Config)
	split := strings.TrimSpace(opts.Split)
	if config == "" || split == "" {
		return "", fmt.Errorf("lca export: config and split are required")
	}
	manifestDir := strings.TrimSpace(opts.ManifestDir)
	if manifestDir == "" {
		return "", fmt.Errorf("lca export: manifestDir is required")
	}

	outPath := filepath.Join(
		manifestDir,
		"datasets",
		slugDatasetName(JetBrainsLCADatasetName),
		config,
		split+".jsonl",
	)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", err
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}

	var rows []map[string]any
	offset := opts.Skip
	remaining := opts.MaxItems
	for remaining > 0 {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		length := remaining
		if length > 100 {
			length = 100
		}
		batch, err := fetchHFRows(ctx, client, opts.RowsAPIBase, config, split, offset, length)
		if err != nil {
			return "", err
		}
		if len(batch) == 0 {
			break
		}
		for _, row := range batch {
			normalized, err := normalizeHFRow(row)
			if err != nil {
				return "", err
			}
			rows = append(rows, normalized)
			remaining--
			if remaining == 0 {
				break
			}
		}
		offset += len(batch)
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("lca export: no rows returned from Hugging Face (config=%s split=%s skip=%d)", config, split, opts.Skip)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	enc := json.NewEncoder(f)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			_ = f.Close()
			return "", err
		}
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return outPath, nil
}

func fetchHFRows(ctx context.Context, client *http.Client, baseURL, config, split string, offset, length int) ([]map[string]any, error) {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = hfRowsAPI
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("dataset", JetBrainsLCADatasetName)
	q.Set("config", config)
	q.Set("split", split)
	q.Set("offset", strconv.Itoa(offset))
	q.Set("length", strconv.Itoa(length))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lca export: hf request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("lca export: hf status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload struct {
		Rows []struct {
			Row map[string]any `json:"row"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("lca export: decode hf response: %w", err)
	}
	out := make([]map[string]any, 0, len(payload.Rows))
	for _, item := range payload.Rows {
		out = append(out, item.Row)
	}
	return out, nil
}

func normalizeHFRow(row map[string]any) (map[string]any, error) {
	changed, err := normalizeChangedFiles(row["changed_files"])
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"repo_owner":    stringField(row, "repo_owner"),
		"repo_name":     stringField(row, "repo_name"),
		"base_sha":      stringField(row, "base_sha"),
		"issue_title":   stringField(row, "issue_title"),
		"issue_body":    stringField(row, "issue_body"),
		"changed_files": changed,
	}
	for _, key := range []string{"issue_url", "pull_url", "diff_url", "diff", "head_sha", "repo_language", "repo_license"} {
		if v := stringField(row, key); v != "" {
			out[key] = v
		}
	}
	if stars, ok := row["repo_stars"]; ok && stars != nil {
		switch v := stars.(type) {
		case float64:
			out["repo_stars"] = int(v)
		case int:
			out["repo_stars"] = v
		}
	}
	// Validate normalized row decodes as LCAHFRow.
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	var check domain.LCAHFRow
	if err := json.Unmarshal(raw, &check); err != nil {
		return nil, fmt.Errorf("lca export: normalized row: %w", err)
	}
	if check.RepoOwner == "" || check.RepoName == "" || check.BaseSHA == "" {
		return nil, fmt.Errorf("lca export: row missing repo identity")
	}
	return out, nil
}

func stringField(row map[string]any, key string) string {
	v, ok := row[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func normalizeChangedFiles(value any) ([]string, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, strings.TrimSpace(fmt.Sprint(item)))
		}
		return out, nil
	case []string:
		return v, nil
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			return nil, nil
		}
		var parsed []string
		if err := json.Unmarshal([]byte(text), &parsed); err == nil {
			return parsed, nil
		}
		return nil, fmt.Errorf("changed_files string did not decode: %q", v)
	default:
		return nil, fmt.Errorf("unsupported changed_files type %T", value)
	}
}
