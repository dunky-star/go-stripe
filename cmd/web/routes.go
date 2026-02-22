package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/virtual-card", app.VirtualCardHandler)
	mux.HandleFunc("POST /payment-succeeded", app.PaymentSucceededHandler)
	//mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)
	//mux.HandleFunc("POST /v1/stripe", app.stripeHandler)
	return mux
}
