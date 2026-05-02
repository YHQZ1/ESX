CREATE TABLE orders (
	id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	participant_id UUID NOT NULL,
	symbol         TEXT NOT NULL,
	side           TEXT NOT NULL,
	type           TEXT NOT NULL,
	time_in_force  TEXT NOT NULL DEFAULT 'GTC',
	quantity       BIGINT NOT NULL,
	filled_qty     BIGINT NOT NULL DEFAULT 0,
	price          BIGINT NOT NULL DEFAULT 0,
	lock_id        UUID NOT NULL,
	status         TEXT NOT NULL DEFAULT 'open',
	created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_participant_id ON orders(participant_id);
CREATE INDEX idx_orders_symbol ON orders(symbol);
CREATE INDEX idx_orders_status ON orders(status);