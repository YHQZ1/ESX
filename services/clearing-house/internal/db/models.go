package db

import (
	"time"

	"github.com/google/uuid"
)

type ClearedTrade struct {
	ID          uuid.UUID `json:"id"`
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
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Lock struct {
	ID            uuid.UUID
	ParticipantID uuid.UUID
	Symbol        string
	Side          string
	Quantity      int64
	Price         int64
	Status        string
}
