package service

import (
	"fmt"
	"path/filepath"
)

type Builder struct {
	// FS
	FS FileSystem
	// Redis
	Rs RedisBuilder
}

type FileSystem struct {
	BaseDir         string // should be persistent across restarts
	TempDir         string // can be non-persistent
	PartStorageDir  string // to store upload parts
	FinalStorageDir string // to store upload final files
}

type RedisBuilder struct {
	CredIDCacheKeyPrefix string
	LoginIdBucKeyPrefix  string
	UploadKeySetPrefix   string
	UploadKeyMetaPrefix  string
	DownloadKeySetPrefix string
}

func NewBuilder(baseDir, tempDir string) *Builder {
	return &Builder{
		FS: FileSystem{
			BaseDir:         baseDir,
			TempDir:         tempDir,
			PartStorageDir:  filepath.Join(tempDir, "/files/parts"),
			FinalStorageDir: filepath.Join(tempDir, "/files/final"),
		},
		Rs: RedisBuilder{
			CredIDCacheKeyPrefix: "buc-reg-cred-id",
			LoginIdBucKeyPrefix:  "buc-log-in-id",
			UploadKeySetPrefix:   "upload-id",
			UploadKeyMetaPrefix:  "upload-id-meta",
			DownloadKeySetPrefix: "download-stage",
		},
	}
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func (s *FileSystem) partsDir(uploadId string) string {
	return filepath.Join(s.PartStorageDir, uploadId)
}

func (s *FileSystem) chunkPath(partsDir, chunkId string) string {
	return filepath.Join(partsDir, fmt.Sprintf("chunk_%s.part", chunkId))
}

func (s *FileSystem) finalDir(uploadId string) string {
	return filepath.Join(s.FinalStorageDir, uploadId)
}

func (s *FileSystem) finalPath(finalDir, fileId string) string {
	return filepath.Join(finalDir, fmt.Sprintf("%s.data", fileId))
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func (s *RedisBuilder) credKey(reqId string) string {
	return fmt.Sprintf("%s:%s", s.CredIDCacheKeyPrefix, reqId)
}

func (s *RedisBuilder) loginIdKey(loginId string) string {
	return fmt.Sprintf("%s:%s", s.LoginIdBucKeyPrefix, loginId)
}

func (s *RedisBuilder) cacheUploadMeta(uploadId string) string {
	return fmt.Sprintf("%s:%s", s.UploadKeyMetaPrefix, uploadId)
}

func (s *RedisBuilder) cacheUploadSet(uploadId string) string {
	return fmt.Sprintf("%s:%s", s.UploadKeySetPrefix, uploadId)
}

func (s *RedisBuilder) cacheDownloadSet(downloadId string) string {
	return fmt.Sprintf("%s:%s", s.DownloadKeySetPrefix, downloadId)
}
