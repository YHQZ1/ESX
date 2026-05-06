package matching

import (
	"context"
	"strings"
	"time"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	pbrisk "github.com/YHQZ1/esx/packages/proto/risk"
	"github.com/YHQZ1/esx/services/matching-engine/internal/circuit"
	"github.com/YHQZ1/esx/services/matching-engine/internal/db"
	"github.com/YHQZ1/esx/services/matching-engine/internal/events"
	"github.com/YHQZ1/esx/services/matching-engine/internal/orderbook"
	"github.com/google/uuid"
)

type Engine struct {
	book       *orderbook.Book
	queries    db.Querier
	breaker    *circuit.Breaker
	producer   *kafka.Producer
	log        *logger.Logger
	riskClient pbrisk.RiskServiceClient
}

func New(book *orderbook.Book, queries db.Querier, breaker *circuit.Breaker, producer *kafka.Producer, log *logger.Logger, riskClient pbrisk.RiskServiceClient) *Engine {
	return &Engine{book: book, queries: queries, breaker: breaker, producer: producer, log: log, riskClient: riskClient}
}

func (e *Engine) Submit(ctx context.Context, order db.Order) (string, error) {
	halted, err := e.book.IsHalted(ctx, order.Symbol)
	if err != nil {
		return "", err
	}
	if halted {
		return "halted", nil
	}

	if strings.ToUpper(order.Side) == "BUY" {
		return e.matchBuy(ctx, order)
	}
	return e.matchSell(ctx, order)
}

func (e *Engine) matchBuy(ctx context.Context, order db.Order) (string, error) {
	remaining := order.Quantity
	for remaining > 0 {
		ask, err := e.book.BestAsk(ctx, order.Symbol)
		if err != nil {
			return "", err
		}
		if ask == nil {
			break
		}
		if order.OrderType == "LIMIT" && order.Price < ask.Price {
			break
		}

		halted, err := e.breaker.Check(ctx, order.Symbol, ask.Price)
		if err != nil {
			return "", err
		}
		if halted {
			return "halted", nil
		}

		fillQty := min(remaining, ask.Quantity)
		tradePrice := ask.Price

		if err := e.executeTrade(ctx, order, *ask, fillQty, tradePrice); err != nil {
			return "", err
		}

		remaining -= fillQty

		if ask.Quantity > fillQty {
			if err := e.book.Update(ctx, order.Symbol, orderbook.SideSell, *ask, ask.Quantity-fillQty); err != nil {
				return "", err
			}
			go e.queries.UpdateOrderStatus(context.Background(), ask.OrderID, "partial", ask.Quantity-(ask.Quantity-fillQty))
		} else {
			if err := e.book.Remove(ctx, order.Symbol, orderbook.SideSell, *ask); err != nil {
				return "", err
			}
			go e.queries.UpdateOrderStatus(context.Background(), ask.OrderID, "filled", ask.Quantity)
		}

		if err := e.book.SetLastPrice(ctx, order.Symbol, tradePrice); err != nil {
			return "", err
		}
	}

	filledQty := order.Quantity - remaining
	if remaining == 0 {
		go e.queries.UpdateOrderStatus(context.Background(), order.ID, "filled", filledQty)
		return "filled", nil
	}

	if filledQty > 0 {
		go e.queries.UpdateOrderStatus(context.Background(), order.ID, "partial", filledQty)
		if order.TimeInForce == "IOC" {
			go e.queries.UpdateOrderStatus(context.Background(), order.ID, "cancelled", filledQty)
			return "partial_cancelled", nil
		}
		entry := orderbook.Entry{
			OrderID:       order.ID,
			ParticipantID: order.ParticipantID,
			Symbol:        order.Symbol,
			Side:          orderbook.SideBuy,
			Price:         order.Price,
			Quantity:      remaining,
			LockID:        order.LockID,
			Timestamp:     time.Now().UnixNano(),
		}
		if err := e.book.Add(ctx, entry); err != nil {
			return "", err
		}
		return "partial", nil
	}

	if order.TimeInForce == "IOC" {
		go e.queries.UpdateOrderStatus(context.Background(), order.ID, "cancelled", 0)
		return "cancelled", nil
	}

	entry := orderbook.Entry{
		OrderID:       order.ID,
		ParticipantID: order.ParticipantID,
		Symbol:        order.Symbol,
		Side:          orderbook.SideBuy,
		Price:         order.Price,
		Quantity:      remaining,
		LockID:        order.LockID,
		Timestamp:     time.Now().UnixNano(),
	}
	if err := e.book.Add(ctx, entry); err != nil {
		return "", err
	}
	return "open", nil
}

