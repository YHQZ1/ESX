package circuit

import (
	"context"
	"math"
	"time"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/matching-engine/internal/events"
	"github.com/YHQZ1/esx/services/matching-engine/internal/orderbook"
)

type Breaker struct {
	book      *orderbook.Book
	producer  *kafka.Producer
	threshold float64
	log       *logger.Logger
}

func New(book *orderbook.Book, producer *kafka.Producer, threshold float64, log *logger.Logger) *Breaker {
	return &Breaker{book: book, producer: producer, threshold: threshold, log: log}
}

func (b *Breaker) Check(ctx context.Context, symbol string, newPrice int64) (bool, error) {
	halted, err := b.book.IsHalted(ctx, symbol)
	if err != nil {
		return false, err
	}
	if halted {
		return true, nil
	}

	lastPrice, err := b.book.LastPrice(ctx, symbol)
	if err != nil {
		return false, err
	}
	if lastPrice == 0 {
		return false, nil
	}

	movePct := math.Abs(float64(newPrice-lastPrice) / float64(lastPrice) * 100)
	if movePct >= b.threshold {
		if err := b.book.Halt(ctx, symbol, 60*time.Second); err != nil {
			return false, err
		}

		event := events.CircuitBreakerTriggered{
			Symbol:       symbol,
			LastPrice:    lastPrice,
			CurrentPrice: newPrice,
			MovePct:      movePct,
			Timestamp:    time.Now().UnixNano(),
		}

		if err := b.producer.Publish(ctx, symbol, event); err != nil {
			b.log.Error("failed to publish circuit breaker event", err, logger.Str("symbol", symbol))
		}

		b.log.Info("circuit breaker triggered",
			logger.Str("symbol", symbol),
			logger.Int64("last_price", lastPrice),
			logger.Int64("new_price", newPrice),
		)

		return true, nil
	}

	return false, nil
}

func (b *Breaker) Lift(ctx context.Context, symbol string) error {
	if err := b.book.Lift(ctx, symbol); err != nil {
		return err
	}

	event := events.CircuitBreakerLifted{
		Symbol:    symbol,
		Timestamp: time.Now().UnixNano(),
	}

	if err := b.producer.Publish(ctx, symbol, event); err != nil {
		b.log.Error("failed to publish circuit breaker lifted event", err, logger.Str("symbol", symbol))
	}

	b.log.Info("circuit breaker lifted", logger.Str("symbol", symbol))
	return nil
}
