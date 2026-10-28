package web

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

type templateTestData struct {
	Title string
}

func templateTestFS() fstest.MapFS {
	return fstest.MapFS{
		"templates/layout.html": {
			Data: []byte(`{{define "layout"}}<html><body class="full">{{template "content" .}}</body></html>{{end}}`),
		},
		"templates/bare.html": {
			Data: []byte(`{{define "layout"}}<main class="bare">{{template "content" .}}</main>{{end}}`),
		},
		"templates/partials.html": {
			Data: []byte(`{{define "header"}}<header>{{.Title}}</header>{{end}}`),
		},
		"templates/one.html": {
			Data: []byte(`{{define "content"}}{{template "header" .}}<h1>{{.Title}}</h1>{{end}}`),
		},
		"templates/two.html": {
			Data: []byte(`{{define "content"}}<h2>{{.Title}}</h2>{{end}}`),
		},
	}
}

func loadTemplateTestSet(t *testing.T, fsys fstest.MapFS, options ...TemplateOption) *TemplateSet {
	t.Helper()
	base := []TemplateOption{
		WithLayouts("layout", "bare"),
		WithPartials("partials"),
	}
	base = append(base, options...)
	set, err := LoadTemplates(fsys, "templates", base...)
	if err != nil {
		t.Fatalf("LoadTemplates() error = %v", err)
	}
	return set
}

func registerTemplate(server *Server, set *TemplateSet, method string, relativePath string, handler TemplateHandler) *Route {
	return server.Template(method, relativePath, handler, WithTemplateSet(set))
}

func templateTestRequest(handler http.Handler, method string, target string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, target, nil))
	return recorder
}

func TestTemplateRouteRendersLayoutAndPartials(t *testing.T) {
	server := DefaultServer()
	set := loadTemplateTestSet(t, templateTestFS())
	registerTemplate(server, set, http.MethodGet, "/one", func(view *Template) error {
		view.View("one").Data(templateTestData{Title: "Dashboard"})
		return nil
	})

	recorder := templateTestRequest(server.GetHandler(), http.MethodGet, "/one")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", contentType)
	}
	body := recorder.Body.String()
	for _, expected := range []string{`class="full"`, `<header>Dashboard</header>`, `<h1>Dashboard</h1>`} {
		if !strings.Contains(body, expected) {
			t.Errorf("body does not contain %q: %s", expected, body)
		}
	}
}

func TestTemplateMethodRouteHelpers(t *testing.T) {
	server := DefaultServer()
	set := loadTemplateTestSet(t, templateTestFS())

	routes := []struct {
		method   string
		register func(*Server, string, TemplateHandler, ...TemplateOption) *Route
	}{
		{method: http.MethodGet, register: func(server *Server, path string, handler TemplateHandler, options ...TemplateOption) *Route {
			return server.GetTemplate(path, handler, options...)
		}},
		{method: http.MethodPost, register: func(server *Server, path string, handler TemplateHandler, options ...TemplateOption) *Route {
			return server.PostTemplate(path, handler, options...)
		}},
		{method: http.MethodPut, register: func(server *Server, path string, handler TemplateHandler, options ...TemplateOption) *Route {
			return server.PutTemplate(path, handler, options...)
		}},
		{method: http.MethodDelete, register: func(server *Server, path string, handler TemplateHandler, options ...TemplateOption) *Route {
			return server.DeleteTemplate(path, handler, options...)
		}},
		{method: http.MethodPatch, register: func(server *Server, path string, handler TemplateHandler, options ...TemplateOption) *Route {
			return server.PatchTemplate(path, handler, options...)
		}},
	}
	for _, route := range routes {
		method := route.method
		route.register(server, "/"+strings.ToLower(method), func(view *Template) error {
			view.View("two").
				Layout("bare").
				Data(templateTestData{Title: method})
			return nil
		}, WithTemplateSet(set))
	}
	handler := server.GetHandler()

	for _, route := range routes {
		recorder := templateTestRequest(handler, route.method, "/"+strings.ToLower(route.method))
		if recorder.Code != http.StatusOK {
			t.Errorf("%s status = %d, want %d", route.method, recorder.Code, http.StatusOK)
			continue
		}
		if body := recorder.Body.String(); !strings.Contains(body, "<h2>"+route.method+"</h2>") {
			t.Errorf("%s rendered unexpected body: %s", route.method, body)
		}
	}
}

