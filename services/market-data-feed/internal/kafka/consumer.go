package consumer

import (
	"context"
	"time"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/market-data-feed/internal/channels"
	"github.com/YHQZ1/esx/services/market-data-feed/internal/ws"
	"github.com/google/uuid"
)

type TradeExecutedEvent struct {
	TradeID     uuid.UUID `json:"trade_id"`
	Symbol      string    `json:"symbol"`
	BuyerID     uuid.UUID `json:"buyer_id"`
	SellerID    uuid.UUID `json:"seller_id"`
	BuyOrderID  uuid.UUID `json:"buy_order_id"`
	SellOrderID uuid.UUID `json:"sell_order_id"`
	Price       int64     `json:"price"`
	Quantity    int64     `json:"quantity"`
	Timestamp   int64     `json:"timestamp"`
}

type TickerData struct {
	Symbol    string `json:"symbol"`
	LastPrice int64  `json:"last_price"`
	Quantity  int64  `json:"quantity"`
	Timestamp int64  `json:"timestamp"`
}

type TradeData struct {
	TradeID   string `json:"trade_id"`
	Symbol    string `json:"symbol"`
	Price     int64  `json:"price"`
	Quantity  int64  `json:"quantity"`
	Timestamp int64  `json:"timestamp"`
}

type Handler struct {
	hub *ws.Hub
	log *logger.Logger
}

func New(hub *ws.Hub, log *logger.Logger) *Handler {
	return &Handler{hub: hub, log: log}
}

func (h *Handler) HandleTradeExecuted(ctx context.Context, msg kafka.Message) error {
	event, err := kafka.Decode[TradeExecutedEvent](msg)
	if err != nil {
		h.log.Error("failed to decode trade.executed", err)
		return err
	}

	h.hub.Broadcast(channels.Trades(event.Symbol), TradeData{
		TradeID:   event.TradeID.String(),
		Symbol:    event.Symbol,
		Price:     event.Price,
		Quantity:  event.Quantity,
		Timestamp: time.Now().UnixNano(),
	})

	h.hub.Broadcast(channels.Ticker(event.Symbol), TickerData{
		Symbol:    event.Symbol,
		LastPrice: event.Price,
		Quantity:  event.Quantity,
		Timestamp: time.Now().UnixNano(),
	})

	h.log.Debug("broadcast trade",
		logger.Str("symbol", event.Symbol),
		logger.Int64("price", event.Price),
		logger.Int64("quantity", event.Quantity),
	)

	return nil
}
