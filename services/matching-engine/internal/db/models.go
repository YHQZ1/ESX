package db

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID            uuid.UUID `json:"id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`
	OrderType     string    `json:"order_type"`
	TimeInForce   string    `json:"time_in_force"`
	Quantity      int64     `json:"quantity"`
	FilledQty     int64     `json:"filled_qty"`
	Price         int64     `json:"price"`
	LockID        uuid.UUID `json:"lock_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
