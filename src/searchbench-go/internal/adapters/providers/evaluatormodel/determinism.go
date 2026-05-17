package evaluatormodel

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"

	run "github.com/becker63/searchbench-go/internal/pure/execution"
)

// GenerationConfig records pinned model sampling settings for live reports (#87).
type GenerationConfig struct {
	Temperature float32 `json:"temperature"`
	TopP        float32 `json:"top_p"`
	Seed        int64   `json:"seed"`
}

// DefaultLiveGenerationConfig returns deterministic defaults for live SearchBench runs.
func DefaultLiveGenerationConfig(spec run.Spec, attempt int) GenerationConfig {
	return GenerationConfig{
		Temperature: 0,
		TopP:        1,
		Seed:        stableSeed(spec, attempt),
	}
}

func stableSeed(spec run.Spec, attempt int) int64 {
	key := fmt.Sprintf("%s|%s|%s|%d",
		strings.TrimSpace(string(spec.Match.ID)),
		strings.TrimSpace(string(spec.System.ID)),
		strings.TrimSpace(string(spec.ID)),
		attempt,
	)
	sum := sha256.Sum256([]byte(key))
	return int64(binary.BigEndian.Uint64(sum[:8]))
}
