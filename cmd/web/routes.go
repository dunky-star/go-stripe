package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.HomeHandler)
	mux.HandleFunc("GET /v1/virtual-terminal", app.VirtualCardHandler)
	mux.HandleFunc("POST /v1/payment-succeeded", app.PaymentSucceededHandler)
	mux.HandleFunc("GET /v1/widget/{id}", app.ChargeOnce)
	mux.HandleFunc("GET /v1/receipt", app.ReceiptHandler)
	//mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)
	//mux.HandleFunc("POST /v1/stripe", app.stripeHandler)

	// Serving static files (GET only to avoid conflict with "GET /" in Go 1.22+ ServeMux)
	fileserver := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))

	return SessionLoad(mux)
}
