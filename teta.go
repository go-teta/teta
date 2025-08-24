package teta

import (
	"context"
	"encoding/json"
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
	w.WriteHeader(httpErr.code)
	json.NewEncoder(w).Encode(&HTTPErrorMessage{httpErr.message})
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

const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextXML                          = "text/xml"
	MIMETextXMLCharsetUTF8               = MIMETextXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

const (
	charsetUTF8   = "charset=UTF-8"
	PROPFIND      = "PROPFIND"
	REPORT        = "REPORT"
	RouteNotFound = "echo_route_not_found"
)

// Headers
const (
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderRetryAfter          = "Retry-After"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-Ip"
	HeaderXRequestID          = "X-Request-Id"
	HeaderXCorrelationID      = "X-Correlation-Id"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"
	HeaderCacheControl        = "Cache-Control"
	HeaderConnection          = "Connection"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXXSSProtection                  = "X-XSS-Protection"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderXCSRFToken                      = "X-CSRF-Token"
	HeaderReferrerPolicy                  = "Referrer-Policy"
)
