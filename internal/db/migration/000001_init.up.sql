BEGIN;
-- user
CREATE TABLE IF NOT EXISTS public.user(
  uid UUID NOT NULL,
  id VARCHAR(255) UNIQUE NOT NULL,
  owner_type VARCHAR(255),
  email VARCHAR(255) UNIQUE NOT NULL,
  plan VARCHAR(255),
  billing_id VARCHAR(255),
  first_name VARCHAR(255),
  last_name VARCHAR(255),
  org_name VARCHAR(255),
  role VARCHAR(255),
  newsletter_subscription BOOL NOT NULL DEFAULT FALSE,
  cookie_token VARCHAR(255),
  create_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  update_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  CONSTRAINT user_pkey PRIMARY KEY (uid)
);
CREATE INDEX user_id_create_time_pagination ON public.user (uid, create_time);
COMMIT;
