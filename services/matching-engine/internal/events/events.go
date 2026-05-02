package events

import "github.com/google/uuid"

type TradeExecuted struct {
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

type OrderPartiallyFilled struct {
	OrderID       uuid.UUID `json:"order_id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`
	FilledQty     int64     `json:"filled_qty"`
	RemainingQty  int64     `json:"remaining_qty"`
	Price         int64     `json:"price"`
}

type CircuitBreakerTriggered struct {
	Symbol       string  `json:"symbol"`
	LastPrice    int64   `json:"last_price"`
	CurrentPrice int64   `json:"current_price"`
	MovePct      float64 `json:"move_pct"`
	Timestamp    int64   `json:"timestamp"`
}

type CircuitBreakerLifted struct {
	Symbol    string `json:"symbol"`
	Timestamp int64  `json:"timestamp"`
}
