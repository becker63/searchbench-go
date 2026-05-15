package logging

// Mode controls how event helpers render output.
type Mode string

const (
	// ModeNop disables all log output.
	ModeNop Mode = "nop"
	// ModeDev renders compact human-readable event lines.
	ModeDev Mode = "dev"
	// ModeJSON emits structured JSON logs through Zap.
	ModeJSON Mode = "json"
)
