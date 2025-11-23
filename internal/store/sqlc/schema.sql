CREATE TABLE IF NOT EXISTS public.buckets
(
    bucket_id bigint NOT NULL DEFAULT nextval('buckets_bucket_id_seq'::regclass),
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    key character varying(10) COLLATE pg_catalog."default" NOT NULL,
    name character varying(40) COLLATE pg_catalog."default" NOT NULL,
    cred_id text COLLATE pg_catalog."default" NOT NULL,
    record text COLLATE pg_catalog."default" NOT NULL,
    cipher text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT buckets_pkey PRIMARY KEY (bucket_id)
);


CREATE TABLE IF NOT EXISTS public.files
(
    file_id bigint NOT NULL DEFAULT nextval('files_file_id_seq'::regclass),
    buc_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid boolean NOT NULL DEFAULT true,
    data_file text COLLATE pg_catalog."default" NOT NULL,
    file_uuid uuid NOT NULL,
    meta_file text COLLATE pg_catalog."default" NOT NULL,
    digest_file text COLLATE pg_catalog."default" NOT NULL,
    upload_id text COLLATE pg_catalog."default" NOT NULL,
    base_path text COLLATE pg_catalog."default" NOT NULL,
    file_info text COLLATE pg_catalog."default" NOT NULL,
    file_info_nonce text COLLATE pg_catalog."default" NOT NULL,
    data_file_size bigint NOT NULL,
    CONSTRAINT files_pkey PRIMARY KEY (file_id)
);