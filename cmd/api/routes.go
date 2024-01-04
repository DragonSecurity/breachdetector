package main

import (
	"github.com/alexedwards/flow"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := flow.New()

	mux.NotFound = http.HandlerFunc(app.notFound)
	mux.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	// mux.Use(app.logAccess)
	// mux.Use(app.recoverPanic)
	// mux.Use(app.authenticate)

	mux.HandleFunc("/status", app.status, "GET")

	return mux
}
