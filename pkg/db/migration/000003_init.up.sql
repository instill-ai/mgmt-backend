BEGIN;
ALTER TABLE public.user ALTER COLUMN email drop not null;;
COMMIT;
