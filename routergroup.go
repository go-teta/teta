// Package teta: router
package teta

import (
	"net/http"
	"path"

	"strings"
)

type RouterGroup struct {
	prefix           string
	handler          *http.ServeMux
	middlewares      []Middleware
	parent           *RouterGroup
	validator        Validator
	httpErrorHandler HTTPErrorHandler
}

type HandlerFunc func(c *Context) error
type Middleware func(next HandlerFunc) HandlerFunc

func newRouteGroup(v Validator, her HTTPErrorHandler) *RouterGroup {
	return &RouterGroup{
		prefix:           "/",
		handler:          http.NewServeMux(),
		middlewares:      nil,
		parent:           nil,
		validator:        v,
		httpErrorHandler: her,
	}
}

func (rg *RouterGroup) next(handler HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, rg.validator)
		defer ctx.Release()

		if err := rg.applyMiddleware(handler)(ctx); err != nil {
			rg.httpErrorHandler(err, ctx)
		}
	})
}

func (rg *RouterGroup) applyMiddleware(handler HandlerFunc) HandlerFunc {
	if len(rg.middlewares) == 0 {
		return handler
	}

	compiled := handler
	for i := len(rg.middlewares) - 1; i >= 0; i-- {
		compiled = rg.middlewares[i](compiled)
	}
	return compiled
}

func (rg *RouterGroup) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rg.handler.ServeHTTP(w, r)
}

func (rg *RouterGroup) Route(pattern string, fn func(r *RouterGroup)) {
	newGroup := &RouterGroup{
		prefix:           path.Join(rg.prefix, pattern),
		handler:          rg.handler,
		middlewares:      rg.middlewares,
		parent:           rg,
		validator:        rg.validator,
		httpErrorHandler: rg.httpErrorHandler,
	}
	fn(newGroup)
}

func (rg *RouterGroup) Group(fn func(r *RouterGroup)) {
	newGroup := &RouterGroup{
		prefix:           rg.prefix,
		handler:          rg.handler,
		middlewares:      rg.middlewares,
		parent:           rg,
		validator:        rg.validator,
		httpErrorHandler: rg.httpErrorHandler,
	}
	fn(newGroup)
}

func (rg *RouterGroup) Use(middleware ...Middleware) {
	rg.middlewares = append(rg.middlewares, middleware...)
}

func (rg *RouterGroup) With(middleware ...Middleware) *RouterGroup {
	return &RouterGroup{
		prefix:           rg.prefix,
		handler:          rg.handler,
		middlewares:      append(rg.middlewares, middleware...),
		parent:           rg,
		validator:        rg.validator,
		httpErrorHandler: rg.httpErrorHandler,
	}
}

func (rg *RouterGroup) Handle(pattern string, handler HandlerFunc) {
	method, pathPattern := parsePattern(pattern)
	fullPath := path.Join(method, rg.prefix, pathPattern)

	rg.handler.Handle(fullPath, rg.next(handler))
}

func parsePattern(pattern string) (method, path string) {
	if idx := strings.Index(pattern, " "); idx != -1 {
		return pattern[:idx+1], pattern[idx+1:]
	}
	return "", pattern
}

func concatPath(method, pattern string) string {
	var b strings.Builder
	b.WriteString(method)
	b.WriteString(" ")
	b.WriteString(pattern)
	return b.String()
}

func (rg *RouterGroup) Add(method, pattern string, handler HandlerFunc) {
	rg.Handle(concatPath(method, pattern), handler)
}

func (rg *RouterGroup) Get(pattern string, handler HandlerFunc) {
	rg.Add(http.MethodGet, pattern, handler)
}

func (rg *RouterGroup) Post(pattern string, handler HandlerFunc) {
	rg.Add(http.MethodPost, pattern, handler)
}

func (rg *RouterGroup) Put(pattern string, handler HandlerFunc) {
	rg.Add(http.MethodPut, pattern, handler)
}

func (rg *RouterGroup) Delete(pattern string, handler HandlerFunc) {
	rg.Add(http.MethodDelete, pattern, handler)
}

func (rg *RouterGroup) Patch(pattern string, handler HandlerFunc) {
	rg.Add(http.MethodPatch, pattern, handler)
}

func (rg *RouterGroup) Head(pattern string, handler HandlerFunc) {
	rg.Add(http.MethodHead, pattern, handler)
}

func (rg *RouterGroup) Options(pattern string, handler HandlerFunc) {
	rg.Add(http.MethodOptions, pattern, handler)
}

func (rg *RouterGroup) Connect(pattern string, handler HandlerFunc) {
	rg.Add(http.MethodConnect, pattern, handler)
}

func (rg *RouterGroup) Trace(pattern string, handler HandlerFunc) {
	rg.Add(http.MethodTrace, pattern, handler)
}

func (rg *RouterGroup) Any(pattern string, handler HandlerFunc) {
	rg.Handle(pattern, handler)
}
