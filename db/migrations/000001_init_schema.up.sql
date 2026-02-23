-- go-stripe schema: aligns with internal/models/models.go
-- Column names use snake_case; use db struct tags when scanning in Go.

CREATE TABLE IF NOT EXISTS widgets (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image VARCHAR(255),
    inventory_level INT NOT NULL DEFAULT 0,
    price INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transaction_statuses (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    widget_id INT NOT NULL,
    transaction_status_id INT NOT NULL,
    quantity INT NOT NULL,
    amount INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (widget_id) REFERENCES widgets(id) ON DELETE CASCADE,
    FOREIGN KEY (transaction_status_id) REFERENCES transaction_statuses(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transactions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    transaction_status_id INT NOT NULL,
    amount INT NOT NULL,
    currency VARCHAR(10) NOT NULL,
    last_four VARCHAR(4) NOT NULL,
    bank_return_code VARCHAR(255) NOT NULL,
    payment_intent VARCHAR(255) NOT NULL,
    payment_method VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (transaction_status_id) REFERENCES transaction_statuses(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Indexes for better query performance
CREATE INDEX idx_orders_widget_id ON orders(widget_id);
CREATE INDEX idx_orders_transaction_status_id ON orders(transaction_status_id);
CREATE INDEX idx_transactions_order_id ON transactions(order_id);
CREATE INDEX idx_transactions_transaction_status_id ON transactions(transaction_status_id);

-- Seed transaction statuses
INSERT INTO transaction_statuses (name) VALUES 
    ('Pending'),
    ('Cleared'),
    ('Declined'),
    ('Refunded'),
    ('Partially Refunded');

-- Seed a sample widget
INSERT INTO widgets (name, description, image, inventory_level, price) VALUES 
    ('Widget', 'A very nice widget', '/static/widget.png', 10, 1000);
