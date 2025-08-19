package teta

import (
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi"
	"github.com/labstack/echo/v4"

	"net/http"
	"net/http/httptest"
	"testing"
)

// type mockValidator struct{}
//
// func (v mockValidator) StructCtx(ctx context.Context, s any) error { return nil }
//
// func BenchmarkContextPool(b *testing.B) {
// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest("GET", "/", nil)
// 	v := mockValidator{}
//
// 	b.ReportAllocs() // Включаем отчёт об аллокациях
//
// 	for b.Loop() {
// 		ctx := NewContext(w, r, v)
// 		ctx.Release()
// 	}
// }

func BenchmarkTetaHandle(b *testing.B) {
	r := New()
	r.Get("/foo", func(c *Context) error {
		return c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
	}
}

func BenchmarkStdRouterHandle(b *testing.B) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()

	for b.Loop() {
		mux.ServeHTTP(w, req)
	}
}

func BenchmarkChiRouterHandle(b *testing.B) {
	r := chi.NewRouter()

	r.Get("/bar", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/bar", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
	}

}

func BenchmarkGinRouterHandle(b *testing.B) {
	r := gin.New()

	r.GET("/bar", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/bar", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
	}

}

func BenchmarkEchoRouterHandle(b *testing.B) {
	r := echo.New()

	r.GET("/bar", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/bar", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
	}

}
