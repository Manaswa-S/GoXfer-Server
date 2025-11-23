package localStore

import (
	"bytes"
	"context"
	"goxfer/server/internal/storage"
	"io"
	"os"

	"github.com/google/uuid"
)

type Local struct {
	cfg *storage.Config
}

func NewLocalStorage(baseDir, tempDir string) (*Local, error) {
	if err := os.MkdirAll(baseDir, 0744); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(tempDir, 0744); err != nil {
		return nil, err
	}

	return &Local{
		cfg: &storage.Config{
			Type:    storage.Local,
			BaseDir: baseDir,
			TempDir: tempDir,
		},
	}, nil
}

// TODO: should be atomic ?
func (s *Local) Transfer(ctx context.Context, params storage.TransferParams, id uuid.UUID) (returns *storage.TransferReturns, err error) {
	returns = new(storage.TransferReturns)
	returns.DataBase, err = s.transferData(params.Data, id)
	if err != nil {
		return nil, err
	}
	returns.MetaBase, err = s.transferMeta(params.Meta, id)
	if err != nil {
		return nil, err
	}

	returns.DigestBase, err = s.transferDigest(params.Digest, id)
	if err != nil {
		return nil, err
	}

	returns.Dir = s.cfg.BaseDir

	return returns, nil
}

func (s *Local) transferData(data string, id uuid.UUID) (base string, err error) {
	dataSrc, err := os.Open(data)
	if err != nil {
		return "", err
	}

	dataBase := s.buildDataBase(id.String())
	dataPath := s.buildPath(dataBase)
	dst, err := os.Create(dataPath)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(dst, dataSrc)
	if err != nil {
		return "", err
	}

	return dataBase, nil
}

func (s *Local) transferMeta(meta []byte, id uuid.UUID) (base string, err error) {
	metaBase := s.buildMetaBase(id.String())
	metaPath := s.buildPath(metaBase)
	dst, err := os.Create(metaPath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(dst, bytes.NewReader(meta))
	if err != nil {
		return "", err
	}

	return metaBase, nil
}

func (s *Local) transferDigest(digest []byte, id uuid.UUID) (loc string, err error) {
	digBase := s.buildDigestBase(id.String())
	digPath := s.buildPath(digBase)
	dst, err := os.Create(digPath)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(dst, bytes.NewReader(digest))
	if err != nil {
		return "", err
	}

	return digBase, nil
}

// >>>

func (s *Local) Retrieve(ctx context.Context, params storage.RetrieveParams, id uuid.UUID) (returns *storage.RetrieveReturns, err error) {

	return nil, nil
}

// >>>

func (s *Local) Delete(id uuid.UUID) (err error) {
	idStr := id.String()
	dataPath := s.buildPath(s.buildDataBase(idStr))
	metaPath := s.buildPath(s.buildMetaBase(idStr))
	digPath := s.buildPath(s.buildDigestBase(idStr))

	err = os.Remove(dataPath)
	if err != nil {
		return err
	}

	err = os.Remove(metaPath)
	if err != nil {
		return err
	}

	err = os.Remove(digPath)
	if err != nil {
		return err
	}

	return nil
}

// >>>

func (s *Local) Exists(id uuid.UUID) (metadata *storage.Meta, err error) {
	return nil, nil
}

func (s *Local) Check(ctx context.Context, id uuid.UUID) (err error) {
	return nil
}

func (s *Local) Stats(id uuid.UUID) (stats *storage.Stats, err error) {
	return nil, nil
}

// >>>

func (s *Local) Config() storage.Config {
	return *s.cfg
}
