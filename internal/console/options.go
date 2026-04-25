package console

// Options configures console rendering for human terminal output.
type Options struct {
	Color   bool
	Width   int
	Verbose bool
}

// DefaultOptions returns the default console renderer configuration.
func DefaultOptions() Options {
	return Options{
		Color: true,
		Width: 100,
	}
}
