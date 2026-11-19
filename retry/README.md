# retry

Bounded retries with exponential backoff and jitter. No dependencies.

```go
err := retry.Do(ctx, func() error {
	return submit(ctx)
},
	retry.WithAttempts(4),
	retry.WithDelay(50*time.Millisecond),
	retry.WithMaxDelay(2*time.Second),
)
```

`Do` returns the last error if every attempt fails, or nil on the first
success. Every non-nil error is retried, so `fn` is responsible for deciding
what is worth retrying. Defaults are 3 attempts, a 100ms initial delay, and a
5s cap.

The delay doubles after each failure and is capped at `WithMaxDelay`. By
default the actual wait is chosen uniformly from `[0, delay)` (full jitter) to
spread out concurrent callers; `WithoutJitter` makes it deterministic. `Do`
returns `ctx.Err()` if the context is canceled before a retry.
