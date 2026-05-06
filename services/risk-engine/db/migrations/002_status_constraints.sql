ALTER TABLE locks 
  ADD CONSTRAINT locks_status_check 
  CHECK (status IN ('active', 'released', 'consumed'));