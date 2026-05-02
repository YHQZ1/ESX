CREATE TABLE IF NOT EXISTS locks (
    id             TEXT PRIMARY KEY,
    participant_id TEXT        NOT NULL,
    symbol         TEXT        NOT NULL,
    side           TEXT        NOT NULL CHECK (side IN ('BUY', 'SELL')),
    quantity       BIGINT      NOT NULL CHECK (quantity > 0),
    price          BIGINT      NOT NULL CHECK (price >= 0),
    locked_amount  BIGINT      NOT NULL CHECK (locked_amount >= 0),
    status         TEXT        NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'released', 'consumed')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_locks_participant_id ON locks (participant_id);
CREATE INDEX IF NOT EXISTS idx_locks_status ON locks (status);
CREATE INDEX IF NOT EXISTS idx_locks_participant_symbol ON locks (participant_id, symbol);