package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// Models is the wrapper for all models
type Models struct {
	DB DBModel
}

// NewModels returns a Models type with the database connection pool
func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

// Widget is the model for the widget table
type Widget struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Price          int       `json:"price"`
	Description    string    `json:"description"`
	Image          string    `json:"image"`
	IsRecurring    bool      `json:"is_recurring"`
	PlanID         string    `json:"plan_id"`
	InventoryLevel int       `json:"inventory_level"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

// Order is the type for all orders
type Order struct {
	ID            int         `json:"id"`
	WidgetID      int         `json:"widget_id"`
	TransactionID int         `json:"transaction_id"`
	CustomerID    int         `json:"customer_id"`
	StatusID      int         `json:"status_id"`
	Quantity      int         `json:"quantity"`
	Amount        int         `json:"amount"`
	CreatedAt     time.Time   `json:"-"`
	UpdatedAt     time.Time   `json:"-"`
	Widget        Widget      `json:"widget"`
	Transaction   Transaction `json:"transaction"`
	Customer      Customer    `json:"customer"`
}

// Status is the type for order statuses
type Status struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// TransactionStatus is the type for transaction statuses
type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transaction is the type for transactions
type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	BankReturnCode      string    `json:"bank_return_code"`
	TransactionStatusID int       `json:"transaction_status_id"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

// User is the type for users
type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Customer is the type for customers
type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// GetWidget gets one widget by id
func (m *DBModel) GetWidget(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widget Widget

	row := m.DB.QueryRowContext(ctx, `
		SELECT
		    id, name, description, inventory_level, price, coalesce(image, ''),
		 	is_recurring, plan_id, created_at, updated_at
		FROM 
			widgets 
		WHERE id = ?`, id)
	err := row.Scan(
		&widget.ID,
		&widget.Name,
		&widget.Description,
		&widget.InventoryLevel,
		&widget.Price,
		&widget.Image,
		&widget.IsRecurring,
		&widget.PlanID,
		&widget.CreatedAt,
		&widget.UpdatedAt,
	)
	if err != nil {
		return widget, err
	}

	return widget, nil
}

