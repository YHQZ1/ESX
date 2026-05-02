package checks

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/YHQZ1/esx/services/risk-engine/internal/db"
	"github.com/google/uuid"
)

type Checker struct {
	db db.Querier
}

func New(database db.Querier) *Checker {
	return &Checker{db: database}
}

func (c *Checker) CheckBuy(ctx context.Context, participantID uuid.UUID, price, quantity int64) error {
	account, err := c.db.GetCashAccount(ctx, participantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("cash account not found")
		}
		return fmt.Errorf("failed to get cash account: %w", err)
	}

	required := price * quantity
	available := account.Balance - account.Locked

	if available < required {
		return fmt.Errorf("insufficient funds: required %d, available %d", required, available)
	}

	return nil
}

func (c *Checker) CheckSell(ctx context.Context, participantID uuid.UUID, symbol string, quantity int64) error {
	account, err := c.db.GetSecuritiesAccount(ctx, participantID, symbol)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("securities account not found for symbol %s", symbol)
		}
		return fmt.Errorf("failed to get securities account: %w", err)
	}

	available := account.Quantity - account.Locked

	if available < quantity {
		return fmt.Errorf("insufficient shares: required %d, available %d", quantity, available)
	}

	return nil
}
