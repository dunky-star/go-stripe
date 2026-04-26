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

type transactionData struct {
	FirstName       string
	LastName        string
	Email           string
	Amount          int    // parsed payment amount (for DB)
	PaymentAmount   string // raw form value (for display)
	PaymentCurrency string
	PaymentIntentID string
	PaymentMethodID string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

// getTransactionData gets txn data from post and stripe
func (app *application) getTransactionData(r *http.Request) (transactionData, error) {
	var txnData transactionData
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}
	firstname := r.Form.Get("first_name")
	lastname := r.Form.Get("last_name")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	cardholderEmail := r.Form.Get("cardholder_email")
	PaymentAmount, err := strconv.Atoi(paymentAmount)
	if err != nil {
		return txnData, err
	}
	amount := fmt.Sprintf("%.2f", float64(PaymentAmount)/100.0)

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: paymentCurrency,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	txnData = transactionData{
		FirstName:       firstname,
		LastName:        lastname,
		Email:           cardholderEmail,
		Amount:          PaymentAmount,
		PaymentAmount:   amount,
		PaymentCurrency: paymentCurrency,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  pi.Charges.Data[0].ID,
	}

	return txnData, nil
}

func (app *application) PaymentSucceededHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Read the posted data

	widgetID, _ := strconv.Atoi(r.Form.Get("product_id"))
	txnData, err := app.getTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Create a new customer
	customerID, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.Email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println(customerID)

	// Create a new transaction
	txn := models.Transaction{
		Amount:              txnData.Amount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryMonth,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
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
		Amount:        txnData.Amount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	orderID, err := app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Println(orderID)

	// Write this data to session, and then redirect user to a new page
	app.Session.Put(r.Context(), "receipt", txnData)

	http.Redirect(w, r, "/v1/receipt", http.StatusSeeOther)
}

func (app *application) vTerminalPaymentSucceededHandler(w http.ResponseWriter, r *http.Request) {

	txnData, err := app.getTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Create a new transaction
	txn := models.Transaction{
		Amount:              txnData.Amount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryMonth,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		TransactionStatusID: 2,
	}

	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Write this data to session, and then redirect user to a new page
	app.Session.Put(r.Context(), "receipt", txnData)

	http.Redirect(w, r, "/v1/receipt", http.StatusSeeOther)
}

func (app *application) ReceiptHandler(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(transactionData)
	data := make(map[string]interface{})
	data["txn"] = txn
	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplate(w, r, "receipt", &templateData{
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

func (app *application) BronzePlanHandler(w http.ResponseWriter, r *http.Request) {
	widget, err := app.DB.GetWidget(2)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "bronze-plan", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) BronzePlanReceiptHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "receipt-plan", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}
