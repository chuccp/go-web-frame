package web

import (
	"path/filepath"
	"strings"
)

// File represents a downloadable file response.
// Path is the file system path, FileName is the download name,
// and Suffix is an optional suffix appended to the filename.
type File struct {
	Path     string // File system path to the file
	FileName string // Name shown to the client on download
	Suffix   string // Optional suffix appended to the filename
}

// CreateFile creates a File from a path, auto-extracting the filename and suffix.
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

// CreateFileWithName creates a File from a path with a custom download filename.
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
