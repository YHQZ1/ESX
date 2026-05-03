package db

import (
	"time"

	"github.com/google/uuid"
)

type Settlement struct {
	ID          uuid.UUID `json:"id"`
	TradeID     uuid.UUID `json:"trade_id"`
	ClearedID   uuid.UUID `json:"cleared_id"`
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
	SettledAt   time.Time `json:"settled_at"`
}
