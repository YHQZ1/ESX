package locks

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/YHQZ1/esx/services/risk-engine/internal/checks"
	"github.com/YHQZ1/esx/services/risk-engine/internal/db"
)

type Manager struct {
	db      *db.DB
	checker *checks.Checker
}

func New(database *db.DB, checker *checks.Checker) *Manager {
	return &Manager{db: database, checker: checker}
}

type LockResult struct {
	LockID       string
	LockedAmount int64
}

func (m *Manager) LockForBuy(ctx context.Context, participantID, symbol string, quantity, price int64) (*LockResult, error) {
	result, err := m.checker.CheckBuy(ctx, participantID, symbol, quantity, price)
	if err != nil {
		return nil, err
	}

	lockID := uuid.New().String()
	err = m.db.CreateLockAndIncrementCashLocked(ctx, lockID, participantID, symbol, quantity, price, result.RequiredAmount)
	if err != nil {
		return nil, fmt.Errorf("create cash lock: %w", err)
	}

	return &LockResult{
		LockID:       lockID,
		LockedAmount: result.RequiredAmount,
	}, nil
}

func (m *Manager) LockForSell(ctx context.Context, participantID, symbol string, quantity, price int64) (*LockResult, error) {
	_, err := m.checker.CheckSell(ctx, participantID, symbol, quantity)
	if err != nil {
		return nil, err
	}

	lockID := uuid.New().String()
	err = m.db.CreateLockAndIncrementSharesLocked(ctx, lockID, participantID, symbol, quantity, price)
	if err != nil {
		return nil, fmt.Errorf("create shares lock: %w", err)
	}

	return &LockResult{
		LockID:       lockID,
		LockedAmount: quantity,
	}, nil
}

func (m *Manager) Release(ctx context.Context, lockID string) error {
	lock, err := m.db.GetLock(ctx, lockID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return fmt.Errorf("lock %s not found", lockID)
		}
		return fmt.Errorf("fetch lock: %w", err)
	}

	if lock.Side == "BUY" {
		return m.db.ReleaseCashLock(ctx, lockID, lock.LockedAmount)
	}
	return m.db.ReleaseSharesLock(ctx, lockID, lock.LockedAmount)
}

func (m *Manager) Consume(ctx context.Context, lockID string) error {
	_, err := m.db.GetLock(ctx, lockID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return fmt.Errorf("lock %s not found", lockID)
		}
		return fmt.Errorf("fetch lock: %w", err)
	}
	return m.db.ConsumeLock(ctx, lockID)
}
