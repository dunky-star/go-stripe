package main

import (
	"net/http"
	"time"
)

type Order struct {
	ID        int       `json:"id"`
	Quantity  int       `json:"quantity"`
	Amount    int       `json:"amount"`
	Product   string    `json:"product"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (app *application) CreateAndSendInvoice(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "Hello, %s", "Create and send invoice")
	// recieve json

	// generate a pdf invoice

	// create mail

	// send mail with attachment

	// send response
}