// InsertTransaction inserts a new txn, and returns its id
func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO transactions
			(amount, currency, last_four, bank_return_code, 
			  expiry_month, expiry_year, payment_intent, payment_method,
			transaction_status_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		txn.ExpiryMonth,
		txn.ExpiryYear,
		txn.PaymentIntent,
		txn.PaymentMethod,
		txn.TransactionStatusID,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// InsertOrder inserts a new order, and returns its id
func (m *DBModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO orders 
		  (widget_id, customer_id, transaction_id, transaction_status_id, quantity, amount, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := m.DB.ExecContext(ctx, stmt,
		order.WidgetID,
		order.CustomerID,
		order.TransactionID,
		order.StatusID,
		order.Quantity,
		order.Amount,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// InsertCustomer inserts a new customer, and returns its id
func (m *DBModel) InsertCustomer(customer Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO customers
		(first_name, last_name, email, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		`
	result, err := m.DB.ExecContext(ctx, stmt,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetCustomerIDByEmail returns customer id for an existing email.
func (m *DBModel) GetCustomerIDByEmail(email string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	row := m.DB.QueryRowContext(ctx, `SELECT id FROM customers WHERE email = ?`, email)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetTransactionIDByPaymentRefs returns transaction id for payment_method + payment_intent.
func (m *DBModel) GetTransactionIDByPaymentRefs(paymentMethod, paymentIntent string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	row := m.DB.QueryRowContext(ctx, `
		SELECT id
		FROM transactions
		WHERE payment_method = ? AND payment_intent = ?
		LIMIT 1
	`, paymentMethod, paymentIntent)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetOrderIDByTransactionID returns order id for an existing transaction.
func (m *DBModel) GetOrderIDByTransactionID(transactionID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	row := m.DB.QueryRowContext(ctx, `
		SELECT id
		FROM orders
		WHERE transaction_id = ?
		LIMIT 1
	`, transactionID)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// IsNotFound reports whether a DB query returned no rows.
func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// GetUserByEmail gets a user by email address.
func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	email = strings.ToLower(email)
	var u User

	row := m.DB.QueryRowContext(ctx, `
		SELECT
			id, first_name, last_name, email, password
		FROM
			users
		WHERE email = ?`, email)

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
	)
	if err != nil {
		return u, err
	}

	return u, nil
}

// GetUserByID loads a user row by primary key (used after Authenticate).
func (m *DBModel) GetUserByID(id int) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u User
	row := m.DB.QueryRowContext(ctx, `
		SELECT
			id, first_name, last_name, email, password
		FROM
			users
		WHERE id = ?`, id)

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
	)
	if err != nil {
		return u, err
	}

	return u, nil
}

// Authenticate returns the user id if email and password match.
func (m *DBModel) Authenticate(email, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, `SELECT id, password FROM users WHERE email = ?`, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, errors.New("incorrect password")
	} else if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *DBModel) UpdatePasswordForUser(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update users set password = ? where id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, hash, u.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetAllOrdersPaginated returns a slice of a subset of orders
func (m *DBModel) GetAllOrdersPaginated(pageSize, page int) ([]*Order, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offset := (page - 1) * pageSize

	var orders []*Order

	query := `
	select
		o.id, o.widget_id, o.transaction_id, o.customer_id, 
		o.transaction_status_id, o.quantity, o.amount, o.created_at,
		o.updated_at, w.id, w.name, t.id, t.amount, t.currency,
		t.last_four, t.expiry_month, t.expiry_year, t.payment_intent,
		t.bank_return_code, c.id, c.first_name, c.last_name, c.email
		
	from
		orders o
		left join widgets w on (o.widget_id = w.id)
		left join transactions t on (o.transaction_id = t.id)
		left join customers c on (o.customer_id = c.id)
	where
		w.is_recurring = 0
	order by
		o.created_at desc
	limit ? offset ?
	`

	rows, err := m.DB.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Widget.ID,
			&o.Widget.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, 0, 0, err
		}
		orders = append(orders, &o)
	}

	query = `
		select 
			count(o.id)
		from 
			orders o
			left join widgets w on (o.widget_id = w.id)
		where
			w.is_recurring = 0
	`
	var totalRecords int
	countRow := m.DB.QueryRowContext(ctx, query)
	err = countRow.Scan(&totalRecords)
	if err != nil {
		return nil, 0, 0, err
	}

	lastPage := totalRecords / pageSize

	return orders, lastPage, totalRecords, nil
}

// GetAllSubscriptionsPaginated returns a slice of a subset of subscriptions
func (m *DBModel) GetAllSubscriptionsPaginated(pageSize, page int) ([]*Order, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offset := (page - 1) * pageSize

	var orders []*Order

	query := `
	select
		o.id, o.widget_id, o.transaction_id, o.customer_id, 
		o.transaction_status_id, o.quantity, o.amount, o.created_at,
		o.updated_at, w.id, w.name, t.id, t.amount, t.currency,
		t.last_four, t.expiry_month, t.expiry_year, t.payment_intent,
		t.bank_return_code, c.id, c.first_name, c.last_name, c.email
		
	from
		orders o
		left join widgets w on (o.widget_id = w.id)
		left join transactions t on (o.transaction_id = t.id)
		left join customers c on (o.customer_id = c.id)
	where
		w.is_recurring = 1
	order by
		o.created_at desc
	limit ? offset ?
	`

	rows, err := m.DB.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Widget.ID,
			&o.Widget.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, 0, 0, err
		}
		orders = append(orders, &o)
	}

	query = `
		select 
			count(o.id)
		from 
			orders o
			left join widgets w on (o.widget_id = w.id)
		where
			w.is_recurring = 1
	`
	var totalRecords int
	countRow := m.DB.QueryRowContext(ctx, query)
	err = countRow.Scan(&totalRecords)
	if err != nil {
		return nil, 0, 0, err
	}

	lastPage := totalRecords / pageSize

	return orders, lastPage, totalRecords, nil
}

// GetOrderByID gets one order by id and returns the Order
func (m *DBModel) GetOrderByID(id int) (Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o Order

	query := `
		select
			o.id, o.widget_id, o.transaction_id, o.customer_id,
			o.transaction_status_id, o.quantity, o.amount, o.created_at,
			o.updated_at, w.id, w.name, t.id, t.amount, t.currency,
			t.last_four, t.expiry_month, t.expiry_year, t.payment_intent,
			t.bank_return_code, c.id, c.first_name, c.last_name, c.email
		from
			orders o
			left join widgets w on (o.widget_id = w.id)
			left join transactions t on (o.transaction_id = t.id)
			left join customers c on (o.customer_id = c.id)
		where
			o.id = ?
	`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&o.ID,
		&o.WidgetID,
		&o.TransactionID,
		&o.CustomerID,
		&o.StatusID,
		&o.Quantity,
		&o.Amount,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.Widget.ID,
		&o.Widget.Name,
		&o.Transaction.ID,
		&o.Transaction.Amount,
		&o.Transaction.Currency,
		&o.Transaction.LastFour,
		&o.Transaction.ExpiryMonth,
		&o.Transaction.ExpiryYear,
		&o.Transaction.PaymentIntent,
		&o.Transaction.BankReturnCode,
		&o.Customer.ID,
		&o.Customer.FirstName,
		&o.Customer.LastName,
		&o.Customer.Email,
	)
	if err != nil {
		return o, err
	}

	return o, nil
}

func (m *DBModel) UpdateOrderStatus(id, statusID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "update orders set transaction_status_id = ? where id = ?"

	_, err := m.DB.ExecContext(ctx, stmt, statusID, id)
	if err != nil {
		return err
	}

	return nil
}
