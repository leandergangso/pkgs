// Package log provides small conveniences on top of [log/slog]: default text
// and JSON handlers with a consistent format, and request-scoped loggers in
// context.
package log

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"time"
)

// NewText returns a logger that writes human-readable text to w. With nil opts
// it uses UTC RFC3339 timestamps and lowercase level names. Passing opts keeps
// those defaults and applies the caller's settings on top, so native slog
// options remain the way to extend or change behavior.
func NewText(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, handlerOptions(opts)))
}

// NewJSON returns a logger that writes JSON to w. Options behave as in
// [NewText].
func NewJSON(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, handlerOptions(opts)))
}

func handlerOptions(opts *slog.HandlerOptions) *slog.HandlerOptions {
	o := slog.HandlerOptions{}
	if opts != nil {
		o = *opts
	}
	base := o.ReplaceAttr
	o.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		a = normalize(groups, a)
		if base != nil {
			a = base(groups, a)
		}
		return a
	}
	return &o
}

func normalize(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		if t, ok := a.Value.Any().(time.Time); ok {
			return slog.String(slog.TimeKey, t.UTC().Format(time.RFC3339Nano))
		}
	case slog.LevelKey:
		if level, ok := a.Value.Any().(slog.Level); ok {
			return slog.String(slog.LevelKey, strings.ToLower(level.String()))
		}
	}
	return a
}

type loggerKey struct{}

// With returns a copy of ctx carrying logger.
func With(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// From returns the logger stored in ctx by [With], or [slog.Default] if none
// is present. It never returns nil.
func From(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}

// WithAttrs returns a copy of ctx whose logger has attrs attached. It is a
// shorthand for calling With with From(ctx).With(attrs...).
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return With(ctx, From(ctx).With(args...))
}
