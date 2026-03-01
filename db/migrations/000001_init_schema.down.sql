-- Disable FK checks so drop order works for both old and new schema (transactions↔orders)
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS transaction_statuses;
DROP TABLE IF EXISTS widgets;

SET FOREIGN_KEY_CHECKS = 1;
