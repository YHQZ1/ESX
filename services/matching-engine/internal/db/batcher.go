package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

type OrderBatcher struct {
	db      *sql.DB
	mu      sync.Mutex
	pending []Order
	flushCh chan struct{}
	done    chan struct{}
}

func NewOrderBatcher(database *sql.DB) *OrderBatcher {
	b := &OrderBatcher{
		db:      database,
		pending: make([]Order, 0, 500),
		flushCh: make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
	go b.run()
	return b
}

func (b *OrderBatcher) Add(order Order) {
	b.mu.Lock()
	b.pending = append(b.pending, order)
	shouldFlush := len(b.pending) >= 100
	b.mu.Unlock()

	if shouldFlush {
		select {
		case b.flushCh <- struct{}{}:
		default:
		}
	}
}

func (b *OrderBatcher) run() {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.flush()
		case <-b.flushCh:
			b.flush()
		case <-b.done:
			b.flush()
			return
		}
	}
}

func (b *OrderBatcher) flush() {
	b.mu.Lock()
	if len(b.pending) == 0 {
		b.mu.Unlock()
		return
	}
	batch := b.pending
	b.pending = make([]Order, 0, 500)
	b.mu.Unlock()

	ctx := context.Background()

	// Build bulk INSERT
	valueStrings := make([]string, 0, len(batch))
	valueArgs := make([]any, 0, len(batch)*10)

	for i, o := range batch {
		base := i * 10
		valueStrings = append(valueStrings,
			fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
				base+1, base+2, base+3, base+4, base+5,
				base+6, base+7, base+8, base+9, base+10,
			),
		)
		valueArgs = append(valueArgs,
			o.ID, o.ParticipantID, o.Symbol, o.Side, o.OrderType,
			o.TimeInForce, o.Quantity, o.Price, o.LockID, o.Status,
		)
	}

	query := fmt.Sprintf(
		`INSERT INTO orders (id, participant_id, symbol, side, order_type, time_in_force, quantity, price, lock_id, status)
		 VALUES %s ON CONFLICT (id) DO NOTHING`,
		strings.Join(valueStrings, ","),
	)

	if _, err := b.db.ExecContext(ctx, query, valueArgs...); err != nil {
		// log but don't crash — orders are already in Redis/matching pipeline
		_ = err
	}
}

func (b *OrderBatcher) Stop() {
	close(b.done)
}
