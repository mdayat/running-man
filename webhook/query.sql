-- name: UpdateUserSubscription :exec
UPDATE "user" SET subscription_expired_at = $2 WHERE id = $1;

-- name: GetUserIDByInvoiceID :one
SELECT user_id FROM invoice WHERE id = $1;

-- name: CreatePayment :exec
INSERT INTO payment (id, user_id, invoice_id, amount_paid, status) VALUES ($1, $2, $3, $4, $5);