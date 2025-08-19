package teta

import (
	"context"
	"encoding/json/v2"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
)

type Router struct {
	*RouterGroup
	Logger
}

func New() *Router {
	defaultValidator := NewDefaultValidator()
	errorHandler := defaultHTTPErrorHandler

	return &Router{
		RouterGroup: newRouteGroup(defaultValidator, errorHandler),
		Logger:      newLogger(os.Stdout),
	}
}

func (t *Router) SetCustomValidator(v Validator) {
	t.validator = v
}

func (t *Router) SetCustomHTTPErrorHandler(handler HTTPErrorHandler) {
	t.httpErrorHandler = handler
}

func defaultHTTPErrorHandler(err error, c *Context) {
	w := c.Writer
	r := c.Request

	httpErr, ok := err.(*HTTPError)
	if !ok {
		httpErr = &HTTPError{
			code:    http.StatusBadRequest,
			message: err.Error(),
		}
	}

	slog.Error(
		"Server error",
		"error", httpErr,
		"path", r.URL.Path,
		"method", r.Method,
		"ip", r.RemoteAddr,
		"code", httpErr.code,
	)

	w.Header().Set("Content-Type", "application/json")
	json.MarshalWrite(w, &HTTPErrorMessage{httpErr.message})
}

type HTTPErrorHandler func(err error, c *Context)

type DefaultValidator struct {
	validate *validator.Validate
}

func NewDefaultValidator() *DefaultValidator {
	v := validator.New()

	return &DefaultValidator{
		validate: v,
	}
}

func (v *DefaultValidator) StructCtx(ctx context.Context, data any) error {
	return v.validate.StructCtx(ctx, data)
}

type Validator interface {
	StructCtx(ctx context.Context, s any) error
}

func (t *Router) Start(addr string) error {
	return http.ListenAndServe(addr, t.handler)
}

func (t *Router) StartTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, t.handler)
}

type StdMiddleware func(http.Handler) http.Handler

func CreateStdStack(middlewares ...StdMiddleware) StdMiddleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

func FromStdMiddleware(middlewares ...StdMiddleware) Middleware {
	mw := CreateStdStack(middlewares...)
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Writer = w
				c.Request = r
				next(c)
			})

			mw(handler).ServeHTTP(c.Writer, c.Request)

			return nil
		}
	}
}
