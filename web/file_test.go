package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestAddEmbedFsServesFilesFromEmbeddedRoot(t *testing.T) {
	embedded := fstest.MapFS{
		"static/app.css": {
			Data: []byte("body { color: black; }"),
		},
		"static/nested/data.txt": {
			Data: []byte("nested"),
		},
	}

	server := DefaultServer()
	server.AddEmbedFs("/assets", EmbedFs("static", embedded))
	handler := server.GetHandler()

	css := staticTestRequest(handler, "/assets/app.css")
	if css.Code != http.StatusOK {
		t.Fatalf("CSS status = %d, want %d", css.Code, http.StatusOK)
	}
	if !strings.Contains(css.Body.String(), "color: black") {
		t.Fatalf("CSS body = %q", css.Body.String())
	}

	nested := staticTestRequest(handler, "/assets/nested/data.txt")
	if nested.Code != http.StatusOK || nested.Body.String() != "nested" {
		t.Fatalf("nested response = %d %q", nested.Code, nested.Body.String())
	}
}

func TestEmbedFsRejectsInvalidEmbeddedRoot(t *testing.T) {
	embedded := EmbedFs("../static", fstest.MapFS{})
	if _, err := embedded.FileSystem(); err == nil {
		t.Fatal("FileSystem() accepted an invalid embedded root")
	}
}

func TestEmbedFsImplementsFS(t *testing.T) {
	embedded := fstest.MapFS{
		"static/app.css": {
			Data: []byte("embedded"),
		},
	}
	active := EmbedFs("static", embedded)
	data, err := fs.ReadFile(active, "app.css")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "embedded" {
		t.Fatalf("file = %q, want embedded", data)
	}
}

func TestEmbedFsReloadUsesDiskPath(t *testing.T) {
	embedded := fstest.MapFS{
		"static/app.css": {
			Data: []byte("embedded"),
		},
	}
	diskRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(diskRoot, "app.css"), []byte("disk"), 0o600); err != nil {
		t.Fatal(err)
	}

	server := DefaultServer()
	server.AddEmbedFs(
		"/assets",
		EmbedFs("static", embedded).Reload(true).ReloadPath(diskRoot),
	)
	handler := server.GetHandler()

	response := staticTestRequest(handler, "/assets/app.css")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "disk" {
		t.Fatalf("body = %q, want disk", response.Body.String())
	}

	if err := os.WriteFile(filepath.Join(diskRoot, "app.css"), []byte("updated"), 0o600); err != nil {
		t.Fatal(err)
	}
	updated := staticTestRequest(handler, "/assets/app.css")
	if updated.Body.String() != "updated" {
		t.Fatalf("reloaded body = %q, want updated", updated.Body.String())
	}
}

func staticTestRequest(handler http.Handler, target string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}
