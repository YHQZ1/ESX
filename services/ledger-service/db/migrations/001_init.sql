CREATE TABLE cash_journal (
	id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	trade_id       UUID NOT NULL,
	participant_id UUID NOT NULL,
	entry_type     TEXT NOT NULL,
	amount         BIGINT NOT NULL,
	balance_after  BIGINT NOT NULL,
	created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE securities_journal (
	id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	trade_id       UUID NOT NULL,
	participant_id UUID NOT NULL,
	symbol         TEXT NOT NULL,
	entry_type     TEXT NOT NULL,
	quantity       BIGINT NOT NULL,
	balance_after  BIGINT NOT NULL,
	created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_cash_journal_participant_id ON cash_journal(participant_id);
CREATE INDEX idx_cash_journal_trade_id ON cash_journal(trade_id);
CREATE INDEX idx_securities_journal_participant_id ON securities_journal(participant_id);
CREATE INDEX idx_securities_journal_trade_id ON securities_journal(trade_id);