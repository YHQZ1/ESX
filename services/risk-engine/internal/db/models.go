package db

import (
	"time"

	"github.com/google/uuid"
)

type Lock struct {
	ID            uuid.UUID `json:"id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`
	Quantity      int64     `json:"quantity"`
	Price         int64     `json:"price"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CashAccount struct {
	ID            uuid.UUID
	ParticipantID uuid.UUID
	Balance       int64
	Locked        int64
}

type SecuritiesAccount struct {
	ID            uuid.UUID
	ParticipantID uuid.UUID
	Symbol        string
	Quantity      int64
	Locked        int64
}
