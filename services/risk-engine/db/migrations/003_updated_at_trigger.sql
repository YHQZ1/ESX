CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER locks_updated_at
  BEFORE UPDATE ON locks
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();