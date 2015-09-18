package main

import (
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/testutils"
	"github.com/mailgun/oxy/utils"
	"net/http"
	"os"
)

func main() {
	logger := utils.NewFileLogger(os.Stdout, utils.INFO)
	// Forwards incoming requests to whatever location URL points to, adds proper forwarding headers
	fwd, _ := forward.New(forward.Logger(logger))

	redirect := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// let us forward this request to another server
		req.URL = testutils.ParseURI("http://localhost")
		fwd.ServeHTTP(w, req)
	})

	// that's it! our reverse proxy is ready!
	s := &http.Server{
		Addr:    ":8088",
		Handler: redirect,
	}
	s.ListenAndServe()
}
