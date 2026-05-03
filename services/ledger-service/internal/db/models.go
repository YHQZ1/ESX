package db

import (
	"time"

	"github.com/google/uuid"
)

type CashEntry struct {
	ID            uuid.UUID `json:"id"`
	TradeID       uuid.UUID `json:"trade_id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	EntryType     string    `json:"entry_type"`
	Amount        int64     `json:"amount"`
	BalanceAfter  int64     `json:"balance_after"`
	CreatedAt     time.Time `json:"created_at"`
}

type SecuritiesEntry struct {
	ID            uuid.UUID `json:"id"`
	TradeID       uuid.UUID `json:"trade_id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	Symbol        string    `json:"symbol"`
	EntryType     string    `json:"entry_type"`
	Quantity      int64     `json:"quantity"`
	BalanceAfter  int64     `json:"balance_after"`
	CreatedAt     time.Time `json:"created_at"`
}
