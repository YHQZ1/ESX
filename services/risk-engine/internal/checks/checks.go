package checks

import (
	"context"
	"errors"
	"fmt"

	"github.com/YHQZ1/esx/services/risk-engine/internal/db"
)

var (
	ErrInsufficientCash   = errors.New("insufficient available cash balance")
	ErrInsufficientShares = errors.New("insufficient available share position")
	ErrAccountNotFound    = errors.New("account not found")
	ErrInvalidPrice       = errors.New("price must be greater than zero")
	ErrInvalidQuantity    = errors.New("quantity must be greater than zero")
)

type Checker struct {
	db *db.DB
}

func New(database *db.DB) *Checker {
	return &Checker{db: database}
}

type BuyCheckResult struct {
	AvailableBalance int64
	RequiredAmount   int64
}

type SellCheckResult struct {
	AvailableShares int64
	RequiredShares  int64
}

func (c *Checker) CheckBuy(ctx context.Context, participantID, symbol string, quantity, price int64) (*BuyCheckResult, error) {
	if price <= 0 {
		return nil, ErrInvalidPrice
	}
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	account, err := c.db.GetCashAccount(ctx, participantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("fetch cash account: %w", err)
	}

	required := price * quantity
	available := account.Balance - account.Locked

	if available < required {
		return nil, fmt.Errorf("%w: available %d, required %d", ErrInsufficientCash, available, required)
	}

	return &BuyCheckResult{
		AvailableBalance: available,
		RequiredAmount:   required,
	}, nil
}

func (c *Checker) CheckSell(ctx context.Context, participantID, symbol string, quantity int64) (*SellCheckResult, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	account, err := c.db.GetSecuritiesAccount(ctx, participantID, symbol)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("fetch securities account: %w", err)
	}

	available := account.Quantity - account.Locked

	if available < quantity {
		return nil, fmt.Errorf("%w: available %d, required %d", ErrInsufficientShares, available, quantity)
	}

	return &SellCheckResult{
		AvailableShares: available,
		RequiredShares:  quantity,
	}, nil
}
