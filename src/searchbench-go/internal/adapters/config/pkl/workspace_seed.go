package config

import (
	"errors"
	"strings"
)

var (
	ErrWorkspaceSeedBuckTargetRequired = errors.New("config: workspaceSeed.buckDescriptorTarget is required when provider is buck_descriptor")
	ErrWorkspaceSeedUnknownProvider    = errors.New("config: unknown workspace seed provider (only buck_descriptor is supported on this branch)")
)

// ValidateWorkspaceSeedConfig checks Buck-descriptor IC workspace seed wiring.
func ValidateWorkspaceSeedConfig(provider string, buckTarget *string) error {
	p := strings.TrimSpace(provider)
	if p != "buck_descriptor" {
		return ErrWorkspaceSeedUnknownProvider
	}
	if buckTarget == nil || strings.TrimSpace(*buckTarget) == "" {
		return ErrWorkspaceSeedBuckTargetRequired
	}
	return nil
}
