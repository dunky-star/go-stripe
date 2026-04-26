package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/dunky-star/go-stripe/internal/cards"
	"github.com/dunky-star/go-stripe/internal/models"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	CardBrand     string `json:"card_brand"`
	ExpiryMonth   int `json:"exp_month"`
	ExpiryYear    int `json:"exp_year"`
	LastFour      string `json:"last_four"`
	Plan          string `json:"plan"`
	ProductID     string `json:"product_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
}

type jsonResponse struct {
	OK           bool   `json:"ok"`
	Message      string `json:"message,omitempty"`
	Content      string `json:"content,omitempty"`
	ID           int    `json:"id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var payload stripePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: payload.Currency,
	}

	okay := true

	pi, msg, err := card.Charge(payload.Currency, amount)
	if err != nil {
		okay = false
	}

	if okay {
		j := jsonResponse{
			OK:           true,
			ClientSecret: pi.ClientSecret,
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "   ")
		if err := enc.Encode(j); err != nil {
			app.errorLog.Println(err)
		}
	} else {
		j := jsonResponse{
			OK:      false,
			Message: msg,
			Content: "",
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "   ")
		if err := enc.Encode(j); err != nil {
			app.errorLog.Println(err)
		}
	}
}

// GetWidgetByID gets one widget by id and returns as JSON
func (app *application) GetWidgetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "   ")
	if err := enc.Encode(widget); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) CreateCustomerAndSubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	var data stripePayload

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		app.errorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(jsonResponse{
			OK:      false,
			Message: "invalid request payload",
		})
		return
	}
	app.infoLog.Println(data.Email, data.LastFour, data.PaymentMethod, data.Plan)

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: data.Currency,
	}

	subscriptionID, err := card.EnsureCustomerAndSubscribe(data.PaymentMethod, data.Email, data.Plan, data.LastFour, "")
	if err != nil {
		app.errorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(jsonResponse{
			OK:      false,
			Message: cards.SafeClientMessage(err),
		})
		return
	}

	app.infoLog.Println("subscription id is", subscriptionID)

	// Idempotent DB handling: repeated API calls for the same Stripe subscription
	// should not create duplicate local transaction/order rows.
	if existingTxnID, err := app.DB.GetTransactionIDByPaymentRefs(data.PaymentMethod, subscriptionID); err == nil {
		if _, ordErr := app.DB.GetOrderIDByTransactionID(existingTxnID); ordErr == nil {
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			_ = enc.Encode(jsonResponse{
				OK:      true,
				Message: "Subscription already processed",
			})
			return
		} else if !models.IsNotFound(ordErr) {
			app.errorLog.Println(ordErr)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(jsonResponse{
				OK:      false,
				Message: "failed checking existing order",
			})
			return
		}
	} else if !models.IsNotFound(err) {
		app.errorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(jsonResponse{
			OK:      false,
			Message: "failed checking existing transaction",
		})
		return
	}

	productID, err := strconv.Atoi(data.ProductID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(jsonResponse{
			OK:      false,
			Message: "invalid product id",
		})
		return
	}

	amount, err := strconv.Atoi(data.Amount)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(jsonResponse{
			OK:      false,
			Message: "invalid amount",
		})
		return
	}

	expiryMonth := data.ExpiryMonth
	expiryYear := data.ExpiryYear

	customerID, err := app.SaveCustomer(data.FirstName, data.LastName, data.Email)
	if err != nil {
		app.errorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(jsonResponse{
			OK:      false,
			Message: "subscription created, but failed saving customer",
		})
		return
	}

	currency := data.Currency
	if currency == "" {
		currency = "usd"
	}

	txn := models.Transaction{
		Amount:              amount,
		Currency:            currency,
		LastFour:            data.LastFour,
		ExpiryMonth:         expiryMonth,
		ExpiryYear:          expiryYear,
		BankReturnCode:      subscriptionID,
		TransactionStatusID: 2,
		PaymentIntent:       subscriptionID,
		PaymentMethod:       data.PaymentMethod,
	}

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(jsonResponse{
			OK:      false,
			Message: "subscription created, but failed saving transaction",
		})
		return
	}

	order := models.Order{
		WidgetID:      productID,
		TransactionID: txnID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        amount,
	}

	if _, err := app.SaveOrder(order); err != nil {
		app.errorLog.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(jsonResponse{
			OK:      false,
			Message: "subscription created, but failed saving order",
		})
		return
	}

	okay := true

	resp := jsonResponse{
		OK:      okay,
		Message: "Subscription created and saved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		app.errorLog.Println(err)
	}
}

// SaveCustomer saves a customer and returns id.
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err == nil {
		return id, nil
	}

	// Idempotent fallback: if customer already exists locally, reuse it.
	errText := err.Error()
	if strings.Contains(errText, "Duplicate entry") && strings.Contains(errText, "for key 'email'") {
		return app.DB.GetCustomerIDByEmail(email)
	}

	return 0, err
}

// SaveTransaction saves a transaction and returns id.
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	return app.DB.InsertTransaction(txn)
}

// SaveOrder saves an order and returns id.
func (app *application) SaveOrder(order models.Order) (int, error) {
	return app.DB.InsertOrder(order)
}
