package orderbook

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

type Entry struct {
	OrderID       uuid.UUID `json:"order_id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	Symbol        string    `json:"symbol"`
	Side          Side      `json:"side"`
	Price         int64     `json:"price"`
	Quantity      int64     `json:"quantity"`
	LockID        uuid.UUID `json:"lock_id"`
	Timestamp     int64     `json:"timestamp"`
}

type Book struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Book {
	return &Book{rdb: rdb}
}

func (b *Book) Add(ctx context.Context, e Entry) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	key := b.key(e.Symbol, e.Side)

	var score float64
	if e.Side == SideBuy {
		score = float64(-e.Price)*1e10 + float64(e.Timestamp)
	} else {
		score = float64(e.Price)*1e10 + float64(e.Timestamp)
	}

	return b.rdb.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: string(data),
	}).Err()
}

func (b *Book) BestAsk(ctx context.Context, symbol string) (*Entry, error) {
	return b.peek(ctx, symbol, SideSell)
}

func (b *Book) BestBid(ctx context.Context, symbol string) (*Entry, error) {
	return b.peek(ctx, symbol, SideBuy)
}

func (b *Book) peek(ctx context.Context, symbol string, side Side) (*Entry, error) {
	key := b.key(symbol, side)
	results, err := b.rdb.ZRangeWithScores(ctx, key, 0, 0).Result()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	var e Entry
	if err := json.Unmarshal([]byte(results[0].Member.(string)), &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (b *Book) Remove(ctx context.Context, symbol string, side Side, e Entry) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return b.rdb.ZRem(ctx, b.key(symbol, side), string(data)).Err()
}

func (b *Book) Update(ctx context.Context, symbol string, side Side, old Entry, newQty int64) error {
	if err := b.Remove(ctx, symbol, side, old); err != nil {
		return err
	}
	old.Quantity = newQty
	return b.Add(ctx, old)
}

func (b *Book) Depth(ctx context.Context, symbol string, side Side, limit int64) ([]Entry, error) {
	key := b.key(symbol, side)
	results, err := b.rdb.ZRange(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(results))
	for _, r := range results {
		var e Entry
		if err := json.Unmarshal([]byte(r), &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (b *Book) IsHalted(ctx context.Context, symbol string) (bool, error) {
	val, err := b.rdb.Get(ctx, b.haltKey(symbol)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

func (b *Book) Halt(ctx context.Context, symbol string, duration time.Duration) error {
	return b.rdb.Set(ctx, b.haltKey(symbol), "1", duration).Err()
}

func (b *Book) Lift(ctx context.Context, symbol string) error {
	return b.rdb.Del(ctx, b.haltKey(symbol)).Err()
}

func (b *Book) LastPrice(ctx context.Context, symbol string) (int64, error) {
	val, err := b.rdb.Get(ctx, b.priceKey(symbol)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

func (b *Book) SetLastPrice(ctx context.Context, symbol string, price int64) error {
	return b.rdb.Set(ctx, b.priceKey(symbol), fmt.Sprintf("%d", price), 0).Err()
}

func (b *Book) key(symbol string, side Side) string {
	return fmt.Sprintf("ob:%s:%s", symbol, side)
}

func (b *Book) haltKey(symbol string) string {
	return fmt.Sprintf("cb:halt:%s", symbol)
}

func (b *Book) priceKey(symbol string) string {
	return fmt.Sprintf("ob:price:%s", symbol)
}
