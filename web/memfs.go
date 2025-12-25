package web

import (
	"net/http"
	"os"
	"path"
	"time"

	"github.com/chuccp/go-web-frame/log"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

type MemFileSystem struct {
	fs           afero.Fs
	serverConfig *ServerConfig
}

func (m *MemFileSystem) Open(name string) (http.File, error) {
	var err0 error
	if m.noLocation() {
		return m.fs.Open(name)
	}
	for _, location := range m.serverConfig.Locations {
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
				return nil, err
			}
			return open, nil
		}
	}
	return nil, err0

}
func (m *MemFileSystem) Exists(name string) (bool, error) {

	if m.noLocation() {
		return afero.Exists(m.fs, name)
	}

	for _, location := range m.serverConfig.Locations {
		exists, err := afero.Exists(m.fs, path.Join(location, name))
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil

}

func (m *MemFileSystem) noLocation() bool {
	if m.serverConfig == nil || m.serverConfig.Locations == nil || len(m.serverConfig.Locations) == 0 {
		return true
	}
	return false
}

func (m *MemFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.noLocation() {
		return m.fs.Stat(name)
	}
	var err0 error
	for _, location := range m.serverConfig.Locations {
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
	return nil, err0
}
func NewMemFileSystem(cacheTime time.Duration, serverConfig *ServerConfig) *MemFileSystem {
	baseFs := afero.NewOsFs()
	cacheLayer := afero.NewMemMapFs()
	return &MemFileSystem{
		afero.NewCacheOnReadFs(baseFs, cacheLayer, cacheTime), serverConfig,
	}
}
func DefaultMemFileSystem(serverConfig *ServerConfig) *MemFileSystem {
	return NewMemFileSystem(10*time.Minute, serverConfig)
}
