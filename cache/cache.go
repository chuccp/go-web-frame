package cache

import (
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
func (l *LocalCache) GetFile(f func(value ...any) ([]byte, error), value ...any) (*os.File, error) {
	filename := l.getKey(value...)
	filepath := path.Join(l.path, filename[0:2], filename)
	return util.ReadOrWriteFile(filepath, func() ([]byte, error) {
		return f(value...)
	})
}
