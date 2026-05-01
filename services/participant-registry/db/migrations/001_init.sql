CREATE TABLE participants (
	id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name       TEXT NOT NULL,
	email      TEXT NOT NULL UNIQUE,
	status     TEXT NOT NULL DEFAULT 'active',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE api_keys (
	id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	participant_id UUID NOT NULL REFERENCES participants(id),
	key_hash       TEXT NOT NULL UNIQUE,
	created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE cash_accounts (
	id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	participant_id UUID NOT NULL UNIQUE REFERENCES participants(id),
	balance        BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
	locked         BIGINT NOT NULL DEFAULT 0 CHECK (locked >= 0),
	currency       TEXT NOT NULL DEFAULT 'INR',
	updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE securities_accounts (
	id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	participant_id UUID NOT NULL REFERENCES participants(id),
	symbol         TEXT NOT NULL,
	quantity       BIGINT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
	locked         BIGINT NOT NULL DEFAULT 0 CHECK (locked >= 0),
	updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (participant_id, symbol)
);

CREATE INDEX idx_api_keys_participant_id ON api_keys(participant_id);
CREATE INDEX idx_securities_accounts_participant_id ON securities_accounts(participant_id);