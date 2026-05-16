package config

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrWorkspaceSeedProviderRequired      = errors.New("config: workspace seed provider is required")
	ErrWorkspaceSeedLocalPathRequired     = errors.New("config: workspaceSeed.localPath is required when provider is local_path")
	ErrWorkspaceSeedBuckTargetRequired    = errors.New("config: workspaceSeed.buckDescriptorTarget is required when provider is buck_descriptor")
	ErrWorkspaceSeedUnknownProvider       = errors.New("config: unknown workspace seed provider")
	ErrWorkspaceSeedGitArchiveUnsupported = errors.New("config: git and archive workspace seed providers are not implemented yet")
)

// ValidateWorkspaceSeedConfig checks IC workspace seed provider wiring.
func ValidateWorkspaceSeedConfig(provider string, localPath, buckTarget *string) error {
	p := strings.TrimSpace(provider)
	if p == "" {
		return ErrWorkspaceSeedProviderRequired
	}
	switch p {
	case "local_path":
		if localPath == nil || strings.TrimSpace(*localPath) == "" {
			return ErrWorkspaceSeedLocalPathRequired
		}
	case "buck_descriptor":
		if buckTarget == nil || strings.TrimSpace(*buckTarget) == "" {
			return ErrWorkspaceSeedBuckTargetRequired
		}
	case "git", "archive":
		return fmt.Errorf("%w: %s", ErrWorkspaceSeedGitArchiveUnsupported, p)
	default:
		return fmt.Errorf("%w: %q", ErrWorkspaceSeedUnknownProvider, p)
	}
	return nil
}
