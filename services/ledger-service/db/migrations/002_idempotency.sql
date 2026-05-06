CREATE UNIQUE INDEX idx_cash_journal_idempotency 
  ON cash_journal(trade_id, participant_id, entry_type);

CREATE UNIQUE INDEX idx_securities_journal_idempotency 
  ON securities_journal(trade_id, participant_id, symbol, entry_type);