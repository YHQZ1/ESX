package locks

import (
	"context"
	"fmt"
	"time"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/risk-engine/internal/db"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// This Lua script executes atomically inside Redis.
// It checks if (balance - locked) >= amount. If yes, it increments the lock and returns 1.
var lockCashLua = redis.NewScript(`
	local balance = tonumber(redis.call('GET', KEYS[1]) or '-1')
	local locked = tonumber(redis.call('GET', KEYS[2]) or '0')
	local amount = tonumber(ARGV[1])

	if balance == -1 then
		return -1 -- Signal to Go that the cache is empty
	end

	if (balance - locked) >= amount then
		redis.call('INCRBY', KEYS[2], amount)
		return 1 -- Success
	end
	return 0 -- Insufficient Funds
`)

type Manager struct {
	db  db.Querier
	rdb *redis.Client
	log *logger.Logger
}

func New(database db.Querier, rdb *redis.Client, log *logger.Logger) *Manager {
	return &Manager{db: database, rdb: rdb, log: log}
}
func (m *Manager) LockCash(ctx context.Context, participantID uuid.UUID, symbol string, price, quantity int64) (uuid.UUID, error) {
	amount := price * quantity
	balanceKey := fmt.Sprintf("participant:%s:cash_balance", participantID)
	lockedKey := fmt.Sprintf("participant:%s:cash_locked", participantID)

	res, err := lockCashLua.Run(ctx, m.rdb, []string{balanceKey, lockedKey}, amount).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("redis lua error: %w", err)
	}

	if res.(int64) == -1 {
		acc, err := m.db.GetCashAccount(ctx, participantID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to fetch account from db: %w", err)
		}

		m.rdb.Set(ctx, balanceKey, acc.Balance, 0)
		m.rdb.Set(ctx, lockedKey, acc.Locked, 0)

		res, err = lockCashLua.Run(ctx, m.rdb, []string{balanceKey, lockedKey}, amount).Result()
		if err != nil {
			return uuid.Nil, fmt.Errorf("redis lua error on retry: %w", err)
		}
	}

	if res.(int64) == 0 {
		return uuid.Nil, fmt.Errorf("insufficient funds")
	}

	// Write the lock to Postgres for audit trail and downstream verification
	lockID := uuid.New()

	go func() {
		bgCtx := context.Background()
		_, err := m.db.CreateLockWithID(bgCtx, db.CreateLockParams{
			ID:            lockID,
			ParticipantID: participantID,
			Symbol:        symbol,
			Side:          "buy",
			Quantity:      quantity,
			Price:         price,
		})
		if err != nil {
			m.log.Error("async lock write failed", err,
				logger.Str("lock_id", lockID.String()),
			)
		}
		if err := m.db.IncrementCashLocked(bgCtx, participantID, amount); err != nil {
			m.log.Error("async cash locked increment failed", err,
				logger.Str("lock_id", lockID.String()),
			)
		}
	}()
	return lockID, nil
}

// Stub out LockShares and Release for now so the compiler doesn't complain
func (m *Manager) LockShares(ctx context.Context, participantID uuid.UUID, symbol string, quantity int64) (uuid.UUID, error) {
	balanceKey := fmt.Sprintf("participant:%s:sec_balance:%s", participantID, symbol)
	lockedKey := fmt.Sprintf("participant:%s:sec_locked:%s", participantID, symbol)

	res, err := lockCashLua.Run(ctx, m.rdb, []string{balanceKey, lockedKey}, quantity).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("redis lua error: %w", err)
	}

	if res.(int64) == -1 {
		acc, err := m.db.GetSecuritiesAccount(ctx, participantID, symbol)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to fetch securities account: %w", err)
		}
		m.rdb.Set(ctx, balanceKey, acc.Quantity, 0)
		m.rdb.Set(ctx, lockedKey, acc.Locked, 0)
		res, err = lockCashLua.Run(ctx, m.rdb, []string{balanceKey, lockedKey}, quantity).Result()
		if err != nil {
			return uuid.Nil, fmt.Errorf("redis lua error on retry: %w", err)
		}
	}

	if res.(int64) == 0 {
		return uuid.Nil, fmt.Errorf("insufficient shares")
	}

	lockID := uuid.New()
	go func() {
		bgCtx := context.Background()
		_, err := m.db.CreateLockWithID(bgCtx, db.CreateLockParams{
			ID:            lockID,
			ParticipantID: participantID,
			Symbol:        symbol,
			Side:          "sell",
			Quantity:      quantity,
			Price:         0,
		})
		if err != nil {
			m.log.Error("async lock write failed", err, logger.Str("lock_id", lockID.String()))
		}
		if err := m.db.IncrementSecuritiesLocked(bgCtx, participantID, symbol, quantity); err != nil {
			m.log.Error("async securities locked increment failed", err, logger.Str("lock_id", lockID.String()))
		}
	}()

	return lockID, nil
}

func (m *Manager) Release(ctx context.Context, lockID uuid.UUID, filledQuantity int64) error {
	var lock db.Lock
	var err error
	for i := range 5 {
		lock, err = m.db.GetLock(ctx, lockID)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
	}
	if err != nil {
		return fmt.Errorf("failed to get lock after retries: %w", err)
	}

	if lock.Status != "active" {
		return fmt.Errorf("lock %s is not active, status: %s", lockID, lock.Status)
	}

	if lock.Side == "buy" {
		unfilledQuantity := lock.Quantity - filledQuantity
		amountToRelease := unfilledQuantity * lock.Price
		if amountToRelease > 0 {
			if err := m.db.DecrementCashLocked(ctx, lock.ParticipantID, amountToRelease); err != nil {
				return fmt.Errorf("failed to decrement cash locked: %w", err)
			}
			lockedKey := fmt.Sprintf("participant:%s:cash_locked", lock.ParticipantID)
			m.rdb.DecrBy(ctx, lockedKey, amountToRelease)
		}
	} else if lock.Side == "sell" {
		unfilledQuantity := lock.Quantity - filledQuantity
		if unfilledQuantity > 0 {
			if err := m.db.DecrementSecuritiesLocked(ctx, lock.ParticipantID, lock.Symbol, unfilledQuantity); err != nil {
				return fmt.Errorf("failed to decrement securities locked: %w", err)
			}
			lockedKey := fmt.Sprintf("participant:%s:sec_locked:%s", lock.ParticipantID, lock.Symbol)
			m.rdb.DecrBy(ctx, lockedKey, unfilledQuantity)
		}
	}

	if err := m.db.UpdateLockStatus(ctx, lockID, "released"); err != nil {
		return fmt.Errorf("failed to update lock status: %w", err)
	}

	return nil
}
