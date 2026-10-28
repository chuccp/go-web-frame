// Package web: file download and static file system response types.
package web

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// EmbedFileSystem describes a file tree that can be served either from an
// embedded filesystem or from a directory on disk. Name is the tree's root,
// relative to fsys when embedded and to the working directory when reloading.
type EmbedFileSystem struct {
	reload     bool
	name       string
	reloadPath string
	fsys       fs.FS
}

// Reload switches the file source at each request. With reloading enabled, Name
// or ReloadPath is opened as a directory on disk so edits are visible without
// rebuilding.
func (e *EmbedFileSystem) Reload(reload bool) *EmbedFileSystem {
	e.reload = reload
	return e
}

// ReloadEnabled reports whether the filesystem currently reloads from disk.
func (e *EmbedFileSystem) ReloadEnabled() bool {
	return e != nil && e.reload
}

// ReloadPath overrides the disk directory used when Reload(true) is enabled.
// It is useful when the embedded root differs from the source-tree directory.
func (e *EmbedFileSystem) ReloadPath(path string) *EmbedFileSystem {
	e.reloadPath = path
	return e
}

// EmbedFs creates a file tree rooted at name. An optional fsys supplies the
// embedded copy; without one, the tree is read from the working directory.
func EmbedFs(name string, fsys ...fs.FS) *EmbedFileSystem {
	e := &EmbedFileSystem{name: name}
	if len(fsys) > 0 {
		e.fsys = fsys[0]
	}
	return e
}

// FileSystem returns the active filesystem for the current reload setting.
func (e *EmbedFileSystem) FileSystem() (http.FileSystem, error) {
	active, err := e.activeFS()
	if err != nil {
		return nil, err
	}
	return http.FS(active), nil
}

// Open implements fs.FS using the active embedded or disk source.
func (e *EmbedFileSystem) Open(name string) (fs.File, error) {
	active, err := e.activeFS()
	if err != nil {
		return nil, err
	}
	return active.Open(name)
}

func (e *EmbedFileSystem) activeFS() (fs.FS, error) {
	if e == nil {
		return nil, fmt.Errorf("embedded filesystem is nil")
	}

	name := e.name
	if name == "" {
		name = "."
	}
	if e.reload {
		root := e.reloadPath
		if root == "" {
			root = name
		}
		return os.DirFS(root), nil
	}
	if e.fsys == nil {
		return os.DirFS(name), nil
	}
	if !fs.ValidPath(name) {
		return nil, fmt.Errorf("invalid embedded filesystem root %q", name)
	}
	name = path.Clean(name)
	if name == "." {
		return e.fsys, nil
	}
	sub, err := fs.Sub(e.fsys, name)
	if err != nil {
		return nil, fmt.Errorf("open embedded filesystem root %q: %w", name, err)
	}
	return sub, nil
}

// FileResponse represents a downloadable file response.
// Path is the file system path, FileName is the download name,
// and Suffix is an optional suffix appended to the filename.
type FileResponse struct {
	Path     string // File system path to the file
	FileName string // Name shown to the client on download
	Suffix   string // Optional suffix appended to the filename
}

// CreateFileResponse creates a downloadable file response from a path.
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

// CreateFileResponseWithName creates a downloadable file response with a custom filename.
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

// FileSystemResponse serves a file from an embedded or in-memory http.FileSystem.
type FileSystemResponse struct {
	Filepath string
	FS       http.FileSystem
}

// EmbedFileSystemResponse serves a request from an EmbedFileSystem.
type EmbedFileSystemResponse struct {
	filepath string
	fs       *EmbedFileSystem
}
