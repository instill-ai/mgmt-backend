BEGIN;
ALTER TABLE public.owner ADD COLUMN "profile_avatar" TEXT DEFAULT NULL;
ALTER TABLE public.owner ADD COLUMN "profile_data" JSONB DEFAULT '{}';
COMMIT;
