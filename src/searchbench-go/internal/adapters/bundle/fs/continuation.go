package bundlefs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	pureround "github.com/becker63/searchbench-go/internal/pure/round"
)

var (
	ErrIncompleteBundle      = errors.New("bundle: completed marker is missing")
	ErrMissingContinuation   = errors.New("bundle: continuation.json is missing")
	ErrMalformedContinuation = errors.New("bundle: continuation.json is malformed")
)

// LoadContinuation requires an explicit completed round bundle path and reads
// the continuation metadata that later manifests continue from.
func LoadContinuation(bundlePath domain.HostPath) (pureround.Continuation, error) {
	if _, err := os.Stat(filepath.Join(string(bundlePath), completeMarkerName)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return pureround.Continuation{}, ErrIncompleteBundle
		}
		return pureround.Continuation{}, err
	}

	data, err := os.ReadFile(filepath.Join(string(bundlePath), "continuation.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return pureround.Continuation{}, ErrMissingContinuation
		}
		return pureround.Continuation{}, err
	}

	var continuation pureround.Continuation
	if err := json.Unmarshal(data, &continuation); err != nil {
		return pureround.Continuation{}, fmt.Errorf("%w: %w", ErrMalformedContinuation, err)
	}
	if err := continuation.Validate(); err != nil {
		return pureround.Continuation{}, fmt.Errorf("%w: %w", ErrMalformedContinuation, err)
	}
	return continuation, nil
}
