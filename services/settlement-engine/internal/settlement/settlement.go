package settlement

import (
	"context"
	"fmt"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/settlement-engine/internal/db"
	"github.com/google/uuid"
)

type Settler struct {
	db  *db.Queries
	log *logger.Logger
}

func New(database *db.Queries, log *logger.Logger) *Settler {
	return &Settler{db: database, log: log}
}

func (s *Settler) Settle(ctx context.Context, arg SettleParams) (db.Settlement, error) {
	settled, err := s.db.Settle(ctx, db.CreateSettlementParams{
		TradeID:     arg.TradeID,
		ClearedID:   arg.ClearedID,
		Symbol:      arg.Symbol,
		BuyerID:     arg.BuyerID,
		SellerID:    arg.SellerID,
		BuyOrderID:  arg.BuyOrderID,
		SellOrderID: arg.SellOrderID,
		BuyLockID:   arg.BuyLockID,
		SellLockID:  arg.SellLockID,
		Price:       arg.Price,
		Quantity:    arg.Quantity,
	})
	if err != nil {
		return db.Settlement{}, fmt.Errorf("settlement failed: %w", err)
	}

	s.log.Info("trade settled",
		logger.Str("trade_id", arg.TradeID.String()),
		logger.Str("symbol", arg.Symbol),
		logger.Int64("price", arg.Price),
		logger.Int64("quantity", arg.Quantity),
		logger.Int64("total_cash", arg.Price*arg.Quantity),
	)

	return settled, nil
}

type SettleParams struct {
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
