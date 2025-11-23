package transfer

import (
	"math"

	"github.com/google/uuid"
)

type Upload struct {
	cfg UploadCfg
}

type UploadCfg struct {
	baseUpSpeed      float32
	baseChunkSize    int64
	maxParallelConns int
	connDensity      int64
	chunkSizeMul     int
}

func NewUpload() *Upload {
	return &Upload{
		cfg: UploadCfg{
			baseUpSpeed:      12.00, // 12 mbps
			baseChunkSize:    8 * 1024 * 1024,
			maxParallelConns: 6,
			connDensity:      3,
			chunkSizeMul:     2,
		},
	}
}

type UploadPlan struct {
	UploadID      string
	ChunkSize     int64
	TotalChunks   int64
	ParallelConns int
}

func (s *Upload) getTotalChunks(chunkSize, totalSize int64) int64 {
	return int64(math.Ceil(float64(totalSize) / float64(chunkSize)))
}

func (s *Upload) getChunkSize(upSpeed float32, totalSize int64) int64 {
	chunkSize := int64(1024)
	if upSpeed < s.cfg.baseUpSpeed {
		chunkSize = s.cfg.baseChunkSize
	} else {
		chunkSize = s.cfg.baseChunkSize * int64(s.cfg.chunkSizeMul)
	}
	return int64(math.Min(float64(chunkSize), float64(totalSize)))
}

func (s *Upload) getParallelConns(totalChunks int64) int {
	return int(math.Min(float64(s.cfg.maxParallelConns), math.Max(1, float64(totalChunks/s.cfg.connDensity))))
}

func (s *Upload) New(upSpeed float32, totalSize int64) (*UploadPlan, error) {

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	chunkSize := s.getChunkSize(upSpeed, totalSize)
	totalChunks := s.getTotalChunks(chunkSize, totalSize)
	parallelConns := s.getParallelConns(totalChunks)

	return &UploadPlan{
		UploadID:      id.String(),
		ChunkSize:     chunkSize,
		TotalChunks:   totalChunks,
		ParallelConns: parallelConns,
	}, nil
}
