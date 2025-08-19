package teta

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Context struct {
	Writer    http.ResponseWriter
	Request   *http.Request
	Validator Validator
	ctx       context.Context
	binder    defaultBinder
}

var ctxPool = sync.Pool{
	New: func() any {
		return new(Context)
	},
}

func NewContext(w http.ResponseWriter, r *http.Request, v Validator) *Context {
	ctx := ctxPool.Get().(*Context)
	ctx.reset(w, r, v)
	return ctx
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request, v Validator) {
	c.Writer = w
	c.Request = r
	c.Validator = v
	c.ctx = r.Context()
	c.binder.r = r
}

func (c *Context) Release() {
	c.Writer = nil
	c.Request = nil
	c.Validator = nil
	c.ctx = context.Background()
	c.binder.r = nil
	ctxPool.Put(c)
}

func (c *Context) Validate(s any) error {
	return c.Validator.StructCtx(c.Request.Context(), s)
}

func (c *Context) String(status int, str string) error {
	c.Writer.Header().Add("Content-type", "application/json")
	c.Writer.WriteHeader(status)

	_, err := c.Writer.Write([]byte(str))

	return err
}

func (c *Context) JSON(status int, v any) error {
	c.Writer.Header().Add("Content-type", "application/json")
	c.Writer.WriteHeader(status)

	return json.MarshalWrite(c.Writer, &v)
}

func (c *Context) SendStatus(status int) {
	c.Writer.WriteHeader(status)
}

type ContextKey struct{ key string }

func (c *Context) Set(key string, value any) {
	c.ctx = context.WithValue(c.ctx, ContextKey{key}, value)
}

func (c *Context) Get(key string) any {
	return c.ctx.Value(ContextKey{key})
}

func (c *Context) GetString(key string) string {
	if val := c.ctx.Value(ContextKey{key}); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func (c *Context) GetInt(key string) int {
	if val := c.ctx.Value(ContextKey{key}); val != nil {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

func (c *Context) GetBool(key string) bool {
	if val := c.ctx.Value(ContextKey{key}); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
		if s, ok := val.(string); ok {
			if b, err := strconv.ParseBool(s); err == nil {
				return b
			}
		}
	}
	return false
}

func (c *Context) GetFloat(key string) float64 {
	if val := c.ctx.Value(ContextKey{key}); val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return 0
}

func (c *Context) GetStringSlice(key string) []string {
	if val := c.ctx.Value(ContextKey{key}); val != nil {
		if slice, ok := val.([]string); ok {
			return slice
		}
		if slice, ok := val.([]any); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

func (c *Context) GetMap(key string) map[string]any {
	if val := c.ctx.Value(ContextKey{key}); val != nil {
		if m, ok := val.(map[string]any); ok {
			return m
		}
	}
	return nil
}

func (c *Context) GetTime(key string) time.Time {
	if val := c.ctx.Value(ContextKey{key}); val != nil {
		switch v := val.(type) {
		case time.Time:
			return v
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func (c *Context) GetDuration(key string) time.Duration {
	if val := c.ctx.Value(ContextKey{key}); val != nil {
		switch v := val.(type) {
		case time.Duration:
			return v
		case string:
			if d, err := time.ParseDuration(v); err == nil {
				return d
			}
		case int64:
			return time.Duration(v)
		}
	}
	return 0
}

func (c *Context) Bind(dest any) error {
	return c.binder.bind(c.Request, dest)
}

func (c *Context) BindQuery(dest any) error {
	return c.binder.bindQuery(c.Request, dest)
}

func (c *Context) BindBody(dest any) error {
	return c.binder.bindBody(c.Request, dest)
}

func (c *Context) BindPath(dest any) error {
	return c.binder.bindPath(c.Request, dest)
}
