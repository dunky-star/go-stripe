package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/payment-intent", app.GetPaymentIntent)
	mux.HandleFunc("GET /api/v1/widget/{id}", app.GetWidgetByID)
	mux.HandleFunc("POST /api/v1/create-customer-and-subscribe-to-plan", app.CreateCustomerAndSubscribeToPlan)

	return app.enableCORS(mux)
}
