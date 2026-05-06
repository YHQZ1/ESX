ALTER TABLE settlements 
  ADD CONSTRAINT settlements_status_check 
  CHECK (status IN ('settled', 'failed'));