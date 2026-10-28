package wf

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

//go:embed testdata/template-integration/templates
var integrationTemplatesFS embed.FS

type templateIntegrationData struct {
	Name string
}

type templateIntegrationRest struct {
	core.IRest
}

func (r *templateIntegrationRest) Init(ctx *core.Context) error {
	templateOption, err := web.LoadTemplateOption(
		integrationTemplatesFS,
		"testdata/template-integration/templates",
		web.WithLayouts("layout", "bare"),
		web.WithPartials("partials"),
	)
	if err != nil {
		return err
	}

	ctx.GetTemplate("/template-get", func(view *web.Template) error {
		view.View("page").Data(templateIntegrationData{Name: view.Request().Query("name")})
		return nil
	}, templateOption)
	ctx.PostTemplate("/template-post", func(view *web.Template) error {
		view.View("page").
			Layout("bare").
			Status(http.StatusCreated).
			Header("X-Template-Integration", "ok").
			Data(templateIntegrationData{Name: view.Request().GetFormParam("name")})
		return nil
	}, templateOption)
	return nil
}

func TestTemplateIntegrationThroughBuilder(t *testing.T) {
	app := NewBuilder(config2.NewConfig()).
		Rest(&templateIntegrationRest{}).
		Build()
	handler := app.GetHandler()
	if handler == nil {
		t.Fatal("GetHandler() returned nil")
	}

	getRecorder := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "http://localhost:19009/template-get?name=Alice", nil)
	handler.ServeHTTP(getRecorder, getRequest)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d: %s", getRecorder.Code, http.StatusOK, getRecorder.Body.String())
	}
	if body := getRecorder.Body.String(); !strings.Contains(body, `<header>Alice</header>`) ||
		!strings.Contains(body, `<p>Alice</p>`) {
		t.Errorf("GET body = %q", body)
	}

	form := url.Values{"name": {"Bob"}}.Encode()
	postRecorder := httptest.NewRecorder()
	postRequest := httptest.NewRequest(http.MethodPost, "http://localhost:19009/template-post", strings.NewReader(form))
	postRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.ServeHTTP(postRecorder, postRequest)
	if postRecorder.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d: %s", postRecorder.Code, http.StatusCreated, postRecorder.Body.String())
	}
	if got := postRecorder.Header().Get("X-Template-Integration"); got != "ok" {
		t.Errorf("X-Template-Integration = %q, want ok", got)
	}
	if body := postRecorder.Body.String(); !strings.Contains(body, `<main><header>Bob</header><p>Bob</p></main>`) {
		t.Errorf("POST body = %q", body)
	}
}
