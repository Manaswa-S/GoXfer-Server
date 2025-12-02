package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"goxfer/server/internal/auth"
	"goxfer/server/internal/consts/errs"
	"goxfer/server/internal/dto"
	"goxfer/server/internal/middleware"
	"goxfer/server/internal/storage"
	sqlc "goxfer/server/internal/store/sqlc/generate"
	transfer "goxfer/server/internal/transfer"
	"goxfer/server/internal/utils"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bytemare/opaque"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	queries *sqlc.Queries
	redis   *redis.Client
	opaque  *Opaque
	auth    auth.Authenticator
	storage storage.Storage
	upload  *transfer.Upload

	build  *Builder
	consts *Consts
}

type Consts struct {
	CredIDCacheTTL     time.Duration // in seconds
	NewLoginSessionTTL time.Duration // in seconds
}

func NewService(queries *sqlc.Queries, redis *redis.Client,
	opaque *Opaque, auth auth.Authenticator, storage storage.Storage,
	upload *transfer.Upload, build *Builder) *Service {
	return &Service{
		queries: queries,
		redis:   redis,
		opaque:  opaque,
		auth:    auth,
		storage: storage,
		upload:  upload,
		build:   build,

		consts: &Consts{
			CredIDCacheTTL:     time.Duration(300) * time.Second,
			NewLoginSessionTTL: time.Duration(1800) * time.Second,
		},
	}
}

// >>>

func (s *Service) getCtxVals(ctx *gin.Context) string {
	return ctx.GetString(middleware.CtxSetBucKey)
}

// >>>

// TODO: replace all plain bytes with base64
func (s *Service) CreateBucketS1(ctx *gin.Context, req *dto.CreateBucketS1Req) (*dto.CreateBucketS1Resp, *errs.Errorf) {
	request, err := s.opaque.server.Deserialize.RegistrationRequest(utils.DecodeBase64(req.S1Req))
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	credID, err := utils.Rand(64)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	pks, err := s.opaque.server.Deserialize.DecodeAkePublicKey(s.opaque.publicKey)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	response := s.opaque.server.RegistrationResponse(request, pks, credID, s.opaque.oprfSeed)

	id, err := uuid.NewV7()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	reqID := id.String()
	credKey := s.build.Rs.credKey(reqID)
	err = s.redis.SetEx(ctx, credKey, credID, s.consts.CredIDCacheTTL).Err()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	return &dto.CreateBucketS1Resp{
		S1Resp:   utils.EncodeBase64(response.Serialize()),
		ReqID:    reqID,
		ServerID: utils.EncodeBase64(s.opaque.serverID),
	}, nil
}

func (s *Service) CreateBucketS2(ctx *gin.Context, req *dto.CreateBucketS2Req) (*dto.CreateBucketS2Resp, *errs.Errorf) {
	if req.Cipher == "" {
		return nil, &errs.Errorf{}
	}

	credKey := s.build.Rs.credKey(req.ReqID)
	result, err := s.redis.Get(ctx, credKey).Result()
	if err != nil || result == "" {
		if err == redis.Nil {
			return nil, &errs.Errorf{Error: fmt.Errorf("invalid request")}
		}
		return nil, &errs.Errorf{Error: err}
	}
	credID := []byte(result)

	record, err := s.opaque.server.Deserialize.RegistrationRecord(utils.DecodeBase64(req.S2Req))
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	bucKey, err := utils.GenerateBucketKey()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	if _, err = s.queries.InsertBucket(ctx, sqlc.InsertBucketParams{
		Key:    bucKey,
		Name:   req.BucName,
		CredID: utils.EncodeBase64(credID),
		Record: utils.EncodeBase64(record.Serialize()),
		Cipher: req.Cipher,
	}); err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	time.Sleep(3 * time.Second)

	return &dto.CreateBucketS2Resp{
		BucketKey: bucKey,
		Name:      req.BucName,
	}, nil
}

func (s *Service) GetConfigs(ctx *gin.Context) (*dto.GetOpaqueConfigs, *errs.Errorf) {
	return &dto.GetOpaqueConfigs{
		ServerID: utils.EncodeBase64(s.opaque.serverID),
		Config:   utils.EncodeBase64(s.opaque.config),
	}, nil
}