func TestTemplatePagesAndLayoutsAreIsolated(t *testing.T) {
	server := DefaultServer()
	set := loadTemplateTestSet(t, templateTestFS())
	registerTemplate(server, set, http.MethodGet, "/one", func(view *Template) error {
		view.View("one").Data(templateTestData{Title: "Page one"})
		return nil
	})
	registerTemplate(server, set, http.MethodGet, "/two", func(view *Template) error {
		view.View("two").Layout("bare").Data(templateTestData{Title: "Page two"})
		return nil
	})
	handler := server.GetHandler()

	one := templateTestRequest(handler, http.MethodGet, "/one")
	if body := one.Body.String(); !strings.Contains(body, `<h1>Page one</h1>`) || strings.Contains(body, "Page two") {
		t.Errorf("page one rendered unexpected content: %s", body)
	}

	two := templateTestRequest(handler, http.MethodGet, "/two")
	if body := two.Body.String(); !strings.Contains(body, `<main class="bare"><h2>Page two</h2></main>`) || strings.Contains(body, "Page one") {
		t.Errorf("page two rendered unexpected content: %s", body)
	}
}

func TestTemplateHandlerConfiguresPage(t *testing.T) {
	server := DefaultServer()
	set := loadTemplateTestSet(t, templateTestFS())
	calls := 0
	registerTemplate(server, set, http.MethodGet, "/configured", func(view *Template) error {
		calls++
		view.View("two").Data(templateTestData{Title: "Configured"})
		return nil
	})

	recorder := templateTestRequest(server.GetHandler(), http.MethodGet, "/configured")
	if calls != 1 {
		t.Errorf("handler calls = %d, want 1", calls)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `<h2>Configured</h2>`) {
		t.Errorf("configured page was not rendered: %s", body)
	}
}

func TestTemplateOptionOverridesLayoutAndFunctions(t *testing.T) {
	views := templateTestFS()
	views["templates/option.html"] = &fstest.MapFile{
		Data: []byte(`{{define "content"}}<p>{{upper .Title}}</p>{{end}}`),
	}
	server := DefaultServer()
	set := loadTemplateTestSet(t, views, WithTemplateFuncs(template.FuncMap{
		"upper": func(value string) string { return value },
	}))
	registerTemplate(server, set, http.MethodGet, "/option", func(view *Template) error {
		view.View("option").
			Option(
				WithLayouts("bare", "layout"),
				WithTemplateFuncs(template.FuncMap{"upper": strings.ToUpper}),
			).
			Data(templateTestData{Title: "option"})
		return nil
	})

	recorder := templateTestRequest(server.GetHandler(), http.MethodGet, "/option")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `<main class="bare"><p>OPTION</p></main>`) {
		t.Errorf("runtime template options were not applied: %s", body)
	}
}

func TestTemplateHandlerErrorStopsRendering(t *testing.T) {
	server := DefaultServer()
	set := loadTemplateTestSet(t, templateTestFS())
	registerTemplate(server, set, http.MethodGet, "/error", func(view *Template) error {
		view.View("one")
		return errors.New("prepare failed")
	})

	recorder := templateTestRequest(server.GetHandler(), http.MethodGet, "/error")
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if strings.Contains(recorder.Body.String(), `<html>`) {
		t.Errorf("template was rendered after handler error: %s", recorder.Body.String())
	}
}

func TestTemplateStatusHeaderAndRedirect(t *testing.T) {
	server := DefaultServer()
	set := loadTemplateTestSet(t, templateTestFS())
	registerTemplate(server, set, http.MethodGet, "/status", func(view *Template) error {
		view.View("one").
			Data(templateTestData{Title: "Unauthorized"}).
			Status(http.StatusUnauthorized).
			Header("X-Template", "rendered")
		return nil
	})
	registerTemplate(server, set, http.MethodGet, "/redirect", func(view *Template) error {
		view.Redirect(http.StatusFound, "/login")
		return nil
	})
	handler := server.GetHandler()

	status := templateTestRequest(handler, http.MethodGet, "/status")
	if status.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", status.Code, http.StatusUnauthorized)
	}
	if got := status.Header().Get("X-Template"); got != "rendered" {
		t.Errorf("X-Template = %q, want rendered", got)
	}

	redirect := templateTestRequest(handler, http.MethodGet, "/redirect")
	if redirect.Code != http.StatusFound {
		t.Errorf("redirect status = %d, want %d", redirect.Code, http.StatusFound)
	}
	if got := redirect.Header().Get("Location"); got != "/login" {
		t.Errorf("Location = %q, want /login", got)
	}
	// Gin may include a small HTML anchor in the redirect body; the page
	// template itself must not have been rendered.
	if strings.Contains(redirect.Body.String(), `<html`) {
		t.Errorf("redirect rendered a page: %s", redirect.Body.String())
	}
}

