package usage

import (
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/testing/importcheck"
)

func TestCollectorReportedUsageCreatesReportedRecord(t *testing.T) {
	t.Parallel()

	collector, err := NewCollector(Config{})
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}

	callID := collector.StartCall(StartEvent{
		Phase: "run_evaluator",
		Node:  "model-node",
		Input: []string{"find retry interceptor"},
	})
	collector.EndCall(callID, EndEvent{
		Output: []string{"src/main.go"},
		Reported: ReportedUsage{
			InputTokens:  MaybeTokenCount{Value: 11, Set: true},
			OutputTokens: MaybeTokenCount{Value: 4, Set: true},
			TotalTokens:  MaybeTokenCount{Value: 15, Set: true},
		},
	})

	records := collector.Records()
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	record := records[0]
	if got, want := record.Source, SourceReported; got != want {
		t.Fatalf("record.Source = %q, want %q", got, want)
	}
	if got, want := record.TotalTokens, domain.TokenCount(15); got != want {
		t.Fatalf("record.TotalTokens = %d, want %d", got, want)
	}
}

func TestCollectorMissingProviderUsageFallsBackToEstimate(t *testing.T) {
	t.Parallel()

	collector, err := NewCollector(Config{})
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}

	callID := collector.StartCall(StartEvent{
		Phase: "run_evaluator",
		Node:  "model-node",
		Input: []string{"find retry interceptor"},
	})
	collector.EndCall(callID, EndEvent{
		Output: []string{"src/main.go"},
	})

	record := collector.Records()[0]
	if got, want := record.Source, SourceEstimated; got != want {
		t.Fatalf("record.Source = %q, want %q", got, want)
	}
	if record.InputTokens == 0 || record.OutputTokens == 0 || record.TotalTokens == 0 {
		t.Fatalf("record = %#v, want estimated counts", record)
	}
}

func TestCollectorTokenizerFailureRecordsIncompleteUsageWithoutZeroingAvailableData(t *testing.T) {
	t.Parallel()

	collector, err := NewCollector(Config{
		Tokenizer: failTokenizer{err: ErrTokenizerUnavailable},
	})
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}

	callID := collector.StartCall(StartEvent{
		Phase: "run_evaluator",
		Node:  "model-node",
		Input: []string{"find retry interceptor"},
	})
	collector.EndCall(callID, EndEvent{
		Reported: ReportedUsage{
			InputTokens: MaybeTokenCount{Value: 7, Set: true},
		},
	})

	record := collector.Records()[0]
	if got, want := record.Source, SourceReported; got != want {
		t.Fatalf("record.Source = %q, want %q", got, want)
	}
	if record.InputTokens == 0 {
		t.Fatalf("record.InputTokens = %d, want non-zero reported fallback", record.InputTokens)
	}
	if len(record.Issues) == 0 {
		t.Fatalf("record.Issues = %#v, want incomplete usage issues", record.Issues)
	}
}

func TestUsagePackageAvoidsForbiddenImports(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	dir := filepath.Dir(currentFile)
	fs := token.NewFileSet()
	files, err := importcheck.ParseNonTestGoFilesImportsOnly(fs, dir)
	if err != nil {
		t.Fatalf("ParseNonTestGoFilesImportsOnly() error = %v", err)
	}

	forbiddenSubstrings := []string{
		"cloudwego/eino",
		"langsmith",
		"langfuse",
		"tracing",
	}

	for _, file := range files {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			for _, forbidden := range forbiddenSubstrings {
				if strings.Contains(path, forbidden) {
					t.Fatalf("forbidden import %q contains %q", path, forbidden)
				}
			}
		}
	}
}

type failTokenizer struct {
	err error
}

func (t failTokenizer) CountStrings([]string) (domain.TokenCount, error) {
	return 0, t.err
}
