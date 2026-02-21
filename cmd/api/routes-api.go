package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	// mux.HandleFunc("GET /v1/virtual-card", app.VirtualCardHandler)
	return app.enableCORS(mux)
}
