# reqx

Small helpers for common HTTP API requests.

```go
user, err := reqx.Do[User](ctx, http.MethodGet, "https://api.example.com/users/1")

created, err := reqx.Do[User](ctx, http.MethodPost, "https://api.example.com/users", input)

data, err := reqx.Do[map[string]any](ctx, http.MethodGet, "https://api.example.com/stats")
```

`Do` encodes the body as JSON and decodes the JSON response into the requested
type. Use `http.Method*` for the method.

## Options

```go
user, err := reqx.Do[User](ctx, http.MethodGet, url,
	reqx.WithClient(client),
	reqx.WithHeader("Authorization", "Bearer "+token),
	reqx.WithQuery("page", "2"),
)
```

`WithRequest` applies a function to the `*http.Request` before it is sent:

```go
reqx.WithRequest(func(req *http.Request) error {
	req.SetBasicAuth(user, pass)
	return nil
})
```

Non-2xx responses return `*reqx.Error`, which includes the status code, status
text, and body. For a raw response, use `Send`:

```go
resp, err := reqx.Send(ctx, http.MethodGet, url, nil)
if err != nil {
	return err
}
defer resp.Body.Close()
```
