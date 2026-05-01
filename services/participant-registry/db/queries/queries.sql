-- name: CreateParticipant :one
INSERT INTO participants (name, email)
VALUES ($1, $2)
RETURNING *;

-- name: GetParticipantByID :one
SELECT * FROM participants WHERE id = $1;

-- name: CreateAPIKey :one
INSERT INTO api_keys (participant_id, key_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1;

-- name: CreateCashAccount :one
INSERT INTO cash_accounts (participant_id)
VALUES ($1)
RETURNING *;

-- name: GetCashAccount :one
SELECT * FROM cash_accounts WHERE participant_id = $1;

-- name: Deposit :one
UPDATE cash_accounts
SET balance = balance + $2, updated_at = now()
WHERE participant_id = $1
RETURNING *;

-- name: GetSecuritiesAccount :one
SELECT * FROM securities_accounts
WHERE participant_id = $1 AND symbol = $2;

-- name: GetAllSecuritiesAccounts :many
SELECT * FROM securities_accounts WHERE participant_id = $1;