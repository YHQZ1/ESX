CREATE TABLE locks (
	id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	participant_id UUID NOT NULL,
	symbol         TEXT NOT NULL,
	side           TEXT NOT NULL,
	quantity       BIGINT NOT NULL,
	price          BIGINT NOT NULL,
	status         TEXT NOT NULL DEFAULT 'active',
	created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_locks_participant_id ON locks(participant_id);
CREATE INDEX idx_locks_status ON locks(status);