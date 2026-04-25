package logging

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps Zap's SugaredLogger with Searchbench-specific helpers.
//
// The zero value is safe and behaves like a nop logger.
type Logger struct {
	sugar  *zap.SugaredLogger
	mode   Mode
	out    io.Writer
	styles Styles
	name   string
	fields []any
}

// New constructs a Logger from a SugaredLogger.
//
// Passing nil returns a nop logger.
func New(sugar *zap.SugaredLogger) Logger {
	if sugar == nil {
		return NewNop()
	}
	return Logger{sugar: sugar, mode: ModeJSON}
}

// NewNop returns a logger that discards all log events.
func NewNop() Logger {
	return Logger{sugar: zap.NewNop().Sugar(), mode: ModeNop}
}

// NewDevelopment constructs a development Zap logger and its cleanup function.
func NewDevelopment() (Logger, func() error, error) {
	return NewDevelopmentWithWriter(os.Stderr, true)
}

// NewDevelopmentWithWriter constructs a pretty development logger.
func NewDevelopmentWithWriter(w io.Writer, color bool) (Logger, func() error, error) {
	return Logger{
		sugar:  zap.NewNop().Sugar(),
		mode:   ModeDev,
		out:    outputSyncer(w),
		styles: NewStyles(color),
	}, func() error { return nil }, nil
}

// NewProduction constructs a production Zap logger and its cleanup function.
func NewProduction() (Logger, func() error, error) {
	return NewProductionWithWriter(os.Stderr)
}

// NewProductionWithWriter constructs a structured JSON logger.
func NewProductionWithWriter(w io.Writer) (Logger, func() error, error) {
	base, err := newProductionLogger(w)
	if err != nil {
		return Logger{}, nil, err
	}
	logger := Logger{sugar: base.Sugar(), mode: ModeJSON}
	return logger, logger.Sync, nil
}

// Sugar returns the underlying SugaredLogger.
//
// The returned logger is always non-nil.
func (l Logger) Sugar() *zap.SugaredLogger {
	return l.base()
}

// Named returns a child logger with the given name.
func (l Logger) Named(name string) Logger {
	out := l
	if l.mode == ModeJSON {
		out.sugar = l.base().Named(name)
	} else if out.name == "" {
		out.name = name
	} else if name != "" {
		out.name = out.name + "." + name
	}
	return out
}

// With returns a child logger with additional structured context.
func (l Logger) With(args ...any) Logger {
	out := l
	if l.mode == ModeJSON {
		out.sugar = l.base().With(args...)
		return out
	}
	out.fields = append(append([]any{}, l.fields...), args...)
	return out
}

// Sync flushes any buffered log entries.
func (l Logger) Sync() error {
	if l.mode == ModeDev || l.mode == ModeNop {
		return nil
	}
	return l.base().Sync()
}

func (l Logger) base() *zap.SugaredLogger {
	if l.sugar == nil {
		return zap.NewNop().Sugar()
	}
	return l.sugar
}

func newProductionLogger(w io.Writer) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	encoder := zapcore.NewJSONEncoder(cfg.EncoderConfig)
	syncer := zapcore.AddSync(outputSyncer(w))
	core := zapcore.NewCore(encoder, syncer, cfg.Level)
	return zap.New(core), nil
}
