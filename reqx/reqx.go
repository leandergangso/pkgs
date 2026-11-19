// Package reqx provides small helpers for common HTTP API requests.
package reqx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

// Option configures a request. Options are created with [WithClient],
// [WithHeader], [WithQuery], and [WithRequest].
type Option func(*config)

type config struct {
	client *http.Client
	header http.Header
	query  url.Values
	modify []func(*http.Request) error
}

// WithClient sets the HTTP client. A nil client uses [http.DefaultClient].
func WithClient(client *http.Client) Option {
	return func(c *config) {
		if client != nil {
			c.client = client
		}
	}
}

// WithHeader sets a request header.
func WithHeader(key, value string) Option {
	return func(c *config) {
		c.header.Set(key, value)
	}
}

// WithQuery sets a URL query parameter.
func WithQuery(key, value string) Option {
	return func(c *config) {
		c.query.Set(key, value)
	}
}

// WithRequest applies fn to the request before it is sent.
func WithRequest(fn func(*http.Request) error) Option {
	return func(c *config) {
		if fn != nil {
			c.modify = append(c.modify, fn)
		}
	}
}

// Do sends body as JSON and decodes the JSON response into T.
//
// A non-2xx response returns an [*Error].
func Do[T any](ctx context.Context, method, rawURL string, body any, opts ...Option) (T, error) {
	var zero T

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return zero, err
		}
		reader = bytes.NewReader(data)
	}

	c := newConfig(opts)
	if body != nil && c.header.Get("Content-Type") == "" {
		c.header.Set("Content-Type", "application/json")
	}

	resp, err := c.do(ctx, method, rawURL, reader)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return zero, &Error{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       data,
		}
	}
	if resp.StatusCode == http.StatusNoContent {
		return zero, nil
	}

	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		if errors.Is(err, io.EOF) {
			return zero, nil
		}
		return zero, err
	}
	return out, nil
}

// Send sends a request and returns the raw response. The caller must close the
// response body.
func Send(ctx context.Context, method, rawURL string, body io.Reader, opts ...Option) (*http.Response, error) {
	return newConfig(opts).do(ctx, method, rawURL, body)
}

// Error is returned for non-2xx responses.
type Error struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e *Error) Error() string {
	return "reqx: " + e.Status
}

func newConfig(opts []Option) *config {
	c := &config{
		client: http.DefaultClient,
		header: make(http.Header),
		query:  make(url.Values),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

func (c *config) do(ctx context.Context, method, rawURL string, body io.Reader) (*http.Response, error) {
	if len(c.query) > 0 {
		u, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for key, values := range c.query {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		u.RawQuery = q.Encode()
		rawURL = u.String()
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	for key, values := range c.header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	for _, fn := range c.modify {
		if err := fn(req); err != nil {
			return nil, err
		}
	}
	return c.client.Do(req)
}
