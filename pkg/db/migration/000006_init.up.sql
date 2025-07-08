BEGIN;
CREATE TABLE public.fga_migrations (
    "store_id" VARCHAR(255) NOT NULL,
    "authorization_model_id" VARCHAR(255) NOT NULL,
    "md5_hash" VARCHAR(255) NOT NULL,
    PRIMARY KEY ("store_id", "authorization_model_id")
);
COMMIT;
