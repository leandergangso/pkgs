# log

Small conveniences on top of `log/slog`: nicer default output and
request-scoped loggers.

```go
slog.SetDefault(log.NewJSON(os.Stdout, nil))
// or
slog.SetDefault(log.NewText(os.Stderr, nil))
```

With nil options, both handlers format timestamps as UTC RFC3339 and levels as
lowercase (`info` instead of `INFO`). To change or extend behavior, pass native
`slog.HandlerOptions`:

```go
opts := &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}
logger := log.NewText(os.Stderr, opts)
```

Your `ReplaceAttr` runs after the default normalization, so it can override
anything without losing the rest.

## Context

```go
logger := log.WithAttrs(log.From(r.Context()), slog.String("request_id", id))
ctx := log.With(r.Context(), logger)

// Deeper in the call chain:
log.From(ctx).Info("done")
```

`From` falls back to `slog.Default`, so it is safe to call on any context.
