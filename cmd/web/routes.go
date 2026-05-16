package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", app.HomeHandler)

	admin := http.NewServeMux()
	admin.HandleFunc("GET /virtual-terminal", app.VirtualCardHandler)
	admin.HandleFunc("GET /all-sales", app.AllSales)
	admin.HandleFunc("GET /all-subscriptions", app.AllSubscriptions)
	mux.Handle("/v1/admin/", app.Auth(http.StripPrefix("/v1/admin", admin)))
	
	mux.HandleFunc("GET /v1/widget/{id}", app.ChargeOnce)
	mux.HandleFunc("POST /v1/payment-succeeded", app.PaymentSucceededHandler)
	mux.HandleFunc("GET /v1/receipt", app.ReceiptHandler)
	mux.HandleFunc("GET /v1/receipt/bronze", app.BronzePlanReceiptHandler)
	mux.HandleFunc("GET /login", app.LoginPageHandler)
	mux.HandleFunc("POST /login", app.PostLoginHandler)
	mux.HandleFunc("GET /logout", app.LogoutHandler)
	mux.HandleFunc("GET /forgot-password", app.ForgotPasswordHandler)
	mux.HandleFunc("GET /reset-password", app.ShowResetPasswordHandler)
	//mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)
	//mux.HandleFunc("POST /v1/stripe", app.stripeHandler)
	mux.HandleFunc("GET /plans/bronze", app.BronzePlanHandler)

	// Serving static files (GET only to avoid conflict with "GET /" in Go 1.22+ ServeMux)
	fileserver := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))

	return SessionLoad(mux)
}
