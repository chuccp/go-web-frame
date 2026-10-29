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

// Reload overrides the template option's default reload behavior for this response.
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

func (t *Template) resolveOptions() (*templateOptions, error) {
	config := templateOptions{}
	for _, option := range t.options {
		if option != nil {
			option(&config)
		}
	}

	options, err := config.resolve()
	if err != nil {
		return nil, err
	}
	if t.layout == "" {
		t.layout = options.defaultLayout
	}
	if !t.reloadSet {
		t.reload = options.defaultReload
	}
	return options, nil
}

func (t *Template) render() ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("template response is nil")
	}
	options, err := t.resolveOptions()
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

	tmpl, err := options.templateFor(t.view, t.layout, t.reload)
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
	embedFileSystem *EmbedFileSystem
	root            string
	funcs           template.FuncMap
	layouts         []string
	partials        []string
	funcsSet        bool
	layoutsSet      bool
	partialsSet     bool
	reloadSet       bool
	defaultReload   bool
	defaultLayout   string
	layoutSet       map[string]struct{}
	viewSet         map[string]struct{}
	byKey           map[templateKey]*template.Template
}

// TemplateOption configures the filesystem, parse options, or compiled cache
// used by a Template response.
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

// WithEmbedFileSystem selects the embedded or development filesystem used to
// load templates. Its current reload state becomes the default unless
// WithReload is applied later.
func WithEmbedFileSystem(embedFileSystem *EmbedFileSystem) TemplateOption {
	return func(options *templateOptions) {
		options.embedFileSystem = embedFileSystem
		if !options.reloadSet {
			options.defaultReload = embedFileSystem.ReloadEnabled()
		}
		options.layoutSet = nil
		options.viewSet = nil
		options.byKey = nil
	}
}

func withTemplateOptions(compiled *templateOptions) TemplateOption {
	return func(options *templateOptions) {
		previous := *options
		clone := compiled.clone()
		*options = *clone

		if previous.funcsSet {
			options.funcsSet = true
			for name, fn := range previous.funcs {
				options.funcs[name] = fn
			}
		}
		if previous.layoutsSet {
			options.layoutsSet = true
			options.layouts = previous.layouts
		}
		if previous.partialsSet {
			options.partialsSet = true
			options.partials = previous.partials
		}
		if previous.reloadSet {
			options.reloadSet = true
			options.defaultReload = previous.defaultReload
		}
	}
}

// WithReload sets the default reload behavior for all Template responses. A
// handler can override it per request with Template.Reload. Reloading is intended
// for development with an os.DirFS and should be disabled in production.
func WithReload(enabled bool) TemplateOption {
	return func(options *templateOptions) {
		options.reloadSet = true
		options.defaultReload = enabled
	}
}

// loadTemplates parses every page together with each registered layout and the
// configured partials into templateOptions. Logical names omit the .html
// extension.
func loadTemplates(embedFileSystem *EmbedFileSystem, root string, options ...TemplateOption) (*templateOptions, error) {
	config := templateOptions{
		root:    root,
		funcs:   make(template.FuncMap),
		layouts: []string{"layout"},
	}
	WithEmbedFileSystem(embedFileSystem)(&config)
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	return config.compile()
}

func (o *templateOptions) clone() *templateOptions {
	if o == nil {
		return &templateOptions{}
	}
	clone := *o
	clone.funcs = cloneTemplateFuncs(o.funcs)
	if o.funcs != nil && clone.funcs == nil {
		clone.funcs = make(template.FuncMap)
	}
	clone.layouts = append([]string(nil), o.layouts...)
	clone.partials = append([]string(nil), o.partials...)
	return &clone
}

func (o *templateOptions) resolve() (*templateOptions, error) {
	if o == nil || o.embedFileSystem == nil {
		return nil, fmt.Errorf("templates are not configured")
	}
	if o.byKey != nil && !o.funcsSet && !o.layoutsSet && !o.partialsSet {
		return o, nil
	}
	return o.compile()
}

