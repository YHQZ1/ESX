CREATE TABLE settlements (
	id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	trade_id       UUID NOT NULL UNIQUE,
	cleared_id     UUID NOT NULL,
	symbol         TEXT NOT NULL,
	buyer_id       UUID NOT NULL,
	seller_id      UUID NOT NULL,
	buy_order_id   UUID NOT NULL,
	sell_order_id  UUID NOT NULL,
	buy_lock_id    UUID NOT NULL,
	sell_lock_id   UUID NOT NULL,
	price          BIGINT NOT NULL,
	quantity       BIGINT NOT NULL,
	status         TEXT NOT NULL DEFAULT 'settled',
	settled_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_settlements_trade_id ON settlements(trade_id);
CREATE INDEX idx_settlements_buyer_id ON settlements(buyer_id);
CREATE INDEX idx_settlements_seller_id ON settlements(seller_id);