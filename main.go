package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/martinv13/go-shiny/internal/shinyapp"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {

	shinyapp.RunShinyApp()

	origin, _ := url.Parse("http://localhost:3000/")

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
	}

	proxy := &httputil.ReverseProxy{Director: director}

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	r.Get("/*", func(res http.ResponseWriter, req *http.Request) {
		cookie := http.Cookie{
			Name:  "SHINYPROXY_APP",
			Value: "test",
		}
		http.SetCookie(res, &cookie)
		proxy.ServeHTTP(res, req)
	})

	http.ListenAndServe(":4000", r)

}
