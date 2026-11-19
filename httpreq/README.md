# httpreq

Small helpers for common HTTP API requests, built on `encoding/json/v2`.

```go
user, err := httpreq.Do[User](ctx, http.MethodGet, "https://api.example.com/users/1")

created, err := httpreq.Do[User](ctx, http.MethodPost, "https://api.example.com/users", input)

data, err := httpreq.Do[map[string]any](ctx, http.MethodGet, "https://api.example.com/stats")
```

`Do` encodes the body as JSON and decodes the JSON response into the requested
type. Use `http.Method*` for the method. An empty response body returns the
zero value of `T`.

## Options

```go
user, err := httpreq.Do[User](ctx, http.MethodGet, url,
	httpreq.WithClient(client),
	httpreq.WithHeader("Authorization", "Bearer "+token),
	httpreq.WithQuery("page", "2"),
)
```

`WithRequest` applies a function to the `*http.Request` before it is sent:

```go
httpreq.WithRequest(func(req *http.Request) error {
	req.SetBasicAuth(user, pass)
	return nil
})
```

Non-2xx responses return `*httpreq.Error`, which includes the status code,
status text, and body. For a raw response, use `Send`:

```go
resp, err := httpreq.Send(ctx, http.MethodGet, url, nil)
if err != nil {
	return err
}
defer resp.Body.Close()
```

## JSON behavior

`encoding/json/v2` uses stricter defaults than `encoding/json`: duplicate
object names and invalid UTF-8 are rejected, and struct field names are matched
case-sensitively. v2 options can restore the older behavior, but `httpreq` does
not expose them, so it is best for APIs that follow normal JSON conventions.
