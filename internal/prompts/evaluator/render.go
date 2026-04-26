package evaluator

import (
	"bytes"
	"context"
)

// Render returns the evaluator prompt as a plain string.
func Render(ctx context.Context, input Input) (string, error) {
	var buf bytes.Buffer
	if err := Prompt(input).Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
