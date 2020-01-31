package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo"
)

type CustomContext struct {
	echo.Context
}

func (c *CustomContext) Foo() {
	println("foo")
}

func (c *CustomContext) Error(err error) {
	log.Printf("CustomContext.Error, c=%p, c.Context=%p", c, c.Context)
	c.Context.Echo().HTTPErrorHandler(err, c)
	// c.Context.Error(err)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	addr := flag.String("listen", ":9090", "listen address")

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &CustomContext{c}
			err := next(cc)
			log.Printf("middleware cc=%p, c=%p", cc, c)
			if err != nil {
				cc.Error(err)
			}
			return err
		}
	})
	e.HTTPErrorHandler = customHTTPErrorHandler

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, world!")
	})
	e.GET("/err", func(c echo.Context) error {
		return errors.New("intentional error")
	})

	return e.Start(*addr)
}

func customHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	log.Printf("customHTTPErrorHandler c=%p", c)
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	c.String(code, http.StatusText(code))
	c.Logger().Error(fmt.Errorf("context type=%T, err=%v", c, err))
}