func (e *Engine) matchSell(ctx context.Context, order db.Order) (string, error) {
	remaining := order.Quantity
	for remaining > 0 {
		bid, err := e.book.BestBid(ctx, order.Symbol)
		if err != nil {
			return "", err
		}
		if bid == nil {
			break
		}
		if order.OrderType == "LIMIT" && order.Price > bid.Price {
			break
		}

		halted, err := e.breaker.Check(ctx, order.Symbol, bid.Price)
		if err != nil {
			return "", err
		}
		if halted {
			return "halted", nil
		}

		fillQty := min(remaining, bid.Quantity)
		tradePrice := bid.Price

		if err := e.executeTrade(ctx, order, *bid, fillQty, tradePrice); err != nil {
			return "", err
		}

		remaining -= fillQty

		if bid.Quantity > fillQty {
			if err := e.book.Update(ctx, order.Symbol, orderbook.SideBuy, *bid, bid.Quantity-fillQty); err != nil {
				return "", err
			}
			go e.queries.UpdateOrderStatus(context.Background(), bid.OrderID, "partial", bid.Quantity-(bid.Quantity-fillQty))
		} else {
			if err := e.book.Remove(ctx, order.Symbol, orderbook.SideBuy, *bid); err != nil {
				return "", err
			}
			go e.queries.UpdateOrderStatus(context.Background(), bid.OrderID, "filled", bid.Quantity)
		}

		if err := e.book.SetLastPrice(ctx, order.Symbol, tradePrice); err != nil {
			return "", err
		}
	}

	filledQty := order.Quantity - remaining
	if remaining == 0 {
		go e.queries.UpdateOrderStatus(context.Background(), order.ID, "filled", filledQty)
		return "filled", nil
	}

	if filledQty > 0 {
		go e.queries.UpdateOrderStatus(context.Background(), order.ID, "partial", filledQty)
		if order.TimeInForce == "IOC" {
			go e.queries.UpdateOrderStatus(context.Background(), order.ID, "cancelled", filledQty)
			return "partial_cancelled", nil
		}
		entry := orderbook.Entry{
			OrderID:       order.ID,
			ParticipantID: order.ParticipantID,
			Symbol:        order.Symbol,
			Side:          orderbook.SideSell,
			Price:         order.Price,
			Quantity:      remaining,
			LockID:        order.LockID,
			Timestamp:     time.Now().UnixNano(),
		}
		if err := e.book.Add(ctx, entry); err != nil {
			return "", err
		}
		return "partial", nil
	}

	if order.TimeInForce == "IOC" {
		go e.queries.UpdateOrderStatus(context.Background(), order.ID, "cancelled", 0)
		return "cancelled", nil
	}

	entry := orderbook.Entry{
		OrderID:       order.ID,
		ParticipantID: order.ParticipantID,
		Symbol:        order.Symbol,
		Side:          orderbook.SideSell,
		Price:         order.Price,
		Quantity:      remaining,
		LockID:        order.LockID,
		Timestamp:     time.Now().UnixNano(),
	}
	if err := e.book.Add(ctx, entry); err != nil {
		return "", err
	}
	return "open", nil
}

func (e *Engine) executeTrade(ctx context.Context, incoming db.Order, resting orderbook.Entry, quantity, price int64) error {
	var buyerID, sellerID uuid.UUID
	var buyOrderID, sellOrderID uuid.UUID
	var buyLockID, sellLockID uuid.UUID

	if strings.ToUpper(incoming.Side) == "BUY" {
		buyerID = incoming.ParticipantID
		sellerID = resting.ParticipantID
		buyOrderID = incoming.ID
		sellOrderID = resting.OrderID
		buyLockID = incoming.LockID
		sellLockID = resting.LockID
	} else {
		sellerID = incoming.ParticipantID
		buyerID = resting.ParticipantID
		sellOrderID = incoming.ID
		buyOrderID = resting.OrderID
		sellLockID = incoming.LockID
		buyLockID = resting.LockID
	}

	event := events.TradeExecuted{
		TradeID:     uuid.New(),
		Symbol:      incoming.Symbol,
		BuyerID:     buyerID,
		SellerID:    sellerID,
		BuyOrderID:  buyOrderID,
		SellOrderID: sellOrderID,
		BuyLockID:   buyLockID,
		SellLockID:  sellLockID,
		Price:       price,
		Quantity:    quantity,
		Timestamp:   time.Now().UnixNano(),
	}

	// Fire-and-forget Kafka publishing to unblock the matcher loop
	go func() {
		if err := e.producer.Publish(context.Background(), incoming.Symbol, event); err != nil {
			e.log.Error("failed to publish trade.executed", err, logger.Str("symbol", incoming.Symbol))
		}
	}()

	e.log.Info("trade executed", logger.Str("symbol", incoming.Symbol), logger.Int64("price", price), logger.Int64("quantity", quantity), logger.Str("buyer", buyerID.String()), logger.Str("seller", sellerID.String()))
	return nil
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func (e *Engine) ReleaseLock(ctx context.Context, lockID uuid.UUID, filledQty int64) error {
	_, err := e.riskClient.ReleaseLock(ctx, &pbrisk.ReleaseLockRequest{
		LockId:         lockID.String(),
		FilledQuantity: filledQty,
	})
	if err != nil {
		e.log.Error("failed to release lock", err, logger.Str("lock_id", lockID.String()))
		return err
	}
	return nil
}
