package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Querier interface {
	CreateOrder(ctx context.Context, arg CreateOrderParams) (Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (Order, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string, filledQty int64) error
	CancelOrder(ctx context.Context, id uuid.UUID, participantID uuid.UUID) (Order, error)
}

type CreateOrderParams struct {
	ParticipantID uuid.UUID
	Symbol        string
	Side          string
	Type          string
	TimeInForce   string
	Quantity      int64
	Price         int64
	LockID        uuid.UUID
}

type Queries struct {
	db *sql.DB
}

func New(db *sql.DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) CreateOrder(ctx context.Context, arg CreateOrderParams) (Order, error) {
	var o Order
	err := q.db.QueryRowContext(ctx,
		`INSERT INTO orders (participant_id, symbol, side, type, time_in_force, quantity, price, lock_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, participant_id, symbol, side, type, time_in_force, quantity, filled_qty, price, lock_id, status, created_at, updated_at`,
		arg.ParticipantID, arg.Symbol, arg.Side, arg.Type, arg.TimeInForce, arg.Quantity, arg.Price, arg.LockID,
	).Scan(&o.ID, &o.ParticipantID, &o.Symbol, &o.Side, &o.Type, &o.TimeInForce, &o.Quantity, &o.FilledQty, &o.Price, &o.LockID, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (q *Queries) GetOrder(ctx context.Context, id uuid.UUID) (Order, error) {
	var o Order
	err := q.db.QueryRowContext(ctx,
		`SELECT id, participant_id, symbol, side, type, time_in_force, quantity, filled_qty, price, lock_id, status, created_at, updated_at
		 FROM orders WHERE id = $1`,
		id,
	).Scan(&o.ID, &o.ParticipantID, &o.Symbol, &o.Side, &o.Type, &o.TimeInForce, &o.Quantity, &o.FilledQty, &o.Price, &o.LockID, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (q *Queries) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string, filledQty int64) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE orders SET status = $2, filled_qty = $3, updated_at = now() WHERE id = $1`,
		id, status, filledQty,
	)
	return err
}

func (q *Queries) CancelOrder(ctx context.Context, id uuid.UUID, participantID uuid.UUID) (Order, error) {
	var o Order
	err := q.db.QueryRowContext(ctx,
		`UPDATE orders SET status = 'cancelled', updated_at = now()
		 WHERE id = $1 AND participant_id = $2 AND status = 'open'
		 RETURNING id, participant_id, symbol, side, type, time_in_force, quantity, filled_qty, price, lock_id, status, created_at, updated_at`,
		id, participantID,
	).Scan(&o.ID, &o.ParticipantID, &o.Symbol, &o.Side, &o.Type, &o.TimeInForce, &o.Quantity, &o.FilledQty, &o.Price, &o.LockID, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}
