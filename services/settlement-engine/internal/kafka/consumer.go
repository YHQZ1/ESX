package consumer

import (
	"context"
	"time"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/settlement-engine/internal/settlement"
	"github.com/google/uuid"
)

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

type TradeSettledEvent struct {
	SettlementID uuid.UUID `json:"settlement_id"`
	TradeID      uuid.UUID `json:"trade_id"`
	ClearedID    uuid.UUID `json:"cleared_id"`
	Symbol       string    `json:"symbol"`
	BuyerID      uuid.UUID `json:"buyer_id"`
	SellerID     uuid.UUID `json:"seller_id"`
	BuyOrderID   uuid.UUID `json:"buy_order_id"`
	SellOrderID  uuid.UUID `json:"sell_order_id"`
	BuyLockID    uuid.UUID `json:"buy_lock_id"`
	SellLockID   uuid.UUID `json:"sell_lock_id"`
	Price        int64     `json:"price"`
	Quantity     int64     `json:"quantity"`
	Timestamp    int64     `json:"timestamp"`
}

type Handler struct {
	settler  *settlement.Settler
	producer *kafka.Producer
	log      *logger.Logger
}

func New(settler *settlement.Settler, producer *kafka.Producer, log *logger.Logger) *Handler {
	return &Handler{settler: settler, producer: producer, log: log}
}

func (h *Handler) Handle(ctx context.Context, msg kafka.Message) error {
	event, err := kafka.Decode[TradeClearedEvent](msg)
	if err != nil {
		h.log.Error("failed to decode trade.cleared", err)
		return err
	}

	settled, err := h.settler.Settle(ctx, settlement.SettleParams{
		TradeID:     event.TradeID,
		ClearedID:   event.ClearedID,
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
		h.log.Error("failed to settle trade", err, logger.Str("trade_id", event.TradeID.String()))
		return err
	}

	settledEvent := TradeSettledEvent{
		SettlementID: settled.ID,
		TradeID:      settled.TradeID,
		ClearedID:    settled.ClearedID,
		Symbol:       settled.Symbol,
		BuyerID:      settled.BuyerID,
		SellerID:     settled.SellerID,
		BuyOrderID:   settled.BuyOrderID,
		SellOrderID:  settled.SellOrderID,
		BuyLockID:    settled.BuyLockID,
		SellLockID:   settled.SellLockID,
		Price:        settled.Price,
		Quantity:     settled.Quantity,
		Timestamp:    time.Now().UnixNano(),
	}

	if err := h.producer.Publish(ctx, settled.Symbol, settledEvent); err != nil {
		h.log.Error("failed to publish trade.settled", err, logger.Str("trade_id", event.TradeID.String()))
		return err
	}

	return nil
}
