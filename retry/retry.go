// Package retry provides bounded retries with exponential backoff and jitter.
// It is caller-driven: fn decides what to attempt, retry decides when to try
// again.
package retry

import (
	"context"
	"math/rand/v2"
	"time"
)

const (
	defaultAttempts = 3
	defaultDelay    = 100 * time.Millisecond
	defaultMaxDelay = 5 * time.Second
)

// Option configures [Do].
type Option func(*config)

type config struct {
	attempts int
	delay    time.Duration
	maxDelay time.Duration
	jitter   bool
}

// WithAttempts sets the total number of calls to fn, including the first. It
// must be positive.
func WithAttempts(n int) Option {
	return func(c *config) {
		if n <= 0 {
			panic("retry: invalid attempts")
		}
		c.attempts = n
	}
}

// WithDelay sets the delay before the first retry. It must be positive.
func WithDelay(d time.Duration) Option {
	return func(c *config) {
		if d <= 0 {
			panic("retry: invalid delay")
		}
		c.delay = d
	}
}

// WithMaxDelay caps the delay between attempts. It must be positive.
func WithMaxDelay(d time.Duration) Option {
	return func(c *config) {
		if d <= 0 {
			panic("retry: invalid max delay")
		}
		c.maxDelay = d
	}
}

// WithoutJitter disables randomization of the delay. By default the wait is
// chosen uniformly from [0, delay), which spreads out concurrent callers.
func WithoutJitter() Option {
	return func(c *config) { c.jitter = false }
}

// Do calls fn until it returns nil or the attempts are exhausted, then returns
// the last error. Between attempts it waits with exponential backoff and full
// jitter. It stops early if ctx is already canceled or becomes canceled.
func Do(ctx context.Context, fn func() error, opts ...Option) error {
	if ctx == nil {
		panic("retry: nil context")
	}
	if fn == nil {
		panic("retry: nil function")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	c := config{
		attempts: defaultAttempts,
		delay:    defaultDelay,
		maxDelay: defaultMaxDelay,
		jitter:   true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&c)
		}
	}

	var err error
	for i := 0; i < c.attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		if i == c.attempts-1 {
			break
		}
		delay := c.delay << i
		if delay <= 0 || delay > c.maxDelay {
			delay = c.maxDelay
		}
		if c.jitter {
			delay = time.Duration(rand.Int64N(int64(delay)))
		}
		if waitErr := wait(ctx, delay); waitErr != nil {
			return waitErr
		}
	}
	return err
}

func wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
