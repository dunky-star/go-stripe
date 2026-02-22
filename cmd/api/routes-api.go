package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/payment-intent", app.GetPaymentIntent)

	return app.enableCORS(mux)
}
