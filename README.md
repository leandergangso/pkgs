# servemux

A tiny error-returning router built on `net/http`'s `ServeMux`.

Patterns, wildcards, method matching, redirects, 404s, and 405s are all handled
by the standard library. Handlers use the usual `http.ResponseWriter` and
`*http.Request` types and return an error. Middleware is plain
`func(http.Handler) http.Handler`.

Requires Go 1.22+.

## Usage

```go
package main

import (
	"log"
	"net/http"

	"github.com/leander/servemux"
)

func main() {
	mux := servemux.New()

	mux.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("not implemented")
	})

	api := mux.Group("/api")
	api.Use(auth)
	api.Get("/orders", listOrders) // GET /api/orders

	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

## Groups

Groups inherit the parent's middleware and error handler and can override both.

```go
api := mux.Group("/api")
api.SetErrorHandler(myErrorHandler)

v1 := api.Group("/v1")
v1.Get("/status", status) // GET /api/v1/status
```

## Errors

By default, a returned error becomes a plain `500 Internal Server Error`.
Override the handler on the root mux or a group:

```go
mux.SetErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, errNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
})
```

`SetErrorHandler(nil)` restores inheritance. A handler should not write a
response before returning an error.

## Compatibility

- `*servemux.Mux` implements `http.Handler`.
- `Handle` accepts any `http.Handler`.
- Standard middleware works unchanged.
- A mux can be mounted inside a standard `http.ServeMux`:

```go
outer := http.NewServeMux()
outer.Handle("/api/", http.StripPrefix("/api", mux))
```

Middleware must be added before the routes it applies to. The first middleware
added is the outermost. A mux is not safe for concurrent registration.
