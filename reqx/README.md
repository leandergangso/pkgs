# reqx

Small helpers for sending HTTP requests with `net/http`.

`reqx` does not hide `http.Request` or `http.Response`. Use `NewRequest` and `Do`
when you need full control.

## Usage

```go
client := &reqx.Client{
	BaseURL: "https://api.example.com",
	Header: http.Header{
		"Authorization": {"Bearer " + token},
	},
	HTTP: &http.Client{Timeout: 10 * time.Second},
}

resp, err := client.Get(ctx, "/users/1")
if err != nil {
	return err
}

var user User
if err := reqx.DecodeJSON(resp, &user); err != nil {
	return err
}

resp, err = client.PostJSON(ctx, "/users", user)
```

`Get` and `Post` return a raw response, so close its body when you are done.
`DecodeJSON` and `ReadAll` close it for you.

## Full control

```go
req, err := client.NewRequest(ctx, http.MethodPut, "/users/1", body)
if err != nil {
	return err
}
req.Header.Set("If-Match", etag)

resp, err := client.Do(req)
```
