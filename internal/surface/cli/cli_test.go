package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/report"
)

func TestDemoReportCommandRuns(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	err := RunWithWriters(context.Background(), []string{"--quiet", "--no-color", "demo-report"}, &out, io.Discard)
	if err != nil {
		t.Fatalf("RunWithWriters() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Searchbench Candidate Report",
		"Decision",
		"PROMOTE",
		"Metrics",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\n%s", want, got)
		}
	}
	if strings.Contains(got, "def score") {
		t.Fatal("output leaked policy source")
	}
	if strings.Contains(got, "\"source\"") {
		t.Fatal("output leaked policy source field")
	}
}

func TestDemoReportCommandJSONOutput(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	err := RunWithWriters(context.Background(), []string{"--quiet", "demo-report", "--output", "json"}, &out, io.Discard)
	if err != nil {
		t.Fatalf("RunWithWriters() error = %v", err)
	}

	data := out.Bytes()
	if bytes.Contains(data, []byte("def score")) {
		t.Fatal("json output leaked policy source")
	}

	var got report.CandidateReport
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got.Decision.Decision != report.DecisionPromote {
		t.Fatalf("Decision = %q, want %q", got.Decision.Decision, report.DecisionPromote)
	}
}

func TestDemoReportCommandRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "zero tasks",
			args: []string{"--quiet", "demo-report", "--tasks", "0"},
		},
		{
			name: "zero workers",
			args: []string{"--quiet", "demo-report", "--max-workers", "0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := RunWithWriters(context.Background(), tt.args, &bytes.Buffer{}, io.Discard)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
