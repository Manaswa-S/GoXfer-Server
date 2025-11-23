
-- name: InsertBucket :one
INSERT INTO buckets (key, name, cred_id, record, cipher)
VALUES ($1, $2, $3, $4, $5)
RETURNING bucket_id;

-- name: GetBucket :one
SELECT 
    *
FROM buckets
WHERE buckets.key = $1;

-- name: GetBucketID :one
SELECT 
    buckets.bucket_id
FROM buckets
WHERE buckets.key = $1;

-- name: InsertNewFile :exec
INSERT INTO files (buc_id, upload_id, valid, data_file, file_uuid, meta_file, digest_file, base_path, file_info, file_info_nonce, data_file_size)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: GetFiles :many
SELECT
    files.created_at,
    files.file_uuid,
    files.file_info,
    files.file_info_nonce
FROM files
WHERE files.buc_id = $1 
AND files.valid = true;

-- name: GetFileLoc :one
SELECT
    files.file_id,
    files.base_path,
    files.data_file,
    files.meta_file,
    files.digest_file
FROM files
WHERE files.file_uuid = $1
AND files.valid = true
AND files.buc_id = $2;

-- name: GetFileMeta :one
SELECT
    files.data_file_size,
    files.created_at
FROM files
WHERE files.file_id = $1;

-- name: GetFileID :one
SELECT
    files.file_id
FROM files
WHERE files.file_uuid = $1
AND files.valid = true
AND files.buc_id = $2;

-- name: UpdateFileValidity :exec
UPDATE files
SET 
    valid = $2
WHERE file_id = $1;