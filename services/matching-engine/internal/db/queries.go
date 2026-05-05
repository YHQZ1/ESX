package db

import (
	"context"
	"database/sql"
	"time"

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
	return Order{
		ID:            uuid.New(),
		ParticipantID: arg.ParticipantID,
		Symbol:        arg.Symbol,
		Side:          arg.Side,
		Type:          arg.Type,
		TimeInForce:   arg.TimeInForce,
		Quantity:      arg.Quantity,
		FilledQty:     0,
		Price:         arg.Price,
		LockID:        arg.LockID,
		Status:        "open",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
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
	return nil
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
