package locks

import (
	"context"
	"fmt"

	"github.com/YHQZ1/esx/services/risk-engine/internal/db"
	"github.com/google/uuid"
)

type Manager struct {
	db db.Querier
}

func New(database db.Querier) *Manager {
	return &Manager{db: database}
}

func (m *Manager) LockCash(ctx context.Context, participantID uuid.UUID, symbol string, price, quantity int64) (uuid.UUID, error) {
	lock, err := m.db.CreateLock(ctx, db.CreateLockParams{
		ParticipantID: participantID,
		Symbol:        symbol,
		Side:          "BUY",
		Quantity:      quantity,
		Price:         price,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create lock: %w", err)
	}

	if err := m.db.IncrementCashLocked(ctx, participantID, price*quantity); err != nil {
		return uuid.Nil, fmt.Errorf("failed to increment cash locked: %w", err)
	}

	return lock.ID, nil
}

func (m *Manager) LockShares(ctx context.Context, participantID uuid.UUID, symbol string, quantity int64) (uuid.UUID, error) {
	lock, err := m.db.CreateLock(ctx, db.CreateLockParams{
		ParticipantID: participantID,
		Symbol:        symbol,
		Side:          "SELL",
		Quantity:      quantity,
		Price:         0,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create lock: %w", err)
	}

	if err := m.db.IncrementSecuritiesLocked(ctx, participantID, symbol, quantity); err != nil {
		return uuid.Nil, fmt.Errorf("failed to increment securities locked: %w", err)
	}

	return lock.ID, nil
}

func (m *Manager) Release(ctx context.Context, lockID uuid.UUID, filledQuantity int64) error {
	lock, err := m.db.GetLock(ctx, lockID)
	if err != nil {
		return fmt.Errorf("lock not found: %w", err)
	}

	if lock.Status != "active" {
		return fmt.Errorf("lock is not active")
	}

	if lock.Side == "BUY" {
		releaseAmount := lock.Price * (lock.Quantity - filledQuantity)
		if releaseAmount > 0 {
			if err := m.db.DecrementCashLocked(ctx, lock.ParticipantID, releaseAmount); err != nil {
				return fmt.Errorf("failed to decrement cash locked: %w", err)
			}
		}
	} else {
		releaseQuantity := lock.Quantity - filledQuantity
		if releaseQuantity > 0 {
			if err := m.db.DecrementSecuritiesLocked(ctx, lock.ParticipantID, lock.Symbol, releaseQuantity); err != nil {
				return fmt.Errorf("failed to decrement securities locked: %w", err)
			}
		}
	}

	return m.db.UpdateLockStatus(ctx, lockID, "released")
}
