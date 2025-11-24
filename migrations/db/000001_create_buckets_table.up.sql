CREATE TABLE IF NOT EXISTS public.buckets
(
    bucket_id bigint NOT NULL DEFAULT nextval('buckets_bucket_id_seq'::regclass),
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    key character varying(12) COLLATE pg_catalog."default" NOT NULL,
    name character varying(40) COLLATE pg_catalog."default" NOT NULL,
    cred_id text COLLATE pg_catalog."default" NOT NULL,
    record text COLLATE pg_catalog."default" NOT NULL,
    cipher text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT buckets_pkey PRIMARY KEY (bucket_id)
)
