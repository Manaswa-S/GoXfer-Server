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
    CONSTRAINT files_pkey PRIMARY KEY (file_id),
    CONSTRAINT files_buckets_buc_key_fkey FOREIGN KEY (buc_id)
        REFERENCES public.buckets (bucket_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID
);