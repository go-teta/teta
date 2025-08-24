# Teta router

## Install

```sh
go get -u github.com/teta-chi/teta

```

## Example

### Simple

```go
import (
	"log"
	"net/http"

	"github.com/go-teta/teta"
)

func main() {
	r := teta.New()

	r.Get("/", func(c *teta.Context) error {
		return c.String(http.StatusOK, "teta, not Teto!")
	})

	log.Fatal(r.Start(":8080"))
}
```

### Sub-routing

```go
package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/go-teta/teta"
)

func main() {
	r := teta.New()

	r.With(MyMiddleware).Get("/baz", func(c *teta.Context) error { // single route middleware
		return c.String(http.StatusOK, "get baz")
	})

	r.Route("/v1", func(r *teta.RouterGroup) { // sub-route level middleware
		r.Use(MyMiddleware)

		r.Get("/foo", func(c *teta.Context) error {
			return c.String(http.StatusOK, "get foo")
		})
		r.Post("/foo", func(c *teta.Context) error {
			return c.String(http.StatusOK, "post foo")
		})
	})

	r.Group(func(r *teta.RouterGroup) { // middleware groups
        r.Use(MyMiddleware)
		r.Get("/bar", func(c *teta.Context) error {
			return c.String(http.StatusOK, "get bar")
		})
	})

	log.Fatal(r.Start(":8080"))
}

func MyMiddleware(next teta.HandlerFunc) teta.HandlerFunc {
	return func(c *teta.Context) error {
		slog.Info(c.Request.URL.Path)
		return next(c)
	}
}

```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
