// Package reqx provides small helpers for sending HTTP requests.
package reqx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// Client sends HTTP requests. The zero value uses [http.DefaultClient].
type Client struct {
	// BaseURL is an optional prefix for relative request paths.
	BaseURL string
	// Header contains headers added to every request.
	Header http.Header
	// HTTP is the underlying client. If nil, [http.DefaultClient] is used.
	HTTP *http.Client
}

// NewRequest creates a request with the client's base URL and headers.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.resolve(path), body)
	if err != nil {
		return nil, err
	}
	for key, values := range c.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return req, nil
}

// Do sends req.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.HTTP == nil {
		return http.DefaultClient.Do(req)
	}
	return c.HTTP.Do(req)
}

// Get sends a GET request.
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.send(ctx, http.MethodGet, path, nil)
}

// Post sends a POST request with body and the given content type.
func (c *Client) Post(ctx context.Context, path, contentType string, body io.Reader) (*http.Response, error) {
	req, err := c.NewRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.Do(req)
}

// PostJSON sends body as a JSON request.
func (c *Client) PostJSON(ctx context.Context, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := c.NewRequest(ctx, http.MethodPost, path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.Do(req)
}

// DecodeJSON decodes resp.Body into v and closes it.
func DecodeJSON(resp *http.Response, v any) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

// ReadAll reads resp.Body and closes it.
func ReadAll(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *Client) send(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := c.NewRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) resolve(path string) string {
	if c.BaseURL == "" || strings.Contains(path, "://") {
		return path
	}
	return strings.TrimRight(c.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
}
