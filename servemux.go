// Package servemux is a tiny wrapper around [http.ServeMux] whose handlers
// return errors.
//
// Patterns, wildcards, method matching, redirects, 404s, and 405s are handled
// by [http.ServeMux]. Requests and responses use the standard [http.Request]
// and [http.ResponseWriter] types, and middleware uses the standard
// func(http.Handler) http.Handler shape.
//
//	mux := servemux.New()
//	mux.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) error {
//		return json.NewEncoder(w).Encode(findUser(r.PathValue("id")))
//	})
//	http.ListenAndServe(":8080", mux)
package servemux

import (
	"net/http"
	"strings"
)

// HandlerFunc is an [http.HandlerFunc] that returns an error.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ErrorHandler handles an error returned by a [HandlerFunc].
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// Middleware wraps an [http.Handler]. The first middleware added is outermost.
type Middleware func(http.Handler) http.Handler

// DefaultErrorHandler writes a 500 response.
func DefaultErrorHandler(w http.ResponseWriter, _ *http.Request, _ error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Mux is an [http.Handler] that routes requests with an embedded
// [http.ServeMux]. The zero value is ready to use.
type Mux struct {
	mux        http.ServeMux
	parent     *Mux
	prefix     string
	middleware []Middleware
	errHandler ErrorHandler
}

// New returns a new Mux.
func New() *Mux {
	return &Mux{}
}

// ServeHTTP implements [http.Handler].
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	root := m.root()
	root.mux.ServeHTTP(w, r)
}

// Group returns a sub-mux that registers routes with prefix prepended. Groups
// inherit the parent's middleware and error handler. Register middleware
// before the routes it should apply to. Serve the root mux, not a group.
func (m *Mux) Group(prefix string) *Mux {
	return &Mux{
		parent: m,
		prefix: joinPath(m.prefix, prefix),
	}
}

// Use appends middleware to the mux. It applies to routes registered after
// the call.
func (m *Mux) Use(middleware ...Middleware) {
	m.middleware = append(m.middleware, middleware...)
}

// SetErrorHandler sets the error handler for routes registered on the mux.
// Groups inherit it unless they set their own. Passing nil restores the
// parent's handler, or [DefaultErrorHandler] on a root mux.
func (m *Mux) SetErrorHandler(handler ErrorHandler) {
	m.errHandler = handler
}

// Handle registers an [http.Handler] for pattern using
// [http.ServeMux.Handle]'s pattern syntax.
func (m *Mux) Handle(pattern string, handler http.Handler) {
	if handler == nil {
		panic("servemux: nil handler")
	}
	m.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) error {
		handler.ServeHTTP(w, r)
		return nil
	})
}

// HandleFunc registers an error-returning handler for pattern using
// [http.ServeMux.HandleFunc]'s pattern syntax.
func (m *Mux) HandleFunc(pattern string, handler HandlerFunc) {
	if handler == nil {
		panic("servemux: nil handler")
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			m.resolveErrorHandler()(w, r, err)
		}
	})

	m.root().mux.Handle(joinPattern(m.prefix, pattern), chain(h, m.middlewareChain()))
}

// Get registers handler for GET requests matching pattern.
func (m *Mux) Get(pattern string, handler HandlerFunc) {
	m.HandleFunc("GET "+pattern, handler)
}

// Post registers handler for POST requests matching pattern.
func (m *Mux) Post(pattern string, handler HandlerFunc) {
	m.HandleFunc("POST "+pattern, handler)
}

// Put registers handler for PUT requests matching pattern.
func (m *Mux) Put(pattern string, handler HandlerFunc) {
	m.HandleFunc("PUT "+pattern, handler)
}

// Patch registers handler for PATCH requests matching pattern.
func (m *Mux) Patch(pattern string, handler HandlerFunc) {
	m.HandleFunc("PATCH "+pattern, handler)
}

// Delete registers handler for DELETE requests matching pattern.
func (m *Mux) Delete(pattern string, handler HandlerFunc) {
	m.HandleFunc("DELETE "+pattern, handler)
}

// Head registers handler for HEAD requests matching pattern.
func (m *Mux) Head(pattern string, handler HandlerFunc) {
	m.HandleFunc("HEAD "+pattern, handler)
}

// Options registers handler for OPTIONS requests matching pattern.
func (m *Mux) Options(pattern string, handler HandlerFunc) {
	m.HandleFunc("OPTIONS "+pattern, handler)
}

// Connect registers handler for CONNECT requests matching pattern.
func (m *Mux) Connect(pattern string, handler HandlerFunc) {
	m.HandleFunc("CONNECT "+pattern, handler)
}

// Trace registers handler for TRACE requests matching pattern.
func (m *Mux) Trace(pattern string, handler HandlerFunc) {
	m.HandleFunc("TRACE "+pattern, handler)
}

func (m *Mux) root() *Mux {
	for m.parent != nil {
		m = m.parent
	}
	return m
}

func (m *Mux) resolveErrorHandler() ErrorHandler {
	for ; m != nil; m = m.parent {
		if m.errHandler != nil {
			return m.errHandler
		}
	}
	return DefaultErrorHandler
}

func (m *Mux) middlewareChain() []Middleware {
	var levels []*Mux
	for cur := m; cur != nil; cur = cur.parent {
		levels = append(levels, cur)
	}

	var middleware []Middleware
	for i := len(levels) - 1; i >= 0; i-- {
		middleware = append(middleware, levels[i].middleware...)
	}
	return middleware
}

func chain(handler http.Handler, middleware []Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

// joinPattern prefixes the path of pattern with prefix.
func joinPattern(prefix, pattern string) string {
	if prefix == "" {
		return pattern
	}

	method, path := "", pattern
	if i := strings.IndexByte(path, ' '); i >= 0 {
		method, path = path[:i+1], path[i+1:]
	}
	if !strings.HasPrefix(path, "/") {
		// Host-specific pattern; leave it alone.
		return method + path
	}
	return method + joinPath(prefix, path)
}

func joinPath(prefix, path string) string {
	if path == "" {
		return strings.TrimSuffix(prefix, "/")
	}
	return strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(path, "/")
}
