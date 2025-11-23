package dto

import "goxfer/server/internal/transfer"

type UploadFileCache struct {
	FileSize int64                `json:"fileSize"`
	Plan     *transfer.UploadPlan `json:"plan"`
}

type TransferMeta struct {
	EncMeta   string
	MetaNonce string
}

type TransferDigest struct {
	EncDataChecksum string `json:"dataChecksum"`
	EncMetaChecksum string `json:"metaChecksum"`
}
