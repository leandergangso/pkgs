# errmux

A small router built on `net/http`'s `ServeMux`. Patterns, wildcards, method
matching, redirects, 404s, and 405s are handled by the standard library.
Handlers return an error, and middleware is standard
`func(http.Handler) http.Handler`.

Requires Go 1.22+.

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

Middleware added to the router before `Group` is inherited by that group. A
`Group` is not an `http.Handler`; serve the root `Router`. Group patterns are
paths with an optional method and one space (`"GET /orders"`). Host patterns
belong on the root `Router`.

## Errors

By default, errors become `500 Internal Server Error`. Set a global handler:

```go
mux.SetErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
})
```

`SetErrorHandler(nil)` restores the default. Do not write a response before
returning an error.
