BEGIN;
ALTER TABLE public.user ALTER COLUMN email drop not null;;
ALTER TABLE public.user RENAME TO owner;
COMMIT;
