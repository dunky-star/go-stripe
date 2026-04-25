package cards

import (
	"fmt"
	"strings"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
)

type Card struct {
	Secret   string
	Key      string
	Currency string
}

type Transaction struct {
	TransactionStatusID int
	Amount              int
	Currency            string
	LastFour            string
	BankReturnCode      string
}

func (c *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return c.CreatePaymentIntent(currency, amount)
}

func (c *Card) CreatePaymentIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = c.Secret

	// create a payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	//params.AddMetadata("key", "value")

	pi, err := paymentintent.New(params)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}
	return pi, "", nil
}

// GetPaymentMethod gets the payment method by payment intent id
func (c *Card) GetPaymentMethod(paymentMethodID string) (*stripe.PaymentMethod, error) {
	stripe.Key = c.Secret

	pm, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return nil, err
	}
	return pm, nil
}

// RetrievePaymentIntent against an existing payment intent by id
func (c *Card) RetrievePaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error) {
	stripe.Key = c.Secret

	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

// SubscribeToPlan subscribes a stripe customer to a stripe plan
func (c *Card) SubscribeToPlan(cust *stripe.Customer, plan, email, last4, cardType, idempotencyKey string) (string, error) {
	stripeCustomerID := cust.ID
	items := []*stripe.SubscriptionItemsParams{
		{Plan: stripe.String(plan)},
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomerID),
		Items:    items,
	}

	params.AddMetadata("last_four", last4)
	params.AddMetadata("card_type", cardType)
	params.AddExpand("latest_invoice.payment_intent")
	if idempotencyKey != "" {
		params.Params.IdempotencyKey = stripe.String(idempotencyKey)
	}
	subscription, err := sub.New(params)
	if err != nil {
		return "", err
	}
	return subscription.ID, nil
}

// EnsureCustomerAndSubscribe creates/reuses customer and subscribes idempotently.
func (c *Card) EnsureCustomerAndSubscribe(pm, email, plan, last4, cardType string) (string, error) {
	cust, _, err := c.CreateCustomer(pm, email)
	if err != nil {
		if !isPaymentMethodAlreadyAttached(err) {
			return "", err
		}
		// PM already attached: recover by reusing existing customer by email.
		cust, err = c.findCustomerByEmail(email)
		if err != nil {
			return "", err
		}
	}

	idempotencyKey := fmt.Sprintf("sub:%s:%s:%s", email, plan, pm)
	return c.SubscribeToPlan(cust, plan, email, last4, cardType, idempotencyKey)
}

// CreateCustomer creates a stripe customer
func (c *Card) CreateCustomer(pm, email string) (*stripe.Customer, string, error) {
	stripe.Key = c.Secret
	customerParams := &stripe.CustomerParams{
		PaymentMethod: stripe.String(pm),
		Email:         stripe.String(email),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm),
		},
	}

	cust, err := customer.New(customerParams)
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}
	return cust, "", nil
}

func cardErrorMessage(code stripe.ErrorCode) string {
	var msg = ""
	switch code {
	case stripe.ErrorCodeCardDeclined:
		msg = "Your card was declined"
	case stripe.ErrorCodeExpiredCard:
		msg = "Your card is expired"
	case stripe.ErrorCodeIncorrectCVC:
		msg = "Incorrect CVC code"
	case stripe.ErrorCodeIncorrectZip:
		msg = "Incorrect zip/postal code"
	case stripe.ErrorCodeAmountTooLarge:
		msg = "The amount is too large to charge to your card"
	case stripe.ErrorCodeAmountTooSmall:
		msg = "The amount is too small to charge to your card"
	case stripe.ErrorCodeBalanceInsufficient:
		msg = "Insufficient balance"
	case stripe.ErrorCodePostalCodeInvalid:
		msg = "Your postal code is invalid"
	default:
		msg = "Your card was declined"
	}
	return msg
}

// SafeClientMessage returns text safe to show in the browser. Always log the full err on the server.
func SafeClientMessage(err error) string {
	if err == nil {
		return ""
	}
	se, ok := err.(*stripe.Error)
	if !ok {
		return "We couldn’t complete your request. Please try again."
	}
	if strings.Contains(strings.ToLower(se.Msg), "already been attached to a customer") {
		return "This card is already linked to a customer. If you already subscribed, refresh and check your account."
	}
	// Never expose raw Msg (can include API key fragments, request IDs).
	switch se.HTTPStatusCode {
	case 401, 403:
		return "Payment could not be verified. Please try again later."
	}
	switch se.Code {
	case stripe.ErrorCodeResourceMissing:
		return "That plan or product is not available."
	case stripe.ErrorCodeCardDeclined,
		stripe.ErrorCodeExpiredCard,
		stripe.ErrorCodeIncorrectCVC,
		stripe.ErrorCodeIncorrectZip,
		stripe.ErrorCodeAmountTooLarge,
		stripe.ErrorCodeAmountTooSmall,
		stripe.ErrorCodeBalanceInsufficient,
		stripe.ErrorCodePostalCodeInvalid:
		return cardErrorMessage(se.Code)
	}
	if se.Type == stripe.ErrorTypeCard {
		return cardErrorMessage(se.Code)
	}
	return "We couldn’t complete payment. Please try again or use a different card."
}

func isPaymentMethodAlreadyAttached(err error) bool {
	se, ok := err.(*stripe.Error)
	if !ok {
		return false
	}
	return strings.Contains(strings.ToLower(se.Msg), "already been attached to a customer")
}

func (c *Card) findCustomerByEmail(email string) (*stripe.Customer, error) {
	stripe.Key = c.Secret

	params := &stripe.CustomerListParams{
		Email: stripe.String(email),
	}
	iter := customer.List(params)
	for iter.Next() {
		return iter.Customer(), nil
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	return nil, fmt.Errorf("no customer found for email %s", email)
}
