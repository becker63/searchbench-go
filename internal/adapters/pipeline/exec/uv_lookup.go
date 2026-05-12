package execpipeline

import (
	"fmt"
	"os/exec"
)

// LookupUvBinary resolves the uv executable from PATH.
func LookupUvBinary() (string, error) {
	path, err := exec.LookPath("uv")
	if err != nil {
		return "", fmt.Errorf("uv executable not found on PATH: %w", err)
	}
	return path, nil
}
