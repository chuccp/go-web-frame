package web

import (
	"embed"
	"net/http"
	"strings"
	"testing"
)

//go:embed testdata/template-embed/templates
var embeddedTemplateFS embed.FS

func TestLoadTemplatesFromEmbedFS(t *testing.T) {
	set, err := LoadTemplates(
		embeddedTemplateFS,
		"testdata/template-embed/templates",
		WithPartials("partials"),
	)
	if err != nil {
		t.Fatalf("LoadTemplates() error = %v", err)
	}

	server := DefaultServer()
	registerTemplate(server, set, http.MethodGet, "/embedded", func(view *Template) error {
		view.View("embedded").Data(templateTestData{Title: "Embedded"})
		return nil
	})

	recorder := templateTestRequest(server.GetHandler(), http.MethodGet, "/embedded")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `<h1>Embedded</h1>`) {
		t.Errorf("embedded template was not rendered: %s", body)
	}
}