// TODO: ttl needed
func (s *Service) OpenBucketS1(ctx *gin.Context, req *dto.OpenBucketS1Req) (*dto.OpenBucketS1Resp, *errs.Errorf) {
	bucKey := string(utils.DecodeBase64(req.BucketKey))
	fmt.Println(bucKey)

	bucket, err := s.queries.GetBucket(ctx, bucKey)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	server, err := s.newOpaqueServer()
	if err != nil {
		return nil, &errs.Errorf{Error: fmt.Errorf("new opaque server error: %v", err)}
	}

	record, err := server.Deserialize.RegistrationRecord(utils.DecodeBase64(bucket.Record))
	if err != nil {
		return nil, &errs.Errorf{Error: fmt.Errorf("deserialize registration record error: %v", err)}
	}

	ke1, err := server.Deserialize.KE1(utils.DecodeBase64(req.KE1))
	if err != nil {
		return nil, &errs.Errorf{Error: fmt.Errorf("deserialize KE1 error: %v", err)}
	}

	savedRecord := &opaque.ClientRecord{
		CredentialIdentifier: utils.DecodeBase64(bucket.CredID),
		ClientIdentity:       []byte(bucket.Name),
		RegistrationRecord:   record,
	}
	ke2, err := server.LoginInit(ke1, savedRecord)
	if err != nil {
		return nil, &errs.Errorf{Error: fmt.Errorf("login init error: %v", err)}
	}

	loginId, err := uuid.NewV7()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	if err = s.setOpaqueLoginServer(loginId, server); err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	cacheKey := s.build.Rs.loginIdKey(loginId.String())
	if err = s.redis.Set(ctx, cacheKey, bucKey, 0).Err(); err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	return &dto.OpenBucketS1Resp{
		KE2:      utils.EncodeBase64(ke2.Serialize()),
		ClientID: bucket.Name,
		LoginID:  loginId.String(),
	}, nil
}

func (s *Service) OpenBucketS2(ctx *gin.Context, req *dto.OpenBucketS2Req) (*dto.OpenBucketS2Resp, *errs.Errorf) {
	loginID, err := uuid.Parse(req.LoginID)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	server := s.getOpaqueLoginServer(loginID)
	if server == nil {
		return nil, &errs.Errorf{Error: err}
	}

	ke3, err := server.Deserialize.KE3(utils.DecodeBase64(req.KE3))
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	err = server.LoginFinish(ke3)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	s.deleteOpaqueLoginServer(loginID)

	sessionID := loginID.String()

	cacheKey := s.build.Rs.loginIdKey(req.LoginID)
	bucKey, err := s.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	if err = s.auth.NewSession(ctx, sessionID, bucKey, server.SessionKey(), s.consts.NewLoginSessionTTL); err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	bucket, err := s.queries.GetBucket(ctx, bucKey)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	return &dto.OpenBucketS2Resp{
		SessionID:  sessionID,
		SessionTTL: int64(s.consts.NewLoginSessionTTL),
		Cipher:     bucket.Cipher,
	}, nil
}

func (s *Service) InitUpload(ctx *gin.Context, req *dto.InitUploadReq) (*dto.InitUploadResp, *errs.Errorf) {
	plan, err := s.upload.New(req.UpSpeed, req.FileSize)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	cacheBytes, err := json.Marshal(dto.UploadFileCache{
		FileSize: req.FileSize,
		Plan:     plan,
	})
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	key := s.build.Rs.cacheUploadMeta(plan.UploadID)
	if err = s.redis.Set(ctx, key, cacheBytes, 0).Err(); err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	return &dto.InitUploadResp{
		UploadID:      plan.UploadID,
		ChunkSize:     plan.ChunkSize,
		TotalChunks:   plan.TotalChunks,
		ParallelConns: plan.ParallelConns,
	}, nil
}

