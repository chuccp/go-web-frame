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
	options      []TemplateOption
	view         string
	layout       string
	data         any
	status       int
	headers      http.Header
	redirectCode int
	redirectTo   string
	reload       bool
	reloadSet    bool
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

// Option adds runtime template options to this response. Options are applied
// after the handler returns and before the template is rendered.
func (t *Template) Option(options ...TemplateOption) *Template {
	t.options = append(t.options, options...)
	return t
}

// Reload overrides the template set's default reload behavior for this response.
// When enabled, the selected page, layout and partials are parsed from the source
// filesystem again instead of using the compiled cache.
func (t *Template) Reload(reload bool) *Template {
	t.reload = reload
	t.reloadSet = true
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
	return t
}

func (t *Template) resolveOptions() (*TemplateSet, error) {
	config := templateOptions{}
	for _, option := range t.options {
		if option != nil {
			option(&config)
		}
	}

	set := config.templateSet
	if set == nil {
		return nil, fmt.Errorf("templates are not configured")
	}
	if config.funcsSet || config.layoutsSet || config.partialsSet {
		var err error
		set, err = set.withOptions(config)
		if err != nil {
			return nil, err
		}
	}
	if t.layout == "" {
		t.layout = set.defaultLayout
	}
	if !t.reloadSet {
		if config.reloadSet {
			t.reload = config.defaultReload
		} else if config.templateSet != nil {
			t.reload = set.defaultReload
		}
	}
	return set, nil
}

func (t *Template) render() ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("template response is nil")
	}
	set, err := t.resolveOptions()
	if err != nil {
		return nil, err
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

	tmpl, err := set.templateFor(t.view, t.layout, t.reload)
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
	funcs         template.FuncMap
	layouts       []string
	partials      []string
	templateSet   *TemplateSet
	funcsSet      bool
	layoutsSet    bool
	partialsSet   bool
	reloadSet     bool
	defaultReload bool
}

// TemplateOption configures a TemplateSet before templates are parsed.
type TemplateOption func(*templateOptions)

// WithTemplateFuncs registers functions used during template parsing. Multiple
// calls are merged; later functions replace earlier ones with the same name.
func WithTemplateFuncs(funcs template.FuncMap) TemplateOption {
	return func(options *templateOptions) {
		options.funcsSet = true
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
		options.layoutsSet = true
		options.layouts = append([]string{defaultLayout}, others...)
	}
}

// WithPartials registers files that are parsed with every page and layout.
func WithPartials(names ...string) TemplateOption {
	return func(options *templateOptions) {
		options.partialsSet = true
		options.partials = append([]string(nil), names...)
	}
}

// WithTemplateSet selects an already compiled template set for a response.
func WithTemplateSet(set *TemplateSet) TemplateOption {
	return func(options *templateOptions) {
		options.templateSet = set
	}
}

// WithDevReload sets the default reload behavior for all Template responses. A
// handler can override it per request with Template.Reload. Reloading is intended
// for development with an os.DirFS and should be disabled in production.
func WithDevReload(enabled bool) TemplateOption {
	return func(options *templateOptions) {
		options.reloadSet = true
		options.defaultReload = enabled
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
	layoutNames   []string
	defaultLayout string
	partials      []string
	views         map[string]struct{}
	byKey         map[templateKey]*template.Template
	defaultReload bool
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
	if reloadFS, ok := fsys.(interface{ ReloadEnabled() bool }); ok {
		config.defaultReload = reloadFS.ReloadEnabled()
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
		layoutNames:   append([]string(nil), layouts...),
		defaultLayout: defaultLayout,
		partials:      partials,
		views:         make(map[string]struct{}, len(views)),
		byKey:         make(map[templateKey]*template.Template, len(views)*len(layouts)),
		defaultReload: config.defaultReload,
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

func (s *TemplateSet) withOptions(config templateOptions) (*TemplateSet, error) {
	funcs := cloneTemplateFuncs(s.funcs)
	if funcs == nil {
		funcs = make(template.FuncMap)
	}
	for name, fn := range config.funcs {
		funcs[name] = fn
	}
	layouts := append([]string(nil), s.layoutNames...)
	if config.layoutsSet {
		layouts = append([]string(nil), config.layouts...)
	}
	if len(layouts) == 0 {
		return nil, fmt.Errorf("at least one template layout is required")
	}
	partials := append([]string(nil), s.partials...)
	if config.partialsSet {
		partials = append([]string(nil), config.partials...)
	}
	reload := s.defaultReload
	if config.reloadSet {
		reload = config.defaultReload
	}
	return LoadTemplates(
		s.fsys,
		s.root,
		WithTemplateFuncs(funcs),
		WithLayouts(layouts[0], layouts[1:]...),
		WithPartials(partials...),
		WithDevReload(reload),
	)
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

func (s *TemplateSet) templateFor(view string, layout string, reload bool) (*template.Template, error) {
	if _, ok := s.views[view]; !ok {
		return nil, fmt.Errorf("template page %q is not registered", view)
	}
	if _, ok := s.layouts[layout]; !ok {
		return nil, fmt.Errorf("template layout %q is not registered", layout)
	}
	if reload {
		return s.parse(view, layout)
	}
	tmpl, ok := s.byKey[templateKey{view: view, layout: layout}]
	if !ok {
		return nil, fmt.Errorf("template %q with layout %q is not compiled", view, layout)
	}
	return tmpl, nil
}

// TemplateResponse defers rendering to the response converter so template
// preparation errors and rendering errors use the normal response pipeline.
type TemplateResponse struct {
	Handler TemplateHandler
	Options []TemplateOption
}
