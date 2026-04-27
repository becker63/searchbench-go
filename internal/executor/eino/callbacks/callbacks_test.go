package callbacks

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	einocallbacks "github.com/cloudwego/eino/callbacks"
)

func TestBuildCallbacksIsCold(t *testing.T) {
	t.Parallel()

	recorder := &FakeTestRecorder{}

	handlers, err := BuildCallbacks(context.Background(), Config{
		Factories: []Factory{
			NewFakeTestCallbackFactory(recorder),
			func(context.Context) (einocallbacks.Handler, error) {
				return nil, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("BuildCallbacks() error = %v", err)
	}
	if got, want := len(handlers), 1; got != want {
		t.Fatalf("len(handlers) = %d, want %d", got, want)
	}

	snapshot := recorder.Snapshot()
	if got, want := snapshot.Constructed, 1; got != want {
		t.Fatalf("Constructed = %d, want %d", got, want)
	}
	// Building handlers is allowed to allocate and record construction, but it
	// must not observe any execution lifecycle before the evaluator actually runs.
	if snapshot.Attached != 0 || snapshot.ModelStarts != 0 || snapshot.ToolStarts != 0 {
		t.Fatalf("snapshot after cold build = %#v, want no execution events", snapshot)
	}
}

func TestBuildCallbacksPropagatesFactoryError(t *testing.T) {
	t.Parallel()

	_, err := BuildCallbacks(context.Background(), Config{
		Factories: []Factory{
			func(context.Context) (einocallbacks.Handler, error) {
				return nil, assertErr("fixture callback failure")
			},
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	// The evaluator relies on BuildCallbacks failing closed so callback setup
	// errors become typed prepare_callbacks failures instead of being ignored.
	if !strings.Contains(err.Error(), "fixture callback failure") {
		t.Fatalf("error = %q, want fixture callback failure", err.Error())
	}
}

func TestCallbackPackageAvoidsForbiddenImports(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	dir := filepath.Dir(currentFile)
	fs := token.NewFileSet()
	// Parse only the non-test Go files in this package and inspect their import
	// lists directly. This keeps the assertion focused on the production callback
	// boundary rather than on what test helpers happen to import.
	pkgs, err := parser.ParseDir(fs, dir, func(info os.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parser.ParseDir() error = %v", err)
	}

	forbiddenSubstrings := []string{
		"internal/domain",
		"internal/score",
		"internal/report",
		"langsmith",
		"langfuse",
		"tracing",
	}

	// This is an architecture guardrail test, not a style preference test. The
	// callback package is supposed to stay execution-local and lightweight, so we
	// fail if production callback code starts depending on core domain/scoring
	// packages or on tracing-specific integrations that belong in future sibling
	// callback implementations.
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
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
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
