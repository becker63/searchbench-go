package config

import (
	"errors"
	"strings"
)

var (
	ErrWorkspaceSeedLocalPathRequired = errors.New("config: workspaceSeed.localPath is required when provider is local_path")
	ErrWorkspaceSeedUnknownProvider   = errors.New("config: unknown workspace seed provider (only local_path is supported on this branch)")
)

// ValidateWorkspaceSeedConfig checks local-path IC workspace seed wiring.
func ValidateWorkspaceSeedConfig(provider string, localPath *string) error {
	p := strings.TrimSpace(provider)
	if p != "local_path" {
		return ErrWorkspaceSeedUnknownProvider
	}
	if localPath == nil || strings.TrimSpace(*localPath) == "" {
		return ErrWorkspaceSeedLocalPathRequired
	}
	return nil
}
