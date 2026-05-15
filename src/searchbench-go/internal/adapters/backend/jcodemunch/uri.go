package jcodemunch

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// filePathToRootURI turns a local filesystem path into a file-scheme URI suitable
// for MCP roots (absolute path, forward slashes in the URI path segment).
func filePathToRootURI(path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", fmt.Errorf("empty path")
	}
	abs, err := filepath.Abs(trimmed)
	if err != nil {
		return "", err
	}
	u := url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(abs),
	}
	return u.String(), nil
}
