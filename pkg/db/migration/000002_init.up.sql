BEGIN;
ALTER TABLE public.user ADD COLUMN "password_hash" VARCHAR(255) DEFAULT '' NOT NULL;
ALTER TABLE public.user ADD COLUMN "password_update_time" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL;

CREATE TYPE valid_api_token_state AS ENUM (
  'STATE_UNSPECIFIED',
  'STATE_INACTIVE',
  'STATE_ACTIVE'
);

CREATE TABLE IF NOT EXISTS public.token(
  uid UUID NOT NULL,
  id VARCHAR(255) NOT NULL,
  owner VARCHAR(255) NOT NULL,
  access_token VARCHAR(255) UNIQUE,
  state valid_api_token_state DEFAULT 'STATE_UNSPECIFIED' NOT NULL,
  token_type VARCHAR(255),
  last_use_time TIMESTAMPTZ NULL,
  expire_time TIMESTAMPTZ NULL,
  create_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  update_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  CONSTRAINT api_token_pkey PRIMARY KEY (uid)
);
CREATE UNIQUE INDEX unique_owner_id_delete_time ON public.token (owner, id);

UPDATE public.user SET id = 'admin' WHERE id = 'instill-ai';

COMMIT;
