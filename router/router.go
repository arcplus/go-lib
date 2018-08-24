package router

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

// HandlerFunc is http handler func
type HandlerFunc = negroni.HandlerFunc

// Wrap http.HandlerFunc to HandlerFunc
func Wrap(handler http.HandlerFunc) HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		handler.ServeHTTP(rw, r)
		next(rw, r)
	}
}

// Router is http router
type Router struct {
	path     string
	handlers []HandlerFunc
	router   *httprouter.Router
}

func (r *Router) joinPath(path string) string {
	if (r.path + path)[0] != '/' {
		panic("path should start with '/' '" + path + "'")
	}

	return r.path + path
}

// New gen a new http router
func New(handlers ...HandlerFunc) *Router {
	return &Router{
		handlers: handlers,
		router:   httprouter.New(),
	}
}

// Group is http group with prefix
func (r *Router) Group(path string, handlers ...HandlerFunc) *Router {
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	return &Router{
		handlers: append(r.handlers, handlers...),
		path:     r.joinPath(path),
		router:   r.router,
	}
}

// UseFunc use handle func
func (r *Router) UseFunc(handlers ...HandlerFunc) {
	r.handlers = append(r.handlers, handlers...)
}

// Handler handler func
func (r *Router) Handler(method, path string, handlers ...HandlerFunc) {
	n := negroni.New()

	handlers = append(r.handlers, handlers...)

	for i := range handlers {
		n.UseFunc(handlers[i])
	}

	r.router.Handler(method, r.joinPath(path), n)
}

// POST http post method
func (r *Router) POST(path string, handlers ...HandlerFunc) {
	r.Handler(http.MethodPost, path, handlers...)
}

// GET http get method
func (r *Router) GET(path string, handlers ...HandlerFunc) {
	r.Handler(http.MethodGet, path, handlers...)
}

// HEAD http head method
func (r *Router) HEAD(path string, handlers ...HandlerFunc) {
	r.Handler(http.MethodHead, path, handlers...)
}

// OPTIONS http options method
func (r *Router) OPTIONS(path string, handlers ...HandlerFunc) {
	r.Handler(http.MethodOptions, path, handlers...)
}

// Any for suppor all http method
func (r *Router) Any(path string, handlers ...HandlerFunc) {
	r.Handler(http.MethodGet, path, handlers...)
	r.Handler(http.MethodPost, path, handlers...)
	r.Handler(http.MethodPut, path, handlers...)
	r.Handler(http.MethodPatch, path, handlers...)
	r.Handler(http.MethodHead, path, handlers...)
	r.Handler(http.MethodOptions, path, handlers...)
	r.Handler(http.MethodDelete, path, handlers...)
	r.Handler(http.MethodConnect, path, handlers...)
	r.Handler(http.MethodTrace, path, handlers...)
}

// NotFound for 404 handler
func (r *Router) NotFound(handleFunc http.HandlerFunc) {
	r.router.NotFound = handleFunc
}

// PanicHandler for panic handler
func (r *Router) PanicHandler(handler func(http.ResponseWriter, *http.Request, interface{})) {
	r.router.PanicHandler = handler
}

// ServeHTTP serve http
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(rw, req)
}

// GetParams get params from request or request.Context()
func GetParams(r interface{}) httprouter.Params {
	switch v := r.(type) {
	case *http.Request:
		return httprouter.ParamsFromContext(v.Context())
	case context.Context:
		return httprouter.ParamsFromContext(v)
	}
	return nil
}
