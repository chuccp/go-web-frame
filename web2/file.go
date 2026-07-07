package web2

import (
	"net/http"
	"path/filepath"
	"strings"
)

// File represents a downloadable file response.
// Path is the file system path, FileName is the download name,
// and Suffix is an optional suffix appended to the filename.
type FileResponse struct {
	Path     string // File system path to the file
	FileName string // Name shown to the client on download
	Suffix   string // Optional suffix appended to the filename
}

// CreateFile creates a File from a path, auto-extracting the filename and suffix.
func CreateFileResponse(path string) *FileResponse {
	fileName := filepath.Base(path)
	suffix := ""
	if idx := strings.LastIndex(fileName, "."); idx != -1 {
		suffix = fileName[idx+1:]
	}
	return &FileResponse{
		Path:     path,
		FileName: fileName,
		Suffix:   suffix,
	}
}

// CreateFileWithName creates a File from a path with a custom download filename.
func CreateFileResponseWithName(path string, fileName string) *FileResponse {
	suffix := ""
	if idx := strings.LastIndex(fileName, "."); idx != -1 {
		suffix = fileName[idx+1:]
	}
	return &FileResponse{
		Path:     path,
		FileName: fileName,
		Suffix:   suffix,
	}
}

type FileSystemResponse struct {
	Filepath string
	FS       http.FileSystem
}
