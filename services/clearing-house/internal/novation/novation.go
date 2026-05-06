package novation

import (
	"context"
	"fmt"
	"strings"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/clearing-house/internal/db"
	"github.com/google/uuid"
)

type Novator struct {
	db  db.Querier
	log *logger.Logger
}

func New(database db.Querier, log *logger.Logger) *Novator {
	return &Novator{db: database, log: log}
}

func (n *Novator) Clear(ctx context.Context, arg ClearParams) (db.ClearedTrade, error) {
	buyLock, err := n.db.GetLock(ctx, arg.BuyLockID)
	if err != nil {
		return db.ClearedTrade{}, fmt.Errorf("buy lock not found: %w", err)
	}
	if buyLock.Status != "active" {
		return db.ClearedTrade{}, fmt.Errorf("buy lock is not active: %s", buyLock.Status)
	}

	sellLock, err := n.db.GetLock(ctx, arg.SellLockID)
	if err != nil {
		return db.ClearedTrade{}, fmt.Errorf("sell lock not found: %w", err)
	}
	if sellLock.Status != "active" {
		return db.ClearedTrade{}, fmt.Errorf("sell lock is not active: %s", sellLock.Status)
	}

	cleared, err := n.db.CreateClearedTrade(ctx, db.CreateClearedTradeParams{
		TradeID:     arg.TradeID,
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
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			n.log.Info("trade already cleared, skipping duplicate",
				logger.Str("trade_id", arg.TradeID.String()),
			)
			return db.ClearedTrade{}, fmt.Errorf("duplicate: %w", err)
		}
		return db.ClearedTrade{}, fmt.Errorf("failed to create cleared trade: %w", err)
	}

	if err := n.db.UpdateLockStatus(ctx, arg.BuyLockID, "consumed"); err != nil {
		n.log.Error("failed to mark buy lock consumed", err,
			logger.Str("lock_id", arg.BuyLockID.String()),
		)
	}

	if err := n.db.UpdateLockStatus(ctx, arg.SellLockID, "consumed"); err != nil {
		n.log.Error("failed to mark sell lock consumed", err,
			logger.Str("lock_id", arg.SellLockID.String()),
		)
	}

	n.log.Info("trade cleared",
		logger.Str("trade_id", arg.TradeID.String()),
		logger.Str("symbol", arg.Symbol),
		logger.Int64("price", arg.Price),
		logger.Int64("quantity", arg.Quantity),
	)
	return cleared, nil
}

type ClearParams struct {
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
