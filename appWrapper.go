package util

import (
	"fmt"
	"net/http"

	"github.com/mholt/caddy/caddy"
	"github.com/mholt/caddy/caddy/setup"
	"github.com/mholt/caddy/middleware"
)

// Wrap an entire go web application and produce a caddy middleware to run it.
//
// requires only 2 things:
//
// newConf creates an empty config object. This will be magically populated from the caddyfile for your directive.
//
// getMux is a function that takes the populated config and returns an http.ServeMux for your application.
//
// this setup func can be registered in caddy as any other middleware
func appToDirective(newConf func() interface{}, getMux func(interface{}) *http.ServeMux) caddy.SetupFunc {
	return func(c *setup.Controller) (middleware.Middleware, error) {
		conf := newConf()
		err := Unmarshal(c, conf)
		if err != nil {
			return nil, err
		}
		fmt.Println(conf)
		mux := getMux(conf)
		return func(next middleware.Handler) middleware.Handler {
			return myMux{mux}
		}, nil
	}
}

type myMux struct {
	*http.ServeMux
}

func (m myMux) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	rec := middleware.NewResponseRecorder(w)
	m.ServeMux.ServeHTTP(rec, r)
	return rec.Status(), nil
}
