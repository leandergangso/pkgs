package errmux

import (
	"net/http"
	"strings"
)

var (
	_ http.Handler = (*Router)(nil)
	_ Handler      = HandlerFunc(nil)
)

// Handler handles an HTTP request and returns an error.
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) error
}

// HandlerFunc is a function that implements [Handler].
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP implements [Handler].
func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return h(w, r)
}

// ErrorHandler handles an error returned by a [Handler].
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// Middleware wraps an [http.Handler].
type Middleware = func(next http.Handler) http.Handler

// DefaultErrorHandler writes a 500 response.
func DefaultErrorHandler(w http.ResponseWriter, _ *http.Request, _ error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Router is an [http.Handler] that routes requests with an embedded
// [http.ServeMux]. The zero value is ready to use.
//
// A Router must not be copied after first use.
type Router struct {
	mux        http.ServeMux
	middleware []Middleware
	errHandler ErrorHandler
}

// New returns a new Router.
func New() *Router {
	return &Router{}
}

// ServeHTTP implements [http.Handler].
func (m *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// Use appends middleware to the router. Middleware added before registering a
// route or group applies to it.
func (m *Router) Use(middleware ...Middleware) {
	m.middleware = append(m.middleware, middleware...)
}

// SetErrorHandler sets the error handler for the router. Passing nil restores
// [DefaultErrorHandler].
func (m *Router) SetErrorHandler(handler ErrorHandler) {
	m.errHandler = handler
}

// Group registers routes under prefix. The callback receives a registration
// scope that uses the root router and inherits its current middleware.
//
// Group patterns are paths, optionally prefixed with a method and one space,
// such as "GET /users". Host-specific patterns belong on the root Router.
func (m *Router) Group(prefix string, fn func(g *Group)) {
	if fn == nil {
		panic("errmux: nil group function")
	}
	fn(&Group{
		mux:        m,
		prefix:     joinPath("", prefix),
		middleware: append([]Middleware(nil), m.middleware...),
	})
}

// Handle registers an error-returning [Handler] for pattern using
// [http.ServeMux.Handle]'s pattern syntax.
func (m *Router) Handle(pattern string, handler Handler) {
	if handler == nil {
		panic("errmux: nil handler")
	}
	m.mux.Handle(joinPattern("", pattern), chain(wrap(m, handler), m.middleware))
}

// HandleFunc registers an error-returning handler for pattern using
// [http.ServeMux.HandleFunc]'s pattern syntax.
func (m *Router) HandleFunc(pattern string, handler HandlerFunc) {
	if handler == nil {
		panic("errmux: nil handler")
	}
	m.Handle(pattern, handler)
}

// HandleHTTP registers a standard [http.Handler] for pattern using
// [http.ServeMux.Handle]'s pattern syntax.
func (m *Router) HandleHTTP(pattern string, handler http.Handler) {
	if handler == nil {
		panic("errmux: nil handler")
	}
	m.mux.Handle(joinPattern("", pattern), chain(handler, m.middleware))
}

// Get registers handler for GET requests matching pattern.
func (m *Router) Get(pattern string, handler HandlerFunc) {
	m.HandleFunc(http.MethodGet+" "+pattern, handler)
}

// Post registers handler for POST requests matching pattern.
func (m *Router) Post(pattern string, handler HandlerFunc) {
	m.HandleFunc(http.MethodPost+" "+pattern, handler)
}

// Put registers handler for PUT requests matching pattern.
func (m *Router) Put(pattern string, handler HandlerFunc) {
	m.HandleFunc(http.MethodPut+" "+pattern, handler)
}

// Patch registers handler for PATCH requests matching pattern.
func (m *Router) Patch(pattern string, handler HandlerFunc) {
	m.HandleFunc(http.MethodPatch+" "+pattern, handler)
}

// Delete registers handler for DELETE requests matching pattern.
func (m *Router) Delete(pattern string, handler HandlerFunc) {
	m.HandleFunc(http.MethodDelete+" "+pattern, handler)
}

// Head registers handler for HEAD requests matching pattern.
func (m *Router) Head(pattern string, handler HandlerFunc) {
	m.HandleFunc(http.MethodHead+" "+pattern, handler)
}

// Options registers handler for OPTIONS requests matching pattern.
func (m *Router) Options(pattern string, handler HandlerFunc) {
	m.HandleFunc(http.MethodOptions+" "+pattern, handler)
}

// Connect registers handler for CONNECT requests matching pattern.
func (m *Router) Connect(pattern string, handler HandlerFunc) {
	m.HandleFunc(http.MethodConnect+" "+pattern, handler)
}

// Trace registers handler for TRACE requests matching pattern.
func (m *Router) Trace(pattern string, handler HandlerFunc) {
	m.HandleFunc(http.MethodTrace+" "+pattern, handler)
}

// Group is a registration scope for a prefix. It is not an [http.Handler].
type Group struct {
	mux        *Router
	prefix     string
	middleware []Middleware
}

// Use appends middleware to the group. Middleware added before registering a
// route or nested group applies to it.
func (g *Group) Use(middleware ...Middleware) {
	g.middleware = append(g.middleware, middleware...)
}

// Group registers nested routes under prefix.
func (g *Group) Group(prefix string, fn func(g *Group)) {
	if fn == nil {
		panic("errmux: nil group function")
	}
	fn(&Group{
		mux:        g.mux,
		prefix:     joinPath(g.prefix, prefix),
		middleware: append([]Middleware(nil), g.middleware...),
	})
}

// Handle registers an error-returning [Handler] for a group path pattern.
func (g *Group) Handle(pattern string, handler Handler) {
	if handler == nil {
		panic("errmux: nil handler")
	}
	g.mux.mux.Handle(joinPattern(g.prefix, pattern), chain(wrap(g.mux, handler), g.middleware))
}

// HandleFunc registers an error-returning handler for a group path pattern.
func (g *Group) HandleFunc(pattern string, handler HandlerFunc) {
	if handler == nil {
		panic("errmux: nil handler")
	}
	g.Handle(pattern, handler)
}

// HandleHTTP registers a standard [http.Handler] for a group path pattern.
func (g *Group) HandleHTTP(pattern string, handler http.Handler) {
	if handler == nil {
		panic("errmux: nil handler")
	}
	g.mux.mux.Handle(joinPattern(g.prefix, pattern), chain(handler, g.middleware))
}

// Get registers handler for GET requests matching pattern.
func (g *Group) Get(pattern string, handler HandlerFunc) {
	g.HandleFunc(http.MethodGet+" "+pattern, handler)
}

// Post registers handler for POST requests matching pattern.
func (g *Group) Post(pattern string, handler HandlerFunc) {
	g.HandleFunc(http.MethodPost+" "+pattern, handler)
}

// Put registers handler for PUT requests matching pattern.
func (g *Group) Put(pattern string, handler HandlerFunc) {
	g.HandleFunc(http.MethodPut+" "+pattern, handler)
}

// Patch registers handler for PATCH requests matching pattern.
func (g *Group) Patch(pattern string, handler HandlerFunc) {
	g.HandleFunc(http.MethodPatch+" "+pattern, handler)
}

// Delete registers handler for DELETE requests matching pattern.
func (g *Group) Delete(pattern string, handler HandlerFunc) {
	g.HandleFunc(http.MethodDelete+" "+pattern, handler)
}

// Head registers handler for HEAD requests matching pattern.
func (g *Group) Head(pattern string, handler HandlerFunc) {
	g.HandleFunc(http.MethodHead+" "+pattern, handler)
}

// Options registers handler for OPTIONS requests matching pattern.
func (g *Group) Options(pattern string, handler HandlerFunc) {
	g.HandleFunc(http.MethodOptions+" "+pattern, handler)
}

// Connect registers handler for CONNECT requests matching pattern.
func (g *Group) Connect(pattern string, handler HandlerFunc) {
	g.HandleFunc(http.MethodConnect+" "+pattern, handler)
}

// Trace registers handler for TRACE requests matching pattern.
func (g *Group) Trace(pattern string, handler HandlerFunc) {
	g.HandleFunc(http.MethodTrace+" "+pattern, handler)
}

func (m *Router) errorHandler() ErrorHandler {
	if m.errHandler != nil {
		return m.errHandler
	}
	return DefaultErrorHandler
}

func wrap(m *Router, handler Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler.ServeHTTP(w, r); err != nil {
			m.errorHandler()(w, r, err)
		}
	})
}

func chain(handler http.Handler, middleware []Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

// joinPattern prefixes a group pattern. Group patterns are paths with an
// optional method and a single space, such as "GET /users".
func joinPattern(prefix, pattern string) string {
	if prefix == "" {
		return pattern
	}

	method, path, ok := strings.Cut(pattern, " ")
	if !ok {
		path = pattern
	}
	if !strings.HasPrefix(path, "/") {
		panic("errmux: group pattern path must start with /: " + pattern)
	}
	if ok {
		return method + " " + joinPath(prefix, path)
	}
	return joinPath(prefix, path)
}

func joinPath(prefix, path string) string {
	if path == "" {
		return prefix
	}
	return strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(path, "/")
}
