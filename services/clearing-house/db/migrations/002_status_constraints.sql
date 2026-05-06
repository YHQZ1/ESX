ALTER TABLE cleared_trades 
  ADD CONSTRAINT cleared_trades_status_check 
  CHECK (status IN ('cleared', 'failed'));