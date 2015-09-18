package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func main() {
	target, err := url.Parse("http://localhost")
	if err != nil {
		log.Fatal(err)
	}
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}
	rproxy := httputil.NewSingleHostReverseProxy(target)
	rproxy.Director = director
	rproxy.ErrorLog = log.New(os.Stdout, "", log.LstdFlags)

	// that's it! our reverse proxy is ready!
	s := &http.Server{
		Addr:    ":8088",
		Handler: rproxy,
	}
	s.ListenAndServe()
}
