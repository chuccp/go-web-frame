package web

import (
	"log"
	"os"
	"path"

	"github.com/chuccp/go-web-frame/util"
	"github.com/spf13/cast"
	"go.uber.org/zap/buffer"
)

type LocalCache struct {
	path string
}

func NewLocalCache(path string) *LocalCache {
	return &LocalCache{path: path}
}
func (l *LocalCache) getKey(key ...any) string {
	b := new(buffer.Buffer)
	for _, k := range key {
		b.AppendString(cast.ToString(k))
	}
	return util.MD5Str(b.String())
}
func (l *LocalCache) GetFile(f func(value ...any) ([]byte, error), value ...any) (*File, error) {
	filename := l.getKey(value...)
	fileDir := path.Join(l.path, filename[0:2])
	err := util.CreateDirIfNoExists(fileDir)
	if err != nil {
		return nil, err
	}
	filepath := path.Join(l.path, filename[0:2], filename)
	if util.ExistsFile(filepath) {
		return &File{Path: filepath}, nil
	}
	data, err := f(value...)
	if err != nil {
		return nil, err
	}
	writeFile, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	defer func(writeFile *os.File) {
		err := writeFile.Close()
		if err != nil {
			log.Println("file close fail:", err)
		}
	}(writeFile)
	_, err = writeFile.Write(data)
	if err != nil {
		return nil, err
	}
	return &File{Path: filepath}, nil
}
