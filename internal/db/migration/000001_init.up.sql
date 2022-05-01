BEGIN;
-- user
CREATE TABLE IF NOT EXISTS public.user(
  id UUID NOT NULL,
  email VARCHAR(255),
  login VARCHAR(255) UNIQUE NOT NULL,
  company_name VARCHAR(255),
  role VARCHAR(255),
  usage_data_collection BOOL NOT NULL DEFAULT FALSE,
  newsletter_subscription BOOL NOT NULL DEFAULT FALSE,
  create_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  update_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  delete_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT user_pkey PRIMARY KEY (id)
);
CREATE INDEX user_id_create_time_pagination ON public.user (id, create_time);
COMMIT;
