package db

import (
	"time"

	"github.com/google/uuid"
)

type Participant struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type APIKey struct {
	ID            uuid.UUID `json:"id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	KeyHash       string    `json:"key_hash"`
	CreatedAt     time.Time `json:"created_at"`
}

type CashAccount struct {
	ID            uuid.UUID `json:"id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	Balance       int64     `json:"balance"`
	Locked        int64     `json:"locked"`
	Currency      string    `json:"currency"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SecuritiesAccount struct {
	ID            uuid.UUID `json:"id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	Symbol        string    `json:"symbol"`
	Quantity      int64     `json:"quantity"`
	Locked        int64     `json:"locked"`
	UpdatedAt     time.Time `json:"updated_at"`
}
