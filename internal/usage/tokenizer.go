package usage

import (
	"errors"
	"strings"

	"github.com/becker63/searchbench-go/internal/domain"
)

// ErrTokenizerUnavailable reports that local token estimation could not run
// because no working tokenizer was available.
var ErrTokenizerUnavailable = errors.New("usage tokenizer unavailable")

// Tokenizer estimates token counts from provider-neutral text segments.
type Tokenizer interface {
	CountStrings(parts []string) (domain.TokenCount, error)
}

// WhitespaceTokenizer is the deterministic fallback tokenizer used by default
// in offline tests and harness-local estimation paths.
//
// This is intentionally simple. It provides stable estimates without pulling in
// provider-specific SDKs or trace backends.
type WhitespaceTokenizer struct{}

// CountStrings estimates tokens by counting whitespace-separated fields across
// all supplied text parts.
func (WhitespaceTokenizer) CountStrings(parts []string) (domain.TokenCount, error) {
	var count int64
	for _, part := range parts {
		count += int64(len(strings.Fields(part)))
	}
	return domain.TokenCount(count), nil
}
