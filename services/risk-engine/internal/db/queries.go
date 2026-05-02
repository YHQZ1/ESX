package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Querier interface {
	CreateLock(ctx context.Context, arg CreateLockParams) (Lock, error)
	GetLock(ctx context.Context, id uuid.UUID) (Lock, error)
	UpdateLockStatus(ctx context.Context, id uuid.UUID, status string) error
	GetCashAccount(ctx context.Context, participantID uuid.UUID) (CashAccount, error)
	GetSecuritiesAccount(ctx context.Context, participantID uuid.UUID, symbol string) (SecuritiesAccount, error)
	IncrementCashLocked(ctx context.Context, participantID uuid.UUID, amount int64) error
	DecrementCashLocked(ctx context.Context, participantID uuid.UUID, amount int64) error
	IncrementSecuritiesLocked(ctx context.Context, participantID uuid.UUID, symbol string, quantity int64) error
	DecrementSecuritiesLocked(ctx context.Context, participantID uuid.UUID, symbol string, quantity int64) error
}

type CreateLockParams struct {
	ParticipantID uuid.UUID
	Symbol        string
	Side          string
	Quantity      int64
	Price         int64
}

type Queries struct {
	riskDB        *sql.DB
	participantDB *sql.DB
}

func New(riskDB, participantDB *sql.DB) *Queries {
	return &Queries{riskDB: riskDB, participantDB: participantDB}
}

func (q *Queries) CreateLock(ctx context.Context, arg CreateLockParams) (Lock, error) {
	var l Lock
	err := q.riskDB.QueryRowContext(ctx,
		`INSERT INTO locks (participant_id, symbol, side, quantity, price)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, participant_id, symbol, side, quantity, price, status, created_at, updated_at`,
		arg.ParticipantID, arg.Symbol, arg.Side, arg.Quantity, arg.Price,
	).Scan(&l.ID, &l.ParticipantID, &l.Symbol, &l.Side, &l.Quantity, &l.Price, &l.Status, &l.CreatedAt, &l.UpdatedAt)
	return l, err
}

func (q *Queries) GetLock(ctx context.Context, id uuid.UUID) (Lock, error) {
	var l Lock
	err := q.riskDB.QueryRowContext(ctx,
		`SELECT id, participant_id, symbol, side, quantity, price, status, created_at, updated_at
		 FROM locks WHERE id = $1`,
		id,
	).Scan(&l.ID, &l.ParticipantID, &l.Symbol, &l.Side, &l.Quantity, &l.Price, &l.Status, &l.CreatedAt, &l.UpdatedAt)
	return l, err
}

func (q *Queries) UpdateLockStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := q.riskDB.ExecContext(ctx,
		`UPDATE locks SET status = $2, updated_at = now() WHERE id = $1`,
		id, status,
	)
	return err
}

func (q *Queries) GetCashAccount(ctx context.Context, participantID uuid.UUID) (CashAccount, error) {
	var a CashAccount
	err := q.participantDB.QueryRowContext(ctx,
		`SELECT id, participant_id, balance, locked FROM cash_accounts WHERE participant_id = $1`,
		participantID,
	).Scan(&a.ID, &a.ParticipantID, &a.Balance, &a.Locked)
	return a, err
}

func (q *Queries) GetSecuritiesAccount(ctx context.Context, participantID uuid.UUID, symbol string) (SecuritiesAccount, error) {
	var a SecuritiesAccount
	err := q.participantDB.QueryRowContext(ctx,
		`SELECT id, participant_id, symbol, quantity, locked FROM securities_accounts WHERE participant_id = $1 AND symbol = $2`,
		participantID, symbol,
	).Scan(&a.ID, &a.ParticipantID, &a.Symbol, &a.Quantity, &a.Locked)
	return a, err
}

func (q *Queries) IncrementCashLocked(ctx context.Context, participantID uuid.UUID, amount int64) error {
	_, err := q.participantDB.ExecContext(ctx,
		`UPDATE cash_accounts SET locked = locked + $2, updated_at = now() WHERE participant_id = $1`,
		participantID, amount,
	)
	return err
}

func (q *Queries) DecrementCashLocked(ctx context.Context, participantID uuid.UUID, amount int64) error {
	_, err := q.participantDB.ExecContext(ctx,
		`UPDATE cash_accounts SET locked = locked - $2, updated_at = now() WHERE participant_id = $1`,
		participantID, amount,
	)
	return err
}

func (q *Queries) IncrementSecuritiesLocked(ctx context.Context, participantID uuid.UUID, symbol string, quantity int64) error {
	_, err := q.participantDB.ExecContext(ctx,
		`UPDATE securities_accounts SET locked = locked + $3, updated_at = now() WHERE participant_id = $1 AND symbol = $2`,
		participantID, symbol, quantity,
	)
	return err
}

func (q *Queries) DecrementSecuritiesLocked(ctx context.Context, participantID uuid.UUID, symbol string, quantity int64) error {
	_, err := q.participantDB.ExecContext(ctx,
		`UPDATE securities_accounts SET locked = locked - $3, updated_at = now() WHERE participant_id = $1 AND symbol = $2`,
		participantID, symbol, quantity,
	)
	return err
}
