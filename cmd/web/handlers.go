package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dunky-star/go-stripe/internal/cards"
	"github.com/dunky-star/go-stripe/internal/models"
)

func (app *application) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) VirtualCardHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", &templateData{}, "stripe-js"); err != nil {
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
	firstname := r.Form.Get("first_name")
	lastname := r.Form.Get("last_name")
	paymentAmountStr := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	cardHolder := r.Form.Get("cardholder_name")
	cardholderEmail := r.Form.Get("cardholder_email")
	widgetID, _ := strconv.Atoi(r.Form.Get("product_id"))

	// Convert cents to dollars
	paymentAmountCents, _ := strconv.Atoi(paymentAmountStr)
	paymentAmount := fmt.Sprintf("%.2f", float64(paymentAmountCents)/100.0)

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: paymentCurrency,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	// Create a new customer
	customerID, err := app.SaveCustomer(firstname, lastname, cardholderEmail)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println(customerID)

	// Create a new transaction
	amount, _ := strconv.Atoi(paymentAmount)
	txn := models.Transaction{
		Amount:              amount,
		Currency:            paymentCurrency,
		LastFour:            lastFour,
		ExpiryMonth:         int(expiryMonth),
		ExpiryYear:          int(expiryYear),
		BankReturnCode:      pi.Charges.Data[0].ID,
		PaymentIntent:       paymentIntent,
		PaymentMethod:       paymentMethod,
		TransactionStatusID: 2,
	}

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Println(txnID)

	// Create a new order
	order := models.Order{
		WidgetID:      widgetID,
		TransactionID: txnID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        amount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	orderID, err := app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Println(orderID)

	data := make(map[string]interface{})
	data["cardholder"] = cardHolder
	data["cardholder_email"] = cardholderEmail
	data["payment_amount"] = paymentAmount
	data["payment_currency"] = paymentCurrency
	data["payment_intent"] = paymentIntent
	data["payment_method"] = paymentMethod
	data["last_four"] = lastFour
	data["expiry_month"] = expiryMonth
	data["expiry_year"] = expiryYear
	data["bank_return_code"] = pi.Charges.Data[0].ID
	data["first_name"] = firstname
	data["last_name"] = lastname

	// Should write this data to session, and then redirect user to a new page

	if err := app.renderTemplate(w, r, "payment-succeeded", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// SaveCustomer saves a new customer to the database and returns the ID
func (app *application) SaveCustomer(firstname, lastname, email string) (int, error) {

	customer := models.Customer{
		FirstName: firstname,
		LastName:  lastname,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// SaveTransaction saves a new transaction to the database and returns the ID
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {

	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// SaveOrder saves a new order to the database and returns the ID
func (app *application) SaveOrder(order models.Order) (int, error) {

	id, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// ChargeOnce displays the page to buy one widget
func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}