func (o *templateOptions) compile() (*templateOptions, error) {
	if o == nil || o.embedFileSystem == nil {
		return nil, fmt.Errorf("embedded template filesystem is nil")
	}

	root, err := normalizeTemplateRoot(o.root)
	if err != nil {
		return nil, err
	}
	if len(o.layouts) == 0 {
		return nil, fmt.Errorf("at least one template layout is required")
	}

	layouts, err := normalizeTemplateNames(o.layouts)
	if err != nil {
		return nil, fmt.Errorf("invalid template layout: %w", err)
	}
	partials, err := normalizeTemplateNames(o.partials)
	if err != nil {
		return nil, fmt.Errorf("invalid template partial: %w", err)
	}
	defaultLayout := layouts[0]

	info, err := fs.Stat(o.embedFileSystem, root)
	if err != nil {
		return nil, fmt.Errorf("stat template root %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("template root %q is not a directory", root)
	}

	views, err := discoverTemplateViews(o.embedFileSystem, root, layouts, partials)
	if err != nil {
		return nil, err
	}
	if len(views) == 0 {
		return nil, fmt.Errorf("no page templates found under %q", root)
	}

	compiled := &templateOptions{
		embedFileSystem: o.embedFileSystem,
		root:            root,
		funcs:           cloneTemplateFuncs(o.funcs),
		layouts:         append([]string(nil), layouts...),
		defaultLayout:   defaultLayout,
		partials:        partials,
		defaultReload:   o.defaultReload,
		layoutSet:       make(map[string]struct{}, len(layouts)),
		viewSet:         make(map[string]struct{}, len(views)),
		byKey:           make(map[templateKey]*template.Template, len(views)*len(layouts)),
	}
	if compiled.funcs == nil {
		compiled.funcs = make(template.FuncMap)
	}
	for _, layout := range layouts {
		compiled.layoutSet[layout] = struct{}{}
	}
	for _, view := range views {
		compiled.viewSet[view] = struct{}{}
	}

	for _, view := range views {
		for _, layout := range layouts {
			tmpl, err := compiled.parse(view, layout)
			if err != nil {
				return nil, err
			}
			compiled.byKey[templateKey{view: view, layout: layout}] = tmpl
		}
	}
	return compiled, nil
}

// LoadTemplateOption compiles templates once and returns the compiled
// templateOptions as a reusable route option.
func LoadTemplateOption(embedFileSystem *EmbedFileSystem, root string, options ...TemplateOption) (TemplateOption, error) {
	compiled, err := loadTemplates(embedFileSystem, root, options...)
	if err != nil {
		return nil, err
	}
	return withTemplateOptions(compiled), nil
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

func (o *templateOptions) filePath(name string) string {
	file := name + templateExtension
	if o.root == "." {
		return file
	}
	return o.root + "/" + file
}

func (o *templateOptions) parse(view string, layout string) (*template.Template, error) {
	files := make([]string, 0, len(o.partials)+2)
	files = append(files, o.filePath(layout))
	for _, partial := range o.partials {
		files = append(files, o.filePath(partial))
	}
	files = append(files, o.filePath(view))

	tmpl, err := template.New(view+"@"+layout).
		Funcs(o.funcs).
		ParseFS(o.embedFileSystem, files...)
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

func (o *templateOptions) templateFor(view string, layout string, reload bool) (*template.Template, error) {
	if _, ok := o.viewSet[view]; !ok {
		return nil, fmt.Errorf("template page %q is not registered", view)
	}
	if _, ok := o.layoutSet[layout]; !ok {
		return nil, fmt.Errorf("template layout %q is not registered", layout)
	}
	if reload {
		return o.parse(view, layout)
	}
	tmpl, ok := o.byKey[templateKey{view: view, layout: layout}]
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
