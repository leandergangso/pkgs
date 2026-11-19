# assert

Checks programming invariants in production code. A failed assertion panics
with `*assert.Violation`, which includes the call site and can be matched with
`errors.As`.

```go
func New(opts Options) (*Server, error) {
	assert.True(opts.Port > 0, "port must be positive")
	assert.NotNil(opts.Logger, "logger is required")

	// ...
}

func handle(s state) {
	switch s {
	case stateReady, stateBusy:
		// ...
	default:
		assert.Fail("unknown state: %v", s)
	}
}
```

Use assertions only for conditions that indicate a bug: unreachable branches,
violated preconditions, and impossible states. Never use them for user input or
expected I/O failures — return an error for those.

There is no build tag to disable assertions; they always run. Keep them cheap
and reserve them for invariants worth failing fast on.
