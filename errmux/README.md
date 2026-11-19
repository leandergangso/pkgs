# errmux

A small router built on `net/http`'s `ServeMux`. Patterns, wildcards, method
matching, redirects, 404s, and 405s are handled by the standard library.
Handlers return an error, and middleware is standard
`func(http.Handler) http.Handler`.

Requires Go 1.22+.

## Design

- Routing is always `http.ServeMux`: patterns, precedence, wildcards, redirects,
  404s, and 405s remain standard library behavior.
- `Router` is an `http.Handler`; serve it directly with `http.ListenAndServe`.
- Handlers return errors, and one global error handler turns them into responses.
- Middleware is standard `func(http.Handler) http.Handler`, captured at route
  registration time.
- The additions are limited to `Group`, `Use`, and `Mount`.

## Usage

```go
import (
	"log"
	"net/http"

	"github.com/leander/pkgs/errmux"
)

mux := errmux.New()
mux.Use(logging)

mux.Get("/users/{id}", getUser)
mux.HandleHTTP("GET /static/", http.FileServer(http.Dir("static")))

mux.Group("/api", func(api *errmux.Group) {
	api.Use(auth)
	api.Get("/orders", listOrders) // GET /api/orders
})

log.Fatal(http.ListenAndServe(":8080", mux))
```

`Handle` and `HandleFunc` accept error-returning handlers; a custom handler type
can implement `errmux.Handler`. Standard `http.Handler` values use `HandleHTTP`.

A `Group` is not an `http.Handler`; serve the root `Router`. Group patterns are
paths with an optional method and one space (`"GET /orders"`). Host patterns
belong on the root `Router`.

## Errors

By default, errors become `500 Internal Server Error`. Set a global handler:

```go
mux.SetErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
})
```

`SetErrorHandler` is global and live: it applies to all routes, including those
registered earlier. Unlike middleware, it is resolved when an error occurs. Call
it during setup.

`SetErrorHandler(nil)` restores the default. Do not write a response before
returning an error.
