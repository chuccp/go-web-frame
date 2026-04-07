package web

import (
	"path/filepath"
	"strings"
)

type File struct {
	Path     string
	FileName string
	Suffix   string
}

// CreateFile creates a File instance with auto-extracted filename and suffix
func CreateFile(path string) *File {
	fileName := filepath.Base(path)
	suffix := ""
	if idx := strings.LastIndex(fileName, "."); idx != -1 {
		suffix = fileName[idx+1:]
	}
	return &File{
		Path:     path,
		FileName: fileName,
		Suffix:   suffix,
	}
}

// CreateFileWithName creates a File instance with a custom filename
func CreateFileWithName(path string, fileName string) *File {
	suffix := ""
	if idx := strings.LastIndex(fileName, "."); idx != -1 {
		suffix = fileName[idx+1:]
	}
	return &File{
		Path:     path,
		FileName: fileName,
		Suffix:   suffix,
	}
}
