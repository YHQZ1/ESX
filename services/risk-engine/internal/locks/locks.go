package locks

import (
	"context"
	"fmt"

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
}

func New(database db.Querier, rdb *redis.Client) *Manager {
	return &Manager{db: database, rdb: rdb}
}

func (m *Manager) LockCash(ctx context.Context, participantID uuid.UUID, symbol string, price, quantity int64) (uuid.UUID, error) {
	amount := price * quantity
	balanceKey := fmt.Sprintf("participant:%s:cash_balance", participantID)
	lockedKey := fmt.Sprintf("participant:%s:cash_locked", participantID)

	// Execute the atomic Lua script in Redis
	res, err := lockCashLua.Run(ctx, m.rdb, []string{balanceKey, lockedKey}, amount).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("redis lua error: %w", err)
	}

	// Cache Miss! The user's balance isn't in Redis yet. Fetch from Postgres and seed the cache.
	if res.(int64) == -1 {
		acc, err := m.db.GetCashAccount(ctx, participantID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to fetch account from db: %w", err)
		}

		// Seed Redis with the Postgres truth
		m.rdb.Set(ctx, balanceKey, acc.Balance, 0)
		m.rdb.Set(ctx, lockedKey, acc.Locked, 0)

		// Re-run the Lua script now that the cache is warm
		res, _ = lockCashLua.Run(ctx, m.rdb, []string{balanceKey, lockedKey}, amount).Result()
	}

	if res.(int64) == 0 {
		return uuid.Nil, fmt.Errorf("insufficient funds")
	}

	// We generate a virtual LockID since we aren't creating physical Postgres rows anymore
	return uuid.New(), nil
}

// Stub out LockShares and Release for now so the compiler doesn't complain
func (m *Manager) LockShares(ctx context.Context, participantID uuid.UUID, symbol string, quantity int64) (uuid.UUID, error) {
	return uuid.New(), nil
}
func (m *Manager) Release(ctx context.Context, lockID uuid.UUID, filledQuantity int64) error {
	return nil
}
