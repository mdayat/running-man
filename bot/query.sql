-- name: CheckUserExistence :one
SELECT EXISTS(SELECT 1 FROM "user" WHERE id = $1);

-- name: CreateUser :exec
INSERT INTO "user" (id, first_name) VALUES ($1, $2);

-- name: IsUserSubscribed :one
SELECT EXISTS(SELECT 1 FROM "user" WHERE id = $1 AND subscription_expired_at > NOW());

-- name: GetRunningManYears :many
SELECT year FROM library ORDER BY year ASC;

-- name: GetEpisodesByYear :many
SELECT episode FROM video WHERE library_year = $1 ORDER BY episode ASC;

-- name: GetVideoAndLibraryByEpisode :one
SELECT
  v.id AS video_id,
  l.id AS library_id,
  l.year
FROM video v JOIN library l ON v.library_year = l.year
WHERE v.episode = $1;

-- name: GetPaymentURLByInvoiceID :one
SELECT qr_url FROM invoice WHERE id = $1;

-- name: CreateInvoice :exec
INSERT INTO invoice (id, user_id, ref_id, total_amount, qr_url, expired_at) VALUES ($1, $2, $3, $4, $5, $6);

-- name: HasValidInvoice :one
SELECT EXISTS(SELECT 1 FROM invoice WHERE user_id = $1 AND expired_at > NOW());

-- name: IsInvoiceExpired :one
SELECT EXISTS(SELECT 1 FROM invoice WHERE id = $1 AND expired_at < NOW());

-- name: ValidateInvoice :one
SELECT
    expired_at < NOW() AS is_expired,
    EXISTS (SELECT 1 FROM payment WHERE invoice_id = i.id) AS is_used
FROM invoice i
WHERE i.id = $1;

-- name: CreatePayment :exec
INSERT INTO payment (id, user_id, invoice_id) VALUES ($1, $2, $3);