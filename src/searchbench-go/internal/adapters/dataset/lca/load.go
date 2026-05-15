package lca

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func safePathSegment(label, value string) (string, error) {
	s := strings.TrimSpace(value)
	if s == "" {
		return "", fmt.Errorf("%s is required", label)
	}
	if strings.Contains(s, "..") {
		return "", fmt.Errorf("%s %q must not contain parent directory segments", label, value)
	}
	if strings.ContainsAny(s, `/\`) {
		return "", fmt.Errorf("%s %q must be a single path segment", label, value)
	}
	return s, nil
}

// ResolveJetBrainsLCASlicePath returns the JSONL path for one config/split under
// manifestDir/datasets/<slug(name)>/<config>/<split>.jsonl .
func ResolveJetBrainsLCASlicePath(manifestDir, datasetName, config, split string) (string, error) {
	base := strings.TrimSpace(manifestDir)
	if base == "" {
		return "", fmt.Errorf("manifest directory is required")
	}
	cfg, err := safePathSegment("dataset config", config)
	if err != nil {
		return "", err
	}
	spl, err := safePathSegment("dataset split", split)
	if err != nil {
		return "", err
	}
	slug := slugDatasetName(datasetName)
	if slug == "" {
		return "", fmt.Errorf("dataset name is required")
	}
	return filepath.Join(base, "datasets", slug, cfg, spl+".jsonl"), nil
}

func loadRowsJSONL(path string) ([]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, bufio.MaxScanTokenSize)

	var rows []json.RawMessage
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		rows = append(rows, json.RawMessage(line))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("dataset file %s has no JSON rows", path)
	}
	return rows, nil
}
