package web

import (
	"bufio"
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
func (l *LocalCache) GetPath(value ...any) string {
	filename := l.getKey(value...)
	filepath := path.Join(l.path, filename[0:2], filename)
	return filepath
}

func (l *LocalCache) GetFileForSuffix(suffix string, f func(value ...any) ([]byte, error), value ...any) (*File, error) {
	file, err := l.GetFile(f, value...)
	if err == nil {
		file.Suffix = suffix
		return file, nil
	}
	return nil, err
}
func (l *LocalCache) HasFile(value ...any) bool {
	filepath := l.GetPath(value...)
	return util.FileExists(filepath)
}
func (l *LocalCache) GetFileResponseWrite(response Response, f func(fileResponseWriteCloser *FileResponseWriteCloser, value ...any) error, value ...any) error {
	if len(value) == 0 {
		log.Panicln("value len is zero")
	}
	filename := l.getKey(value...)
	fileDir := path.Join(l.path, filename[0:2])
	filepath := path.Join(fileDir, filename)
	if util.ExistsFile(filepath) {
		file, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Println("file close fail:", err)
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
			log.Println("file close fail:", err)
		}
	}(writeFile)
	fileResponseWriteCloser := createFileResponseWriteCloser(response, writeFile)
	err = f(fileResponseWriteCloser, value...)
	if err != nil {
		return err
	}
	return nil
}

func (l *LocalCache) GetFile(f func(value ...any) ([]byte, error), value ...any) (*File, error) {
	filepath := l.GetPath(value...)
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

type FileResponseWriteCloser struct {
	response Response
	file     *os.File
}

func (w *FileResponseWriteCloser) Write(p []byte) (n int, err error) {
	num, err := w.response.Write(p)
	if err != nil {
		return num, err
	}
	return w.file.Write(p)
}
func (w *FileResponseWriteCloser) Close() error {
	w.response.Flush()
	err := w.file.Sync()
	if err != nil {
		return err
	}
	return w.file.Close()
}

func createFileResponseWriteCloser(response Response, file *os.File) *FileResponseWriteCloser {
	return &FileResponseWriteCloser{
		response: response,
		file:     file,
	}
}
