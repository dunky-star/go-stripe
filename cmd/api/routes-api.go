package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/payment-intent", app.GetPaymentIntent)
	mux.HandleFunc("GET /api/v1/widget/{id}", app.GetWidgetByID)
	mux.HandleFunc("POST /api/v1/create-customer-and-subscribe-to-plan", app.CreateCustomerAndSubscribeToPlan)
	mux.HandleFunc("POST /api/v1/authenticate", app.CreateAuthToken)
	mux.HandleFunc("POST /api/v1/is-authenticated", app.CheckAuthentication)
	mux.HandleFunc("POST /api/forgot-password", app.SendPasswordResetEmail)

	admin := http.NewServeMux()
	admin.HandleFunc("GET /test", app.adminTest)
	admin.HandleFunc("POST /virtual-terminal-succeeded", app.VirtualTerminalPaymentSucceeded)

	mux.Handle("/api/v1/admin/", app.Auth(http.StripPrefix("/api/v1/admin", admin)))

	return app.enableCORS(mux)
}