// The same part can be replaced any number of times.
// It just replaces the old with new.
// TODO: The upload window should be TTL`ed
func (s *Service) UploadPart(ctx *gin.Context, uploadId, chunkId string) *errs.Errorf {
	chunkID, err := strconv.ParseInt(chunkId, 10, 64)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	// CHECK IF uploadId is valid
	key := s.build.Rs.cacheUploadMeta(uploadId)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	if exists != 1 {
		return &errs.Errorf{Error: fmt.Errorf("invalid upload ID")}
	}

	// SAVE the chunk
	partsDir := s.build.FS.partsDir(uploadId)
	if err = os.MkdirAll(partsDir, 0755); err != nil {
		return &errs.Errorf{Error: err}
	}
	chunkPath := s.build.FS.chunkPath(partsDir, chunkId)
	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	defer chunkFile.Close()
	if _, err = io.Copy(chunkFile, ctx.Request.Body); err != nil {
		return &errs.Errorf{Error: err}
	}

	setKey := s.build.Rs.cacheUploadSet(uploadId)
	if err = s.redis.SAdd(ctx, setKey, chunkID).Err(); err != nil {
		return &errs.Errorf{Error: err}
	}

	return nil
}

// TODO: not atomic
func (s *Service) CompleteUpload(ctx *gin.Context, req *dto.CompleteUploadReq) *errs.Errorf {
	// Validata and correlate all chunks
	key := s.build.Rs.cacheUploadMeta(req.UploadID)
	cacheBytes, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return &errs.Errorf{Error: fmt.Errorf("unknown upload request")}
		}
		return &errs.Errorf{Error: err}
	}
	cache := new(dto.UploadFileCache)
	if err = json.Unmarshal([]byte(cacheBytes), cache); err != nil {
		return &errs.Errorf{Error: err}
	}

	setKey := s.build.Rs.cacheUploadSet(req.UploadID)
	indexSet, err := s.redis.SMembers(ctx, setKey).Result()
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	if int64(len(indexSet)) < cache.Plan.TotalChunks {
		return &errs.Errorf{Error: fmt.Errorf("incomplete uploads")}
	}
	completed := make([]bool, len(indexSet))
	for _, idx := range indexSet {
		n, err := strconv.ParseInt(idx, 10, 64)
		if err != nil {
			return &errs.Errorf{Error: err}
		}
		completed[n] = true
	}
	for i := int64(0); i < cache.Plan.TotalChunks; i++ {
		if !completed[i] {
			return &errs.Errorf{Error: fmt.Errorf("missing chunk")}
		}
	}

	// Join the chunks, create the final file
	id, err := uuid.NewV7()
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	fileID := id.String()

	partsDir := s.build.FS.partsDir(req.UploadID)
	finalDir := s.build.FS.finalDir(req.UploadID)
	finalPath := s.build.FS.finalPath(finalDir, fileID)

	if err = os.MkdirAll(finalDir, 0744); err != nil {
		return &errs.Errorf{Error: err}
	}
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	defer finalFile.Close()

	copied := int64(0)
	for i := int64(0); i < cache.Plan.TotalChunks; i++ {
		chunkPath := s.build.FS.chunkPath(partsDir, strconv.FormatInt(i, 10))
		chunk, err := os.Open(chunkPath)
		if err != nil {
			return &errs.Errorf{Error: err}
		}
		defer chunk.Close()

		n, err := io.Copy(finalFile, chunk)
		if err != nil {
			return &errs.Errorf{Error: err}
		}
		copied += n
		chunk.Close()
	}
	finalFile.Close()

	// Verify final file and metadata
	stats, err := os.Stat(finalPath)
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	if stats.Size() != (cache.FileSize) {
		return &errs.Errorf{Error: fmt.Errorf("final sizes dont match")}
	}

	dataChksm, err := utils.GenChecksumSHA(finalPath)
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	if dataChksm != req.EncDataChecksum {
		return &errs.Errorf{Error: fmt.Errorf("data checksums dont match")}
	}

	metaChksm, err := utils.GenChecksumSHABytes(utils.DecodeBase64(req.EncMeta))
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	if metaChksm != req.EncMetaChecksum {
		return &errs.Errorf{Error: fmt.Errorf("meta checksums dont match")}
	}

	// Tranfer final file, metadata and digest to storage
	xferMeta := dto.TransferMeta{
		EncMeta:   req.EncMeta,
		MetaNonce: req.EncMetaNonce,
	}
	xferMetaBytes, err := json.Marshal(xferMeta)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	xferDig := dto.TransferDigest{
		EncDataChecksum: req.EncDataChecksum,
		EncMetaChecksum: req.EncMetaChecksum,
	}
	xferDigBytes, err := json.Marshal(xferDig)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	xfer, err := s.storage.Transfer(ctx, storage.TransferParams{
		Data:   finalPath,
		Meta:   xferMetaBytes,
		Digest: xferDigBytes,
	}, id)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	// TODO: this shouldn't be removed immediately,
	// transfer will happen async so this might be a problem
	err = os.RemoveAll(finalDir)
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	err = os.RemoveAll(partsDir)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	bucKey := s.getCtxVals(ctx)
	bucID, err := s.queries.GetBucketID(ctx, bucKey)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	err = s.queries.InsertNewFile(ctx, sqlc.InsertNewFileParams{
		BucID:         bucID,
		UploadID:      req.UploadID,
		Valid:         true,
		DataFile:      xfer.DataBase,
		FileUuid:      pgtype.UUID{Bytes: id, Valid: true},
		MetaFile:      xfer.MetaBase,
		DigestFile:    xfer.DigestBase,
		BasePath:      xfer.Dir,
		FileInfo:      req.EncFileInfo,
		FileInfoNonce: req.EncFileInfoNonce,
		DataFileSize:  copied,
	})
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	// Insert into DB and setup other records
	fmt.Println(xfer.DataBase)
	fmt.Println(xfer.MetaBase)
	fmt.Println(xfer.DigestBase)

	fmt.Println(req.EncFileInfo)

	return nil
}

