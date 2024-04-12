BEGIN;

CREATE OR REPLACE FUNCTION set_update_time() RETURNS TRIGGER AS $$
BEGIN
  NEW.update_time = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS credit (
  uid         UUID          DEFAULT gen_random_uuid() PRIMARY KEY,
  owner_uid   UUID                                    NOT NULL,
  amount      DECIMAL(16,8)                           NOT NULL,
  expire_time TIMESTAMPTZ,
  create_time TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP NOT NULL,
  update_time TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS credit_index_remaining
  ON credit (owner_uid, (amount > 0), expire_time DESC)
  WHERE amount > 0;

CREATE TRIGGER tg_update_time_credit
  BEFORE UPDATE ON credit
  FOR EACH ROW EXECUTE PROCEDURE set_update_time();

COMMIT;
