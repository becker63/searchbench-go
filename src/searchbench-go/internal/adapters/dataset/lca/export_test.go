package lca

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestExportFromHuggingFaceWritesJSONL(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rows" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"rows": []map[string]any{
				{"row": map[string]any{
					"repo_owner":    "acme",
					"repo_name":     "demo",
					"base_sha":      "abc123",
					"issue_title":   "t",
					"issue_body":    "b",
					"changed_files": []string{"main.go"},
				}},
			},
		})
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	path, err := ExportFromHuggingFace(context.Background(), ExportOptions{
		ManifestDir: dir,
		Config:      "py",
		Split:       "dev",
		MaxItems:    1,
		Skip:        0,
		HTTPClient:  srv.Client(),
		RowsAPIBase: srv.URL + "/rows",
	})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat output: %v", err)
	}
	want := filepath.Join(dir, "datasets", "JetBrains-Research_lca-bug-localization", "py", "dev.jsonl")
	if path != want {
		t.Fatalf("path = %q want %q", path, want)
	}
}
