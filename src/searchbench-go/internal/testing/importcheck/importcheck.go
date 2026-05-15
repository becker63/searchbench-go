// Package importcheck provides test helpers that replace deprecated go/parser ParseDir usage.
package importcheck

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ParseNonTestGoFiles parses non-test *.go files in dir with the given [parser.Mode] (no ParseDir).
func ParseNonTestGoFiles(fset *token.FileSet, dir string, mode parser.Mode) ([]*ast.File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	out := make([]*ast.File, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		f, err := parser.ParseFile(fset, path, nil, mode)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, nil
}

// ParseNonTestGoFilesImportsOnly parses non-test *.go files in dir with parser.ImportsOnly.
func ParseNonTestGoFilesImportsOnly(fset *token.FileSet, dir string) ([]*ast.File, error) {
	return ParseNonTestGoFiles(fset, dir, parser.ImportsOnly)
}
