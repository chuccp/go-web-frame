package localcahe

import (
	"bufio"

	"os"
	"path"

	"emperror.dev/errors"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
)

type Config struct {
	Path string
	Open bool
}

type LocalCache struct {
	config *Config
}

func (l *LocalCache) Init(cfg config2.IConfig) error {
	var config Config
	err := cfg.Unmarshal("local_cache", config)
	if err != nil {
		return errors.WithStackIf(err)
	}
	l.config = &config
	if len(config.Path) == 0 {
		config.Open = false
	}
	log.Info("cache:", zap.String("path", config.Path), zap.Bool("write data to the file", config.Open))
	err = util.CreateDirIfNoExists(config.Path)
	if err != nil {
		return errors.WithStackIf(err)
	}
	return nil
}

func (l *LocalCache) Destroy() error {
	return nil
}
func (l *LocalCache) getKey(key ...any) string {
	b := new(buffer.Buffer)
	for _, k := range key {
		b.AppendString(cast.ToString(k))
	}
	return util.MD5Str(b.String())
}
func (l *LocalCache) GetPath(value ...any) string {
	filename := l.getKey(value...)
	filepath := path.Join(l.config.Path, filename[0:2], filename)
	return filepath
}
func (l *LocalCache) SaveBase64File(base64file string) (string, error) {
	savePath := l.GetPath(base64file)
	data, err := util.DecodeFileBase64(base64file)
	if err != nil {
		return "", err
	}
	err = util.WriteFile(data, savePath)
	if err != nil {
		return "", err
	}
	return savePath, nil
}
func (l *LocalCache) SaveBase64FileForPath(base64file string, savePath string, suffix string) (string, error) {
	filename := l.getKey(base64file)
	filename = filename + "." + suffix
	saveFilePath := path.Join(savePath, filename[0:2], filename)
	data, err := util.DecodeFileBase64(base64file)
	if err != nil {
		return "", err
	}
	err = util.WriteFile(data, saveFilePath)
	if err != nil {
		return "", err
	}
	return path.Join(filename[0:2], filename), nil
}

func (l *LocalCache) GetFileForSuffix(suffix string, f func(value ...any) ([]byte, error), value ...any) (*web.File, error) {
	file, err := l.GetFile(f, value...)
	if err == nil {
		file.Suffix = suffix
		return file, nil
	}
	return nil, err
}
func (l *LocalCache) HasFile(value ...any) bool {
	filepath := l.GetPath(value...)
	return util.ExistsFile(filepath)
}
func (l *LocalCache) GetFileResponseWrite(response web.Response, f func(fileResponseWriteCloser *FileResponseWriteCloser, value ...any) error, value ...any) error {
	if len(value) == 0 {
		log.Panic("value len is zero")
	}
	if !l.config.Open {
		fileResponseWriteCloser := createFileResponseWriteCloser(response, nil)
		err := f(fileResponseWriteCloser, value...)
		if err != nil {
			return err
		}
		return nil
	}
	filename := l.getKey(value...)
	fileDir := path.Join(l.config.Path, filename[0:2])
	filepath := path.Join(fileDir, filename)
	if util.ExistsFile(filepath) {
		file, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Error("file close fail:", zap.Error(err))
			}
		}(file)
		var reader = bufio.NewReader(file)
		_, err = reader.WriteTo(response)
		if err != nil {
			return err
		}
		return nil
	}
	err := util.CreateDirIfNoExists(fileDir)
	if err != nil {
		return err
	}
	writeFile, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func(writeFile *os.File) {
		err := writeFile.Close()
		if err != nil {
			log.Error("file close fail:", zap.Error(err))
		}
	}(writeFile)
	fileResponseWriteCloser := createFileResponseWriteCloser(response, writeFile)
	err = f(fileResponseWriteCloser, value...)
	if err != nil {
		return err
	}
	return nil
}

func (l *LocalCache) GetFile(f func(value ...any) ([]byte, error), value ...any) (*web.File, error) {
	filepath := l.GetPath(value...)
	if util.ExistsFile(filepath) {
		return &web.File{Path: filepath}, nil
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
		err := errors.WithStackIf(writeFile.Close())
		if err != nil {
			log.Error("file close fail:", zap.Error(err))
		}
	}(writeFile)
	_, err = writeFile.Write(data)
	if err != nil {
		return nil, err
	}
	return &web.File{Path: filepath}, nil
}

type FileResponseWriteCloser struct {
	response web.Response
	file     *os.File
}

func (w *FileResponseWriteCloser) Write(p []byte) (n int, err error) {
	num, err := w.response.Write(p)
	if err != nil {
		return num, err
	}
	if w.file == nil {
		return num, err
	}
	return w.file.Write(p)
}
func (w *FileResponseWriteCloser) Close() error {
	w.response.Flush()
	if w.file == nil {
		return nil
	}
	err := w.file.Sync()
	if err != nil {
		return err
	}
	return nil
}

func createFileResponseWriteCloser(response web.Response, file *os.File) *FileResponseWriteCloser {
	return &FileResponseWriteCloser{
		response: response,
		file:     file,
	}
}
