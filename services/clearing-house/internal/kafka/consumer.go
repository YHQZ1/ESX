package consumer

import (
	"context"
	"strings"
	"time"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/clearing-house/internal/novation"
	"github.com/google/uuid"
)

type TradeExecutedEvent struct {
	TradeID     uuid.UUID `json:"trade_id"`
	Symbol      string    `json:"symbol"`
	BuyerID     uuid.UUID `json:"buyer_id"`
	SellerID    uuid.UUID `json:"seller_id"`
	BuyOrderID  uuid.UUID `json:"buy_order_id"`
	SellOrderID uuid.UUID `json:"sell_order_id"`
	BuyLockID   uuid.UUID `json:"buy_lock_id"`
	SellLockID  uuid.UUID `json:"sell_lock_id"`
	Price       int64     `json:"price"`
	Quantity    int64     `json:"quantity"`
	Timestamp   int64     `json:"timestamp"`
}

type TradeClearedEvent struct {
	ClearedID   uuid.UUID `json:"cleared_id"`
	TradeID     uuid.UUID `json:"trade_id"`
	Symbol      string    `json:"symbol"`
	BuyerID     uuid.UUID `json:"buyer_id"`
	SellerID    uuid.UUID `json:"seller_id"`
	BuyOrderID  uuid.UUID `json:"buy_order_id"`
	SellOrderID uuid.UUID `json:"sell_order_id"`
	BuyLockID   uuid.UUID `json:"buy_lock_id"`
	SellLockID  uuid.UUID `json:"sell_lock_id"`
	Price       int64     `json:"price"`
	Quantity    int64     `json:"quantity"`
	Timestamp   int64     `json:"timestamp"`
}

type Handler struct {
	novator  *novation.Novator
	producer *kafka.Producer
	log      *logger.Logger
}

func New(novator *novation.Novator, producer *kafka.Producer, log *logger.Logger) *Handler {
	return &Handler{novator: novator, producer: producer, log: log}
}

func (h *Handler) Handle(ctx context.Context, msg kafka.Message) error {
	event, err := kafka.Decode[TradeExecutedEvent](msg)
	if err != nil {
		h.log.Error("failed to decode trade.executed", err)
		return err
	}

	cleared, err := h.novator.Clear(ctx, novation.ClearParams{
		TradeID:     event.TradeID,
		Symbol:      event.Symbol,
		BuyerID:     event.BuyerID,
		SellerID:    event.SellerID,
		BuyOrderID:  event.BuyOrderID,
		SellOrderID: event.SellOrderID,
		BuyLockID:   event.BuyLockID,
		SellLockID:  event.SellLockID,
		Price:       event.Price,
		Quantity:    event.Quantity,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			h.log.Info("skipping already cleared trade",
				logger.Str("trade_id", event.TradeID.String()),
			)
			return nil // commit the offset, don't retry
		}
		h.log.Error("failed to clear trade", err,
			logger.Str("trade_id", event.TradeID.String()),
		)
		return err
	}

	clearedEvent := TradeClearedEvent{
		ClearedID:   cleared.ID,
		TradeID:     cleared.TradeID,
		Symbol:      cleared.Symbol,
		BuyerID:     cleared.BuyerID,
		SellerID:    cleared.SellerID,
		BuyOrderID:  cleared.BuyOrderID,
		SellOrderID: cleared.SellOrderID,
		BuyLockID:   cleared.BuyLockID,
		SellLockID:  cleared.SellLockID,
		Price:       cleared.Price,
		Quantity:    cleared.Quantity,
		Timestamp:   time.Now().UnixNano(),
	}

	if err := h.producer.Publish(ctx, cleared.Symbol, clearedEvent); err != nil {
		h.log.Error("failed to publish trade.cleared", err, logger.Str("trade_id", event.TradeID.String()))
		return err
	}

	return nil
}
