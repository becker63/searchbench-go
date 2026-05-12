package lca

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/ports/dataset"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestSourceMatchesSortsAndWindows(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "datasets", "JetBrains-Research_lca-bug-localization", "py", "dev.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := "{\"repo_owner\":\"z\",\"repo_name\":\"late\",\"base_sha\":\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\",\"issue_title\":\"t\",\"issue_body\":\"b\",\"issue_url\":\"https://example.test/issues/02\",\"changed_files\":[\"b.go\"]}\n" +
		"{\"repo_owner\":\"a\",\"repo_name\":\"early\",\"base_sha\":\"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\",\"issue_title\":\"t\",\"issue_body\":\"b\",\"issue_url\":\"https://example.test/issues/01\",\"changed_files\":[\"a.go\"]}\n"
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}

	max := 1
	out, err := NewSource().Matches(context.Background(), dataset.Request{
		ManifestDir: root,
		Kind:        "lca",
		Name:        JetBrainsLCADatasetName,
		Config:      "py",
		Split:       "dev",
		MaxItems:    &max,
	})
	if err != nil {
		t.Fatalf("Matches: %v", err)
	}
	if out.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", out.Len())
	}
	first := out.Head()
	want := domain.MatchID("JetBrains-Research/lca-bug-localization:py:dev:a/early@bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb:https://example.test/issues/01")
	if first.ID != want {
		t.Fatalf("MatchID = %q, want %q", first.ID, want)
	}
}

func TestIsJetBrainsDatasetNameCaseInsensitive(t *testing.T) {
	t.Parallel()

	req := dataset.Request{Kind: "lca", Name: "  JETBRAINS-RESEARCH/LCA-BUG-LOCALIZATION "}
	if !IsJetBrainsDataset(req) {
		t.Fatal("expected JetBrains match")
	}
}
