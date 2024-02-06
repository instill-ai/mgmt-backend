BEGIN;
CREATE TYPE valid_onboarding_status AS ENUM (
  'ONBOARDING_STATUS_UNSPECIFIED',
  'ONBOARDING_STATUS_IN_PROGRESS',
  'ONBOARDING_STATUS_COMPLETED'
);
ALTER TABLE public.owner ADD COLUMN "onboarding_status" valid_onboarding_status DEFAULT 'ONBOARDING_STATUS_UNSPECIFIED';
ALTER TABLE public.owner ADD COLUMN "public_email" VARCHAR(255) DEFAULT '';
ALTER TABLE public.owner ADD COLUMN "display_name" VARCHAR(255) DEFAULT '';
ALTER TABLE public.owner ADD COLUMN "bio" text DEFAULT '';
ALTER TABLE public.owner ADD COLUMN "social_profile_links" jsonb DEFAULT '{}';
ALTER TABLE public.owner ADD COLUMN "company_name" VARCHAR(255) DEFAULT '';

UPDATE public.owner SET display_name = CONCAT(first_name, ' ', last_name);
UPDATE public.owner SET bio = profile_data ->> 'bio';
UPDATE public.owner SET social_profile_links = profile_data - 'bio';
UPDATE public.owner SET company_name = org_name;

COMMIT;
