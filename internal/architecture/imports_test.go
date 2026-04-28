package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestArchitectureImportBoundaries(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	internalRoot := filepath.Join(root, "internal")

	var violations []string
	err := filepath.WalkDir(internalRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		fs := token.NewFileSet()
		file, err := parser.ParseFile(fs, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			lower := strings.ToLower(importPath)

			if strings.HasPrefix(rel, "internal/pure/") {
				if hasAnyPrefix(importPath,
					"github.com/becker63/searchbench-go/internal/adapters/",
					"github.com/becker63/searchbench-go/internal/surface/",
					"github.com/becker63/searchbench-go/internal/testing/",
					"github.com/becker63/searchbench-go/internal/app/",
				) || hasAnySubstring(lower,
					"github.com/apple/pkl-go",
					"github.com/cloudwego/eino",
					"langsmith",
					"langfuse",
					"/mcp",
					"openai",
					"openrouter",
					"cerebras",
				) || importPath == "os/exec" {
					violations = append(violations, rel+" imports forbidden dependency "+importPath)
				}
			}

			if strings.HasPrefix(rel, "internal/generic/") {
				if strings.HasPrefix(importPath, "github.com/becker63/searchbench-go/internal/") || hasAnySubstring(lower,
					"github.com/apple/pkl-go",
					"github.com/cloudwego/eino",
					"langsmith",
					"langfuse",
					"/mcp",
					"openai",
					"openrouter",
					"cerebras",
				) || importPath == "os/exec" {
					violations = append(violations, rel+" imports forbidden generic dependency "+importPath)
				}
			}

			if strings.HasPrefix(rel, "internal/ports/") {
				if hasAnyPrefix(importPath,
					"github.com/becker63/searchbench-go/internal/adapters/",
					"github.com/becker63/searchbench-go/internal/surface/",
					"github.com/becker63/searchbench-go/internal/testing/",
				) || hasAnySubstring(lower,
					"github.com/apple/pkl-go",
					"github.com/cloudwego/eino",
					"langsmith",
					"langfuse",
					"/mcp",
					"openai",
					"openrouter",
					"cerebras",
				) {
					violations = append(violations, rel+" imports forbidden ports dependency "+importPath)
				}
			}

			if strings.HasPrefix(rel, "internal/") &&
				!strings.HasPrefix(rel, "internal/testing/") &&
				!strings.HasPrefix(rel, "internal/architecture/") {
				if strings.HasPrefix(importPath, "github.com/becker63/searchbench-go/internal/testing/") {
					violations = append(violations, rel+" imports test-only package "+importPath)
				}
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk imports: %v", err)
	}

	if len(violations) > 0 {
		t.Fatalf("import boundary violations:\n%s", strings.Join(violations, "\n"))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}

func hasAnyPrefix(value string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func hasAnySubstring(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
