-- name: CreatePayment :exec
INSERT INTO payment (id, user_id, invoice_id, amount_paid, status) VALUES ($1, $2, $3, $4, $5);