UPDATE cash_journal SET entry_type = LOWER(entry_type);
UPDATE securities_journal SET entry_type = LOWER(entry_type);

ALTER TABLE cash_journal
  ADD CONSTRAINT cash_journal_entry_type_check
  CHECK (entry_type IN ('debit', 'credit'));

ALTER TABLE securities_journal
  ADD CONSTRAINT securities_journal_entry_type_check
  CHECK (entry_type IN ('debit', 'credit'));