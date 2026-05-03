package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Querier interface {
	CreateClearedTrade(ctx context.Context, arg CreateClearedTradeParams) (ClearedTrade, error)
	GetLock(ctx context.Context, id uuid.UUID) (Lock, error)
}

type CreateClearedTradeParams struct {
	TradeID     uuid.UUID
	Symbol      string
	BuyerID     uuid.UUID
	SellerID    uuid.UUID
	BuyOrderID  uuid.UUID
	SellOrderID uuid.UUID
	BuyLockID   uuid.UUID
	SellLockID  uuid.UUID
	Price       int64
	Quantity    int64
}

type Queries struct {
	clearingDB *sql.DB
	riskDB     *sql.DB
}

func New(clearingDB, riskDB *sql.DB) *Queries {
	return &Queries{clearingDB: clearingDB, riskDB: riskDB}
}

func (q *Queries) CreateClearedTrade(ctx context.Context, arg CreateClearedTradeParams) (ClearedTrade, error) {
	var t ClearedTrade
	err := q.clearingDB.QueryRowContext(ctx,
		`INSERT INTO cleared_trades (trade_id, symbol, buyer_id, seller_id, buy_order_id, sell_order_id, buy_lock_id, sell_lock_id, price, quantity)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, trade_id, symbol, buyer_id, seller_id, buy_order_id, sell_order_id, buy_lock_id, sell_lock_id, price, quantity, status, created_at`,
		arg.TradeID, arg.Symbol, arg.BuyerID, arg.SellerID, arg.BuyOrderID, arg.SellOrderID,
		arg.BuyLockID, arg.SellLockID, arg.Price, arg.Quantity,
	).Scan(&t.ID, &t.TradeID, &t.Symbol, &t.BuyerID, &t.SellerID, &t.BuyOrderID, &t.SellOrderID,
		&t.BuyLockID, &t.SellLockID, &t.Price, &t.Quantity, &t.Status, &t.CreatedAt)
	return t, err
}

func (q *Queries) GetLock(ctx context.Context, id uuid.UUID) (Lock, error) {
	var l Lock
	err := q.riskDB.QueryRowContext(ctx,
		`SELECT id, participant_id, symbol, side, quantity, price, status FROM locks WHERE id = $1`,
		id,
	).Scan(&l.ID, &l.ParticipantID, &l.Symbol, &l.Side, &l.Quantity, &l.Price, &l.Status)
	return l, err
}