func TestTemplateEscapesHTMLAndRegistersFunctions(t *testing.T) {
	views := templateTestFS()
	views["templates/escaped.html"] = &fstest.MapFile{
		Data: []byte(`{{define "content"}}<p>{{upper .Title}}</p>{{end}}`),
	}
	server := DefaultServer()
	set := loadTemplateTestSet(
		t,
		views,
		WithTemplateFuncs(template.FuncMap{"upper": strings.ToUpper}),
	)
	registerTemplate(server, set, http.MethodGet, "/escaped", func(view *Template) error {
		view.View("escaped").Layout("bare").Data(templateTestData{Title: `<script>alert(1)</script>`})
		return nil
	})

	recorder := templateTestRequest(server.GetHandler(), http.MethodGet, "/escaped")
	body := recorder.Body.String()
	if !strings.Contains(body, `&lt;SCRIPT&gt;ALERT(1)&lt;/SCRIPT&gt;`) {
		t.Errorf("template output was not escaped: %s", body)
	}
	if strings.Contains(body, `<script>`) {
		t.Errorf("unescaped script found in body: %s", body)
	}
}

func TestTemplateFailuresDoNotWritePartialHTML(t *testing.T) {
	server := DefaultServer()
	views := templateTestFS()
	views["templates/broken.html"] = &fstest.MapFile{
		Data: []byte(`{{define "content"}}SHOULD_NOT_LEAK{{.Missing}}{{end}}`),
	}
	set := loadTemplateTestSet(t, views)
	registerTemplate(server, set, http.MethodGet, "/broken", func(view *Template) error {
		view.View("broken").Data(templateTestData{})
		return nil
	})

	recorder := templateTestRequest(server.GetHandler(), http.MethodGet, "/broken")
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if strings.Contains(recorder.Body.String(), "SHOULD_NOT_LEAK") {
		t.Errorf("partial template output was written: %s", recorder.Body.String())
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want JSON error response", contentType)
	}
}

func TestTemplateMissingConfigurationAndViewReturnServerError(t *testing.T) {
	tests := []struct {
		name      string
		configure bool
		view      string
	}{
		{name: "templates not configured", view: "one"},
		{name: "view not registered", configure: true, view: "missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := DefaultServer()
			var options []TemplateOption
			if tt.configure {
				options = append(options, WithTemplateSet(loadTemplateTestSet(t, templateTestFS())))
			}
			server.Template(http.MethodGet, "/failure", func(view *Template) error {
				view.View(tt.view)
				return nil
			}, options...)

			recorder := templateTestRequest(server.GetHandler(), http.MethodGet, "/failure")
			if recorder.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
			}
		})
	}
}

func TestTemplateDevReloadReadsCurrentFiles(t *testing.T) {
	views := templateTestFS()
	views["templates/reload.html"] = &fstest.MapFile{
		Data: []byte(`{{define "content"}}<p>before</p>{{end}}`),
	}
	server := DefaultServer()
	set := loadTemplateTestSet(t, views, WithDevReload(true))
	registerTemplate(server, set, http.MethodGet, "/reload", func(view *Template) error {
		view.View("reload").Layout("bare")
		return nil
	})
	handler := server.GetHandler()

	before := templateTestRequest(handler, http.MethodGet, "/reload")
	if !strings.Contains(before.Body.String(), "before") {
		t.Fatalf("initial body = %q, want before", before.Body.String())
	}

	views["templates/reload.html"] = &fstest.MapFile{
		Data: []byte(`{{define "content"}}<p>after</p>{{end}}`),
	}
	after := templateTestRequest(handler, http.MethodGet, "/reload")
	if !strings.Contains(after.Body.String(), "after") || strings.Contains(after.Body.String(), "before") {
		t.Errorf("reloaded body = %q, want only after", after.Body.String())
	}
}

func TestLoadTemplatesUsesEmbedFileSystemReloadState(t *testing.T) {
	views := templateTestFS()
	templateFS := EmbedFs("templates", views).
		Reload(true).
		ReloadPath("testdata/template-embed/templates")
	set, err := LoadTemplates(
		templateFS,
		".",
		WithLayouts("layout"),
		WithPartials("partials"),
	)
	if err != nil {
		t.Fatalf("LoadTemplates() error = %v", err)
	}
	if !set.defaultReload {
		t.Fatal("LoadTemplates() ignored the filesystem reload state")
	}
}

