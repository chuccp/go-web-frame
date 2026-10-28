package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

const templateExtension = ".html"

// TemplateHandler prepares one HTML page. It is not middleware and does not
// control a chain; use a Filter for authentication, logging, and other
// cross-cutting work.
type TemplateHandler func(view *Template) error

// Template holds the per-request state of an HTML page. It is converted into an
// HTML response after its TemplateHandler has prepared the page.
type Template struct {
	request      *Request
	templateSet  *TemplateSet
	view         string
	layout       string
	data         any
	status       int
	headers      http.Header
	redirectCode int
	redirectTo   string
	devReload    bool
}

// Request returns the HTTP request associated with this page.
func (t *Template) Request() *Request {
	return t.request
}

// View selects a logical page name, without its .html extension.
func (t *Template) View(name string) *Template {
	t.view = name
	return t
}
func (t *Template) DevReload(devReload bool) *Template {
	t.devReload = devReload
	return t
}

// Layout selects a logical layout name, without its .html extension.
func (t *Template) Layout(name string) *Template {
	t.layout = name
	return t
}

// Data sets the value passed to the template. A later call replaces earlier data.
func (t *Template) Data(data any) *Template {
	t.data = data
	return t
}

// Status sets the final HTTP status code. The default is 200.
func (t *Template) Status(code int) *Template {
	t.status = code
	return t
}

// Header sets a response header written after successful rendering.
func (t *Template) Header(key string, value string) *Template {
	if t.headers == nil {
		t.headers = make(http.Header)
	}
	t.headers.Set(key, value)
	return t
}

// Redirect changes the response to a redirect. The page is not rendered.
func (t *Template) Redirect(code int, location string) *Template {
	t.redirectCode = code
	t.redirectTo = location
	return t
}

func newTemplate(request *Request) *Template {
	t := &Template{
		request: request,
		status:  http.StatusOK,
		headers: make(http.Header),
	}
	if request != nil {
		t.templateSet = request.templateSet
		if t.templateSet != nil {
			t.layout = t.templateSet.defaultLayout
		}
	}
	return t
}

func (t *Template) render() ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("template response is nil")
	}
	if t.templateSet == nil {
		return nil, fmt.Errorf("templates are not configured")
	}
	if t.view == "" {
		return nil, fmt.Errorf("template view is not set")
	}
	if t.layout == "" {
		return nil, fmt.Errorf("template layout is not set")
	}
	if t.status < http.StatusOK || t.status > 599 {
		return nil, fmt.Errorf("invalid template status code %d", t.status)
	}

	tmpl, err := t.templateSet.templateFor(t.view, t.layout)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", t.data); err != nil {
		return nil, fmt.Errorf("execute template %q with layout %q: %w", t.view, t.layout, err)
	}
	return buf.Bytes(), nil
}

type templateKey struct {
	view   string
	layout string
}

type templateOptions struct {
	funcs     template.FuncMap
	layouts   []string
	partials  []string
	devReload bool
}

// TemplateOption configures a TemplateSet before templates are parsed.
type TemplateOption func(*templateOptions)

// WithTemplateFuncs registers functions used during template parsing. Multiple
// calls are merged; later functions replace earlier ones with the same name.
func WithTemplateFuncs(funcs template.FuncMap) TemplateOption {
	return func(options *templateOptions) {
		if options.funcs == nil {
			options.funcs = make(template.FuncMap)
		}
		for name, fn := range funcs {
			options.funcs[name] = fn
		}
	}
}

// WithLayouts registers the available layouts. The first layout is the default.
func WithLayouts(defaultLayout string, others ...string) TemplateOption {
	return func(options *templateOptions) {
		options.layouts = append([]string{defaultLayout}, others...)
	}
}

// WithPartials registers files that are parsed with every page and layout.
func WithPartials(names ...string) TemplateOption {
	return func(options *templateOptions) {
		options.partials = append([]string(nil), names...)
	}
}

// WithDevReload reparses the selected page and layout on every render. It is
// intended for development with an os.DirFS and should be disabled in
// production.
func WithDevReload(enabled bool) TemplateOption {
	return func(options *templateOptions) {
		options.devReload = enabled
	}
}

// TemplateSet contains the compiled page/layout combinations loaded from an
// fs.FS. A TemplateSet is immutable after LoadTemplates returns and can be
// shared by multiple servers and requests.
type TemplateSet struct {
	fsys          fs.FS
	root          string
	funcs         template.FuncMap
	layouts       map[string]struct{}
	defaultLayout string
	partials      []string
	views         map[string]struct{}
	byKey         map[templateKey]*template.Template
	devReload     bool
}

