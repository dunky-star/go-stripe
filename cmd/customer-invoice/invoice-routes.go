package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/invoice/create-and-send", app.CreateAndSendInvoice)

	return app.enableCORS(mux)
}
