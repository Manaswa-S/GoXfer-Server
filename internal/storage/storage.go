package storage

import (
	"context"

	"github.com/google/uuid"
)

type StorageType string

const (
	Local StorageType = "local"
)

type Config struct {
	Type    StorageType
	BaseDir string
	TempDir string
}

type TransferParams struct {
	Data   string
	Meta   []byte
	Digest []byte
}
type TransferReturns struct {
	Dir        string
	DataBase   string
	MetaBase   string
	DigestBase string
}

type RetrieveParams struct {
	Dir        string
	DataBase   string
	MetaBase   string
	DigestBase string
}
type RetrieveReturns struct {
	Data   string
	Meta   []byte
	Digest []byte
}

// THIS IS THE ROOT STORAGE AND SHOULD NOT BE USED FOR INPROCESS OPERATIONS.
// CONSIDER THIS AS A FAR AWAY S3 BUCKET.
type Storage interface {
	// Transfer copies the local path to the storage
	Transfer(ctx context.Context, params TransferParams, id uuid.UUID) (returns *TransferReturns, err error)

	// Retrieves the raw bytes, of given ID
	Retrieve(ctx context.Context, params RetrieveParams, id uuid.UUID) (returns *RetrieveReturns, err error)

	// Deletes the file, given by ID
	Delete(id uuid.UUID) (err error)

	// Returns existence and some metadata proving it, of the given ID
	Exists(id uuid.UUID) (metadata *Meta, err error)
	// Performs all pre-defined checks like checksum, etc on the given file by ID
	Check(ctx context.Context, id uuid.UUID) (err error)
	// Returns full stats of the given ID
	Stats(id uuid.UUID) (stats *Stats, err error)

	// Returns the internal config
	Config() Config
}

type Meta struct {
}

type Stats struct {
}
