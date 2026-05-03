package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Querier interface {
	CreateSettlement(ctx context.Context, arg CreateSettlementParams) (Settlement, error)
}

type CreateSettlementParams struct {
	TradeID     uuid.UUID
	ClearedID   uuid.UUID
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
	settlementDB  *sql.DB
	participantDB *sql.DB
	riskDB        *sql.DB
}

func New(settlementDB, participantDB, riskDB *sql.DB) *Queries {
	return &Queries{
		settlementDB:  settlementDB,
		participantDB: participantDB,
		riskDB:        riskDB,
	}
}

func (q *Queries) CreateSettlement(ctx context.Context, arg CreateSettlementParams) (Settlement, error) {
	var s Settlement
	err := q.settlementDB.QueryRowContext(ctx,
		`INSERT INTO settlements (trade_id, cleared_id, symbol, buyer_id, seller_id, buy_order_id, sell_order_id, buy_lock_id, sell_lock_id, price, quantity)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, trade_id, cleared_id, symbol, buyer_id, seller_id, buy_order_id, sell_order_id, buy_lock_id, sell_lock_id, price, quantity, status, settled_at`,
		arg.TradeID, arg.ClearedID, arg.Symbol, arg.BuyerID, arg.SellerID,
		arg.BuyOrderID, arg.SellOrderID, arg.BuyLockID, arg.SellLockID, arg.Price, arg.Quantity,
	).Scan(&s.ID, &s.TradeID, &s.ClearedID, &s.Symbol, &s.BuyerID, &s.SellerID,
		&s.BuyOrderID, &s.SellOrderID, &s.BuyLockID, &s.SellLockID, &s.Price, &s.Quantity, &s.Status, &s.SettledAt)
	return s, err
}

func (q *Queries) Settle(ctx context.Context, arg CreateSettlementParams) (Settlement, error) {
	tx, err := q.participantDB.BeginTx(ctx, nil)
	if err != nil {
		return Settlement{}, err
	}
	defer tx.Rollback()

	totalCash := arg.Price * arg.Quantity

	if _, err := tx.ExecContext(ctx,
		`UPDATE cash_accounts SET balance = balance - $2, locked = locked - $2, updated_at = now() WHERE participant_id = $1`,
		arg.BuyerID, totalCash,
	); err != nil {
		return Settlement{}, err
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE cash_accounts SET balance = balance + $2, updated_at = now() WHERE participant_id = $1`,
		arg.SellerID, totalCash,
	); err != nil {
		return Settlement{}, err
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE securities_accounts SET quantity = quantity - $3, locked = locked - $3, updated_at = now() WHERE participant_id = $1 AND symbol = $2`,
		arg.SellerID, arg.Symbol, arg.Quantity,
	); err != nil {
		return Settlement{}, err
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO securities_accounts (participant_id, symbol, quantity)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (participant_id, symbol)
		 DO UPDATE SET quantity = securities_accounts.quantity + $3, updated_at = now()`,
		arg.BuyerID, arg.Symbol, arg.Quantity,
	); err != nil {
		return Settlement{}, err
	}

	if err := tx.Commit(); err != nil {
		return Settlement{}, err
	}

	if _, err := q.riskDB.ExecContext(ctx,
		`UPDATE locks SET status = 'consumed', updated_at = now() WHERE id IN ($1, $2)`,
		arg.BuyLockID, arg.SellLockID,
	); err != nil {
		return Settlement{}, err
	}

	return q.CreateSettlement(ctx, arg)
}
