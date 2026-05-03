CREATE TABLE cleared_trades (
	id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	trade_id       UUID NOT NULL UNIQUE,
	symbol         TEXT NOT NULL,
	buyer_id       UUID NOT NULL,
	seller_id      UUID NOT NULL,
	buy_order_id   UUID NOT NULL,
	sell_order_id  UUID NOT NULL,
	buy_lock_id    UUID NOT NULL,
	sell_lock_id   UUID NOT NULL,
	price          BIGINT NOT NULL,
	quantity       BIGINT NOT NULL,
	status         TEXT NOT NULL DEFAULT 'cleared',
	created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_cleared_trades_trade_id ON cleared_trades(trade_id);
CREATE INDEX idx_cleared_trades_buyer_id ON cleared_trades(buyer_id);
CREATE INDEX idx_cleared_trades_seller_id ON cleared_trades(seller_id);