// LoadTemplates parses every page together with each registered layout and the
// configured partials. Logical names omit the .html extension.
func LoadTemplates(fsys fs.FS, root string, options ...TemplateOption) (*TemplateSet, error) {
	if fsys == nil {
		return nil, fmt.Errorf("template filesystem is nil")
	}

	root, err := normalizeTemplateRoot(root)
	if err != nil {
		return nil, err
	}

	config := templateOptions{
		funcs:   make(template.FuncMap),
		layouts: []string{"layout"},
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	if len(config.layouts) == 0 {
		return nil, fmt.Errorf("at least one template layout is required")
	}

	layouts, err := normalizeTemplateNames(config.layouts)
	if err != nil {
		return nil, fmt.Errorf("invalid template layout: %w", err)
	}
	partials, err := normalizeTemplateNames(config.partials)
	if err != nil {
		return nil, fmt.Errorf("invalid template partial: %w", err)
	}
	defaultLayout := layouts[0]

	info, err := fs.Stat(fsys, root)
	if err != nil {
		return nil, fmt.Errorf("stat template root %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("template root %q is not a directory", root)
	}

	views, err := discoverTemplateViews(fsys, root, layouts, partials)
	if err != nil {
		return nil, err
	}
	if len(views) == 0 {
		return nil, fmt.Errorf("no page templates found under %q", root)
	}

	set := &TemplateSet{
		fsys:          fsys,
		root:          root,
		funcs:         cloneTemplateFuncs(config.funcs),
		layouts:       make(map[string]struct{}, len(layouts)),
		defaultLayout: defaultLayout,
		partials:      partials,
		views:         make(map[string]struct{}, len(views)),
		byKey:         make(map[templateKey]*template.Template, len(views)*len(layouts)),
		devReload:     config.devReload,
	}
	for _, layout := range layouts {
		set.layouts[layout] = struct{}{}
	}
	for _, view := range views {
		set.views[view] = struct{}{}
	}

	for _, view := range views {
		for _, layout := range layouts {
			tmpl, err := set.parse(view, layout)
			if err != nil {
				return nil, err
			}
			set.byKey[templateKey{view: view, layout: layout}] = tmpl
		}
	}
	return set, nil
}

func normalizeTemplateRoot(root string) (string, error) {
	if root == "" {
		return ".", nil
	}
	if !fs.ValidPath(root) {
		return "", fmt.Errorf("invalid template root %q", root)
	}
	return path.Clean(root), nil
}

func normalizeTemplateNames(names []string) ([]string, error) {
	result := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		clean, err := normalizeTemplateName(name)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		result = append(result, clean)
	}
	return result, nil
}

func normalizeTemplateName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("template name is empty")
	}
	if strings.HasSuffix(name, templateExtension) {
		return "", fmt.Errorf("template name %q must omit %s", name, templateExtension)
	}
	if !fs.ValidPath(name) || name == "." {
		return "", fmt.Errorf("invalid template name %q", name)
	}
	return path.Clean(name), nil
}

func discoverTemplateViews(
	fsys fs.FS,
	root string,
	layouts []string,
	partials []string,
) ([]string, error) {
	excluded := make(map[string]struct{}, len(layouts)+len(partials))
	for _, name := range layouts {
		excluded[name] = struct{}{}
	}
	for _, name := range partials {
		excluded[name] = struct{}{}
	}

	views := make([]string, 0)
	err := fs.WalkDir(fsys, root, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || path.Ext(filePath) != templateExtension {
			return nil
		}

		relative := filePath
		if root != "." {
			relative = strings.TrimPrefix(filePath, root+"/")
		}
		name := strings.TrimSuffix(relative, templateExtension)
		if _, skip := excluded[name]; skip {
			return nil
		}
		if _, err := normalizeTemplateName(name); err != nil {
			return err
		}
		views = append(views, name)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan templates under %q: %w", root, err)
	}
	return views, nil
}

func cloneTemplateFuncs(funcs template.FuncMap) template.FuncMap {
	if len(funcs) == 0 {
		return nil
	}
	result := make(template.FuncMap, len(funcs))
	for name, fn := range funcs {
		result[name] = fn
	}
	return result
}

func (s *TemplateSet) filePath(name string) string {
	file := name + templateExtension
	if s.root == "." {
		return file
	}
	return s.root + "/" + file
}

func (s *TemplateSet) parse(view string, layout string) (*template.Template, error) {
	files := make([]string, 0, len(s.partials)+2)
	files = append(files, s.filePath(layout))
	for _, partial := range s.partials {
		files = append(files, s.filePath(partial))
	}
	files = append(files, s.filePath(view))

	tmpl, err := template.New(view+"@"+layout).
		Funcs(s.funcs).
		ParseFS(s.fsys, files...)
	if err != nil {
		return nil, fmt.Errorf("parse template %q with layout %q: %w", view, layout, err)
	}
	if tmpl.Lookup("layout") == nil {
		return nil, fmt.Errorf("template layout %q does not define \"layout\"", layout)
	}
	if tmpl.Lookup("content") == nil {
		return nil, fmt.Errorf("template page %q does not define \"content\"", view)
	}
	return tmpl, nil
}

func (s *TemplateSet) templateFor(view string, layout string) (*template.Template, error) {
	if _, ok := s.views[view]; !ok {
		return nil, fmt.Errorf("template page %q is not registered", view)
	}
	if _, ok := s.layouts[layout]; !ok {
		return nil, fmt.Errorf("template layout %q is not registered", layout)
	}
	if s.devReload {
		return s.parse(view, layout)
	}
	tmpl, ok := s.byKey[templateKey{view: view, layout: layout}]
	if !ok {
		return nil, fmt.Errorf("template %q with layout %q is not compiled", view, layout)
	}
	return tmpl, nil
}

// SetTemplates installs the template set used by Template routes on this server.
// It should be called during application initialization, before serving requests.
func (server *Server) SetTemplates(set *TemplateSet) {
	server.templateSet = set
}

// TemplateResponse defers rendering to the response converter so template
// preparation errors and rendering errors use the normal response pipeline.
type TemplateResponse struct {
	Handler TemplateHandler
}
