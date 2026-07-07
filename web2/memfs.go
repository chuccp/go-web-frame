package web2

import (
	"net/http"
	"os"
	"path"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// MemFileSystem is a memory-cached file system that reads from disk locations
// and caches files in memory for faster subsequent access.
type MemFileSystem struct {
	fs        afero.Fs
	locations []string
}

// Open opens a file by name, searching configured locations.
func (m *MemFileSystem) Open(name string) (http.File, error) {
	var err0 error
	for _, location := range m.locations {
		filePath := path.Join(location, name)
		exists, err := afero.Exists(m.fs, filePath)
		if err != nil {
			err0 = err
			continue
		}
		if exists {
			log.Info("open file", zap.String("filePath", filePath))
			open, err := m.fs.Open(filePath)
			if err != nil {
				log.Errors("open file", err)
				return nil, errors.WithStackIf(err)
			}
			return open, nil
		}
	}
	return nil, errors.WithStackIf(err0)

}

// Exists reports whether a file exists in any configured location.
func (m *MemFileSystem) Exists(name string) (bool, error) {
	for _, location := range m.locations {
		exists, err := afero.Exists(m.fs, path.Join(location, name))
		if err != nil {
			return false, errors.WithStackIf(err)
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

// Stat returns file info for the named file, searching configured locations.
func (m *MemFileSystem) Stat(name string) (os.FileInfo, error) {
	var err0 error
	for _, location := range m.locations {
		filePath := path.Join(location, name)
		exists, err := afero.Exists(m.fs, filePath)
		if err != nil {
			err0 = err
			continue
		}
		if exists {
			return m.fs.Stat(filePath)
		}
	}
	return nil, errors.WithStackIf(err0)
}

// NewMemFileSystem creates a MemFileSystem with the given cache duration and server config.
func NewMemFileSystem(cacheTime time.Duration, locations []string) *MemFileSystem {
	baseFs := afero.NewOsFs()
	cacheLayer := afero.NewMemMapFs()
	return &MemFileSystem{
		afero.NewCacheOnReadFs(baseFs, cacheLayer, cacheTime), locations,
	}
}

// DefaultMemFileSystem creates a MemFileSystem with a 10-minute cache duration.
func DefaultMemFileSystem(locations []string) *MemFileSystem {
	return NewMemFileSystem(10*time.Minute, locations)
}
