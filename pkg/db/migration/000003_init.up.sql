BEGIN;
ALTER TABLE public.user RENAME TO owner;

ALTER TABLE public.owner DROP CONSTRAINT user_email_key;
ALTER TABLE public.owner ALTER COLUMN email drop NOT NULL;
CREATE UNIQUE INDEX email_unique ON public.owner (email) WHERE (owner_type = 'user');

COMMIT;
