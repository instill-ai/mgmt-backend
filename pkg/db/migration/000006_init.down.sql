BEGIN;

DROP INDEX IF EXISTS credit_index_remaining;

DROP TABLE IF EXISTS credit;

DROP FUNCTION IF EXISTS set_update_time();

COMMIT;
