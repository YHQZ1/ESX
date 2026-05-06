package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Querier interface {
	CreateParticipant(ctx context.Context, name, email string) (Participant, error)
	GetParticipantByID(ctx context.Context, id uuid.UUID) (Participant, error)
	CreateAPIKey(ctx context.Context, participantID uuid.UUID, keyHash string) (APIKey, error)
	GetAPIKeyByHash(ctx context.Context, keyHash string) (APIKey, error)
	CreateCashAccount(ctx context.Context, participantID uuid.UUID) (CashAccount, error)
	GetCashAccount(ctx context.Context, participantID uuid.UUID) (CashAccount, error)
	Deposit(ctx context.Context, participantID uuid.UUID, amount int64) (CashAccount, error)
	GetSecuritiesAccount(ctx context.Context, participantID uuid.UUID, symbol string) (SecuritiesAccount, error)
	GetAllSecuritiesAccounts(ctx context.Context, participantID uuid.UUID) ([]SecuritiesAccount, error)
	WithTx(ctx context.Context, fn func(q Querier) error) error
}

type DBTX interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Queries struct {
	db DBTX
}

func New(db *sql.DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) CreateParticipant(ctx context.Context, name, email string) (Participant, error) {
	var p Participant
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO participants (name, email) VALUES ($1, $2) RETURNING id, name, email, status, created_at`,
		name, email,
	).Scan(&p.ID, &p.Name, &p.Email, &p.Status, &p.CreatedAt)
	return p, err
}

func (q *Queries) GetParticipantByID(ctx context.Context, id uuid.UUID) (Participant, error) {
	var p Participant
	err := q.db.QueryRowContext(ctx,
		`SELECT id, name, email, status, created_at FROM participants WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Email, &p.Status, &p.CreatedAt)
	return p, err
}

func (q *Queries) CreateAPIKey(ctx context.Context, participantID uuid.UUID, keyHash string) (APIKey, error) {
	var k APIKey
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO api_keys (participant_id, key_hash) VALUES ($1, $2) RETURNING id, participant_id, key_hash, created_at`,
		participantID, keyHash,
	).Scan(&k.ID, &k.ParticipantID, &k.KeyHash, &k.CreatedAt)
	return k, err
}

func (q *Queries) GetAPIKeyByHash(ctx context.Context, keyHash string) (APIKey, error) {
	var k APIKey
	err := q.db.QueryRowContext(ctx,
		`SELECT id, participant_id, key_hash, created_at FROM api_keys WHERE key_hash = $1`,
		keyHash,
	).Scan(&k.ID, &k.ParticipantID, &k.KeyHash, &k.CreatedAt)
	return k, err
}

func (q *Queries) CreateCashAccount(ctx context.Context, participantID uuid.UUID) (CashAccount, error) {
	var a CashAccount
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO cash_accounts (participant_id) VALUES ($1) RETURNING id, participant_id, balance, locked, currency, updated_at`,
		participantID,
	).Scan(&a.ID, &a.ParticipantID, &a.Balance, &a.Locked, &a.Currency, &a.UpdatedAt)
	return a, err
}

func (q *Queries) GetCashAccount(ctx context.Context, participantID uuid.UUID) (CashAccount, error) {
	var a CashAccount
	err := q.db.QueryRowContext(ctx,
		`SELECT id, participant_id, balance, locked, currency, updated_at FROM cash_accounts WHERE participant_id = $1`,
		participantID,
	).Scan(&a.ID, &a.ParticipantID, &a.Balance, &a.Locked, &a.Currency, &a.UpdatedAt)
	return a, err
}

func (q *Queries) Deposit(ctx context.Context, participantID uuid.UUID, amount int64) (CashAccount, error) {
	var a CashAccount
	err := q.db.QueryRowContext(ctx,
		`UPDATE cash_accounts SET balance = balance + $2, updated_at = now() WHERE participant_id = $1 RETURNING id, participant_id, balance, locked, currency, updated_at`,
		participantID, amount,
	).Scan(&a.ID, &a.ParticipantID, &a.Balance, &a.Locked, &a.Currency, &a.UpdatedAt)
	return a, err
}

func (q *Queries) GetSecuritiesAccount(ctx context.Context, participantID uuid.UUID, symbol string) (SecuritiesAccount, error) {
	var a SecuritiesAccount
	err := q.db.QueryRowContext(ctx,
		`SELECT id, participant_id, symbol, quantity, locked, updated_at FROM securities_accounts WHERE participant_id = $1 AND symbol = $2`,
		participantID, symbol,
	).Scan(&a.ID, &a.ParticipantID, &a.Symbol, &a.Quantity, &a.Locked, &a.UpdatedAt)
	return a, err
}

func (q *Queries) GetAllSecuritiesAccounts(ctx context.Context, participantID uuid.UUID) ([]SecuritiesAccount, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, participant_id, symbol, quantity, locked, updated_at FROM securities_accounts WHERE participant_id = $1`,
		participantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []SecuritiesAccount
	for rows.Next() {
		var a SecuritiesAccount
		if err := rows.Scan(&a.ID, &a.ParticipantID, &a.Symbol, &a.Quantity, &a.Locked, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (q *Queries) WithTx(ctx context.Context, fn func(q Querier) error) error {
	sqlDB, ok := q.db.(*sql.DB)
	if !ok {
		return fmt.Errorf("WithTx called on non-DB connection")
	}
	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(&Queries{db: tx}); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
