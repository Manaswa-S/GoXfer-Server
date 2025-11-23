package dto

import "time"

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
type CreateBucketS1Req struct {
	S1Req string `json:"s1Req"` // base64(.Serialize)
}
type CreateBucketS1Resp struct {
	S1Resp   string `json:"s1Resp"` // base64(.Serialize)
	ReqID    string `json:"reqID"`
	ServerID string `json:"serverID"` // base64()
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
type CreateBucketS2Req struct {
	BucName string `json:"bucName"` // string, bucket name
	S2Req   string `json:"s2Req"`   // base64(record.Serialize()), opaque step 2 record
	ReqID   string `json:"reqID"`   // string, request ID
	Cipher  string `json:"cipher"`  // base64(bucketCipher), cipher for bucket
}
type CreateBucketS2Resp struct {
	BucketKey string `json:"bucketKey"`
	Name      string `json:"name"`
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
type OpenBucketS1Req struct {
	BucketKey string `json:"bucketKey"`
	KE1       string `json:"ke1"`
}
type OpenBucketS1Resp struct {
	KE2      string `json:"ke2"`
	ClientID string `json:"clientID"`
	LoginID  string `json:"loginID"`
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
type OpenBucketS2Req struct {
	KE3     string `json:"ke3"`
	LoginID string `json:"loginID"`
}
type OpenBucketS2Resp struct {
	SessionID  string `json:"sessionID"`
	SessionTTL int64  `json:"sessionTTL"`
	Cipher     string `json:"cipher"` // base64(bucketCipher), cipher for bucket
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
type GetOpaqueConfigs struct {
	ServerID string `json:"serverID"`
	Config   string `json:"config"`
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
type InitUploadReq struct {
	UpSpeed  float32 `json:"upSpeed"`
	FileSize int64   `json:"fileSize"`
}
type InitUploadResp struct {
	UploadID      string `json:"uploadID"`
	ChunkSize     int64  `json:"chunkSize"`
	TotalChunks   int64  `json:"totalChunks"`
	ParallelConns int    `json:"parallelConns"`
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
type UploadPart struct {
	Data string `json:"data"`
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
type CompleteUploadReq struct {
	UploadID string `json:"uploadID"` // upload Id

	EncFileInfo      string `json:"encFileInfo"`      // base64(encrypt(FileInfo))
	EncFileInfoNonce string `json:"encFileInfoNonce"` // base64(encrypt(FileInfo))

	EncMeta   string `json:"metadata"`  // base64(encrypt(MetaWrapper))
	MetaNonce string `json:"metaNonce"` // base64(nonce from encrypt(MetaWrapper))

	EncDataChecksum string `json:"dataChecksum"` // sha(Data), sha is already base64()
	EncMetaChecksum string `json:"metaChecksum"` // sha(EncMeta)
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
type GetFilesListResp struct {
	Files []FilesListItem `json:"files"`
}

type FilesListItem struct {
	CreatedAt     time.Time `json:"createdAt"`
	FileUUID      string    `json:"fileUUID"`
	EncFileInfo   string    `json:"encFileInfo"`
	FileInfoNonce string    `json:"fileInfoNonce"`
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type DownloadInitResp struct{}

type DownloadMetaResp struct {
	EncMeta   string `json:"metadata"`  // base64(encrypt(MetaWrapper))
	MetaNonce string `json:"metaNonce"` // base64(nonce from encrypt(MetaWrapper))
}

type DownloadDigestResp struct {
	EncDataChecksum string `json:"dataChecksum"` // sha(Data), sha is already base64()
	EncMetaChecksum string `json:"metaChecksum"` // sha(EncMeta)
}
