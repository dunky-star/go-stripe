package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func (app *application) VirtualCardHandler(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["stripe_key"] = app.config.stripe.key
	if err := app.renderTemplate(w, r, "terminal", &templateData{
		StringMap: stringMap,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) PaymentSucceededHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Read the posted data
	paymentAmountStr := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	cardHolder := r.Form.Get("cardholder_name")
	cardholderEmail := r.Form.Get("cardholder_email")

	// Convert cents to dollars
	paymentAmountCents, _ := strconv.Atoi(paymentAmountStr)
	paymentAmount := fmt.Sprintf("%.2f", float64(paymentAmountCents)/100.0)

	data := make(map[string]interface{})
	data["cardholder"] = cardHolder
	data["cardholder_email"] = cardholderEmail
	data["payment_amount"] = paymentAmount
	data["payment_currency"] = paymentCurrency
	data["payment_intent"] = paymentIntent
	data["payment_method"] = paymentMethod

	if err := app.renderTemplate(w, r, "payment-succeeded", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}