func TestTemplateReloadOverridesSetDefault(t *testing.T) {
	views := templateTestFS()
	views["templates/reload.html"] = &fstest.MapFile{
		Data: []byte(`{{define "content"}}<p>before</p>{{end}}`),
	}
	server := DefaultServer()
	set := loadTemplateTestSet(t, views, WithDevReload(true))
	registerTemplate(server, set, http.MethodGet, "/reload", func(view *Template) error {
		view.View("reload").Layout("bare")
		if view.Request().Query("reload") != "1" {
			view.Reload(false)
		}
		return nil
	})
	handler := server.GetHandler()

	cachedBefore := templateTestRequest(handler, http.MethodGet, "/reload")
	if !strings.Contains(cachedBefore.Body.String(), "before") {
		t.Fatalf("initial body = %q, want before", cachedBefore.Body.String())
	}

	views["templates/reload.html"] = &fstest.MapFile{
		Data: []byte(`{{define "content"}}<p>after</p>{{end}}`),
	}
	cachedAfter := templateTestRequest(handler, http.MethodGet, "/reload")
	if !strings.Contains(cachedAfter.Body.String(), "before") {
		t.Errorf("Reload(false) body = %q, want cached before", cachedAfter.Body.String())
	}
	reloaded := templateTestRequest(handler, http.MethodGet, "/reload?reload=1")
	if !strings.Contains(reloaded.Body.String(), "after") || strings.Contains(reloaded.Body.String(), "before") {
		t.Errorf("Reload(true) body = %q, want only after", reloaded.Body.String())
	}
}

func TestTemplateReloadCanBeEnabledPerRequest(t *testing.T) {
	views := templateTestFS()
	views["templates/reload.html"] = &fstest.MapFile{
		Data: []byte(`{{define "content"}}<p>before</p>{{end}}`),
	}
	server := DefaultServer()
	set := loadTemplateTestSet(t, views)
	registerTemplate(server, set, http.MethodGet, "/reload", func(view *Template) error {
		view.View("reload").Layout("bare").Reload(view.Request().Query("reload") == "1")
		return nil
	})
	handler := server.GetHandler()

	cached := templateTestRequest(handler, http.MethodGet, "/reload")
	if !strings.Contains(cached.Body.String(), "before") {
		t.Fatalf("initial body = %q, want before", cached.Body.String())
	}

	views["templates/reload.html"] = &fstest.MapFile{
		Data: []byte(`{{define "content"}}<p>after</p>{{end}}`),
	}
	stillCached := templateTestRequest(handler, http.MethodGet, "/reload")
	if !strings.Contains(stillCached.Body.String(), "before") {
		t.Errorf("cached body = %q, want before", stillCached.Body.String())
	}
	reloaded := templateTestRequest(handler, http.MethodGet, "/reload?reload=1")
	if !strings.Contains(reloaded.Body.String(), "after") {
		t.Errorf("reloaded body = %q, want after", reloaded.Body.String())
	}
}

func TestLoadTemplatesRejectsInvalidTemplates(t *testing.T) {
	views := templateTestFS()
	views["templates/invalid.html"] = &fstest.MapFile{
		Data: []byte(`{{define "content"}}{{if}}{{end}}`),
	}
	if _, err := LoadTemplates(
		views,
		"templates",
		WithLayouts("layout", "bare"),
		WithPartials("partials"),
	); err == nil {
		t.Fatal("LoadTemplates() error = nil, want parse error")
	}

	if _, err := LoadTemplates(views, "../templates"); err == nil {
		t.Fatal("LoadTemplates() error = nil, want invalid root error")
	}
}

func TestTemplateSetConcurrentRendering(t *testing.T) {
	server := DefaultServer()
	set := loadTemplateTestSet(t, templateTestFS())
	registerTemplate(server, set, http.MethodGet, "/concurrent", func(view *Template) error {
		view.View("one").Data(templateTestData{Title: view.Request().Query("title")})
		return nil
	})
	handler := server.GetHandler()

	const requests = 32
	var wait sync.WaitGroup
	errorsCh := make(chan string, requests)
	for i := 0; i < requests; i++ {
		title := "request"
		wait.Add(1)
		go func() {
			defer wait.Done()
			recorder := templateTestRequest(handler, http.MethodGet, "/concurrent?title="+title)
			if recorder.Code != http.StatusOK {
				errorsCh <- recorder.Body.String()
				return
			}
			if !strings.Contains(recorder.Body.String(), "<h1>"+title+"</h1>") {
				errorsCh <- recorder.Body.String()
			}
		}()
	}
	wait.Wait()
	close(errorsCh)
	for response := range errorsCh {
		t.Errorf("concurrent template response failed: %s", response)
	}
}
