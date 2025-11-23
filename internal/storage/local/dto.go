package localStore

import (
	"fmt"
	"path/filepath"
)

// Builders

func (s *Local) buildPath(base string) string {
	return filepath.Join(s.cfg.BaseDir, base)
}

func (s *Local) buildDataBase(id string) string {
	return fmt.Sprintf("%s.%s", id, "enc")
}

func (s *Local) buildMetaBase(id string) string {
	return fmt.Sprintf("%s.%s", id, "meta")
}

func (s *Local) buildDigestBase(id string) string {
	return fmt.Sprintf("%s.%s", id, "digest")
}
