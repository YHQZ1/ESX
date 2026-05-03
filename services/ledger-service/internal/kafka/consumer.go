package consumer

import (
	"context"

	"github.com/YHQZ1/esx/packages/kafka"
	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/ledger-service/internal/journal"
	"github.com/google/uuid"
)

type TradeSettledEvent struct {
	SettlementID uuid.UUID `json:"settlement_id"`
	TradeID      uuid.UUID `json:"trade_id"`
	Symbol       string    `json:"symbol"`
	BuyerID      uuid.UUID `json:"buyer_id"`
	SellerID     uuid.UUID `json:"seller_id"`
	Price        int64     `json:"price"`
	Quantity     int64     `json:"quantity"`
	Timestamp    int64     `json:"timestamp"`
}

type Handler struct {
	journal *journal.Journal
	log     *logger.Logger
}

func New(j *journal.Journal, log *logger.Logger) *Handler {
	return &Handler{journal: j, log: log}
}

func (h *Handler) Handle(ctx context.Context, msg kafka.Message) error {
	event, err := kafka.Decode[TradeSettledEvent](msg)
	if err != nil {
		h.log.Error("failed to decode trade.settled", err)
		return err
	}

	if err := h.journal.Record(ctx, journal.RecordParams{
		TradeID:  event.TradeID,
		Symbol:   event.Symbol,
		BuyerID:  event.BuyerID,
		SellerID: event.SellerID,
		Price:    event.Price,
		Quantity: event.Quantity,
	}); err != nil {
		h.log.Error("failed to record journal entries", err,
			logger.Str("trade_id", event.TradeID.String()),
		)
		return err
	}

	return nil
}