func (s *Service) GetFilesList(ctx *gin.Context) (*dto.GetFilesListResp, *errs.Errorf) {
	bucKey := s.getCtxVals(ctx)

	bucID, err := s.queries.GetBucketID(ctx, bucKey)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	files, err := s.queries.GetFiles(ctx, bucID)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	list := make([]dto.FilesListItem, 0)
	for _, file := range files {
		list = append(list, dto.FilesListItem{
			CreatedAt:     file.CreatedAt.Time,
			FileUUID:      file.FileUuid.String(),
			EncFileInfo:   file.FileInfo,
			FileInfoNonce: file.FileInfoNonce,
		})
	}

	return &dto.GetFilesListResp{
		Files: list,
	}, nil
}

const (
	StageDownloadInit   = "init"
	StageDownloadData   = "data"
	StageDownloadMeta   = "meta"
	StageDownloadDigest = "digest"
)

func (s *Service) DownloadInit(ctx *gin.Context, fileId string) (*dto.DownloadInitResp, *errs.Errorf) {

	err := uuid.Validate(fileId)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	fileUUID, err := uuid.Parse(fileId)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	bucKey := s.getCtxVals(ctx)
	bucID, err := s.queries.GetBucketID(ctx, bucKey)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	_, err = s.queries.GetFileID(ctx, sqlc.GetFileIDParams{
		FileUuid: pgtype.UUID{Bytes: fileUUID, Valid: true},
		BucID:    bucID,
	})
	if err != nil {
		// TODO: not found
		return nil, &errs.Errorf{Error: err}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	key := s.build.Rs.cacheDownloadSet(id.String())
	err = s.redis.Set(ctx, key, StageDownloadInit, 0).Err()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	ctx.Header("X-Download-ID", id.String())

	return &dto.DownloadInitResp{}, nil
}

func (s *Service) DownloadData(ctx *gin.Context, fileId string) *errs.Errorf {

	err := uuid.Validate(fileId)
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	fileUUID, err := uuid.Parse(fileId)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	downId := ctx.Request.Header.Get("X-Download-ID")
	if downId == "" {
		return &errs.Errorf{}
	}
	key := s.build.Rs.cacheDownloadSet(downId)
	stage, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	if stage != StageDownloadInit {
		return &errs.Errorf{}
	}

	bucKey := s.getCtxVals(ctx)
	bucID, err := s.queries.GetBucketID(ctx, bucKey)
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	locs, err := s.queries.GetFileLoc(ctx, sqlc.GetFileLocParams{
		FileUuid: pgtype.UUID{Bytes: fileUUID, Valid: true},
		BucID:    bucID,
	})
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	fileMeta, err := s.queries.GetFileMeta(ctx, locs.FileID)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	dataPath := filepath.Join(locs.BasePath, locs.DataFile)

	f, err := os.Open(dataPath)
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	defer f.Close()

	// ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(path)))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Length", fmt.Sprintf("%d", fileMeta.DataFileSize))
	ctx.Header("Cache-Control", "private, no-store")

	n, err := io.Copy(ctx.Writer, f)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	if n != fileMeta.DataFileSize {
		return &errs.Errorf{}
	}

	err = s.redis.Set(ctx, key, StageDownloadData, 0).Err()
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	return nil
}

func (s *Service) DownloadMeta(ctx *gin.Context, fileId string) (*dto.DownloadMetaResp, *errs.Errorf) {

	err := uuid.Validate(fileId)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	fileUUID, err := uuid.Parse(fileId)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	downId := ctx.Request.Header.Get("X-Download-ID")
	if downId == "" {
		return nil, &errs.Errorf{Error: fmt.Errorf("download id not found")}
	}
	key := s.build.Rs.cacheDownloadSet(downId)
	stage, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	if stage != StageDownloadData {
		return nil, &errs.Errorf{Error: fmt.Errorf("stage not as expected")}
	}

	bucKey := s.getCtxVals(ctx)
	bucID, err := s.queries.GetBucketID(ctx, bucKey)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	locs, err := s.queries.GetFileLoc(ctx, sqlc.GetFileLocParams{
		FileUuid: pgtype.UUID{Bytes: fileUUID, Valid: true},
		BucID:    bucID,
	})
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	metaPath := filepath.Join(locs.BasePath, locs.MetaFile)
	f, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	xferMeta := new(dto.TransferMeta)
	err = json.Unmarshal(f, xferMeta)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	ctx.Header("Cache-Control", "private, no-store")

	err = s.redis.Set(ctx, key, StageDownloadMeta, 0).Err()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	return &dto.DownloadMetaResp{
		EncMeta:   xferMeta.EncMeta,
		MetaNonce: xferMeta.MetaNonce,
	}, nil
}

func (s *Service) DownloadDigest(ctx *gin.Context, fileId string) (*dto.DownloadDigestResp, *errs.Errorf) {
	err := uuid.Validate(fileId)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	fileUUID, err := uuid.Parse(fileId)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	downId := ctx.Request.Header.Get("X-Download-ID")
	if downId == "" {
		return nil, &errs.Errorf{Error: fmt.Errorf("download id not found")}
	}
	key := s.build.Rs.cacheDownloadSet(downId)
	stage, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	if stage != StageDownloadMeta {
		return nil, &errs.Errorf{Error: fmt.Errorf("stage not as expected")}
	}

	bucKey := s.getCtxVals(ctx)
	bucID, err := s.queries.GetBucketID(ctx, bucKey)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}
	locs, err := s.queries.GetFileLoc(ctx, sqlc.GetFileLocParams{
		FileUuid: pgtype.UUID{Bytes: fileUUID, Valid: true},
		BucID:    bucID,
	})
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	digestPath := filepath.Join(locs.BasePath, locs.DigestFile)
	f, err := os.ReadFile(digestPath)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	xferDig := new(dto.TransferDigest)
	err = json.Unmarshal(f, xferDig)
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	ctx.Header("Cache-Control", "private, no-store")

	err = s.redis.Set(ctx, key, StageDownloadDigest, 0).Err()
	if err != nil {
		return nil, &errs.Errorf{Error: err}
	}

	return &dto.DownloadDigestResp{
		EncDataChecksum: xferDig.EncDataChecksum,
		EncMetaChecksum: xferDig.EncMetaChecksum,
	}, nil
}

func (s *Service) DeleteFile(ctx *gin.Context, fileuuid string) *errs.Errorf {
	err := uuid.Validate(fileuuid)
	if err != nil {
		return &errs.Errorf{Error: err}
	}
	fileUUID, err := uuid.Parse(fileuuid)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	bucKey := s.getCtxVals(ctx)
	bucID, err := s.queries.GetBucketID(ctx, bucKey)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	fileID, err := s.queries.GetFileID(ctx, sqlc.GetFileIDParams{
		FileUuid: pgtype.UUID{Bytes: fileUUID, Valid: true},
		BucID:    bucID,
	})
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	err = s.storage.Delete(fileUUID)
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	err = s.queries.UpdateFileValidity(ctx, sqlc.UpdateFileValidityParams{
		FileID: fileID,
		Valid:  false,
	})
	if err != nil {
		return &errs.Errorf{Error: err}
	}

	return nil
}
