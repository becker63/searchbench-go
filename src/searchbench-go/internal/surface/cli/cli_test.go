package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/testing/reporoot"
)

func TestRunCommandExecutesManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	temp := t.TempDir()
	manifestPath := filepath.Join(reporoot.MonorepoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")

	var out bytes.Buffer
	err := RunWithWriters(
		context.Background(),
		[]string{"--quiet", "round", "run", "--manifest", manifestPath, "--bundle-root", temp, "--bundle-id", "cli-localrun"},
		&out,
		io.Discard,
	)
	if err != nil {
		t.Fatalf("RunWithWriters() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"bundle=",
		"report_id=",
		"objective=localization-v1",
		"final=final:",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\n%s", want, got)
		}
	}
	for _, name := range []string{"resolved-round.json", "round-report.json", "evidence.pkl", "decision.json", "objective.json", "metadata.json"} {
		if _, err := os.Stat(filepath.Join(temp, "games", "code-localization", "rounds", "cli-localrun", name)); err != nil {
			t.Fatalf("os.Stat(%q) error = %v", name, err)
		}
	}
}

func requirePkl(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pkl"); err != nil {
		t.Skip("pkl CLI not available on PATH")
	}
}
