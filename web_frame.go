package wf

import (
	"context"
	"net/http"
	"sync"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	db2 "github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetService retrieves a service of the specified type from the context.
// Deprecated
func GetService[T core.IService](c *core.Context) T {
	return core.GetService[T](c)
}

func MustService[T core.IService](c *core.Context) T {
	return core.MustService[T](c)
}

// GetModel retrieves a model of the specified type from the context.
// Deprecated
func GetModel[T core.IModel](c *core.Context) T {
	return core.GetModel[T](c)
}

func MustModel[T core.IModel](c *core.Context) T {
	return core.MustModel[T](c)
}

// GetReNewModel retrieves a model and creates a fresh instance with the given database connection.
// Deprecated
func GetReNewModel[T core.IModel](db *db2.DB, c *core.Context) T {
	return core.GetReNewModel[T](db, c)
}
func MustReNewModel[T core.IModel](db *db2.DB, c *core.Context) T {
	return core.MustReNewModel[T](db, c)
}

// GetRunner retrieves a runner of the specified type from the context.
// Deprecated
func GetRunner[T core.IRunner](c *core.Context) T {
	return core.GetRunner[T](c)
}
func MustRunner[T core.IRunner](c *core.Context) T {
	return core.MustRunner[T](c)
}

// GetFilter retrieves a filter of the specified type from the context.
// Deprecated
func GetFilter[T core.IFilter](c *core.Context) T {
	return core.GetFilter[T](c)
}
func MustFilter[T core.IFilter](c *core.Context) T {
	return core.MustFilter[T](c)
}

// UnmarshalKeyConfig unmarshals configuration under the given key into the specified type.
func UnmarshalKeyConfig[T any](key string, c *core.Context) (T, error) {
	return core.UnmarshalKeyConfig[T](key, c)
}

type DefaultRest struct {
	ctx *core.Context
}

func (receiver *DefaultRest) Init(ctx *core.Context) error {
	receiver.ctx = ctx
	return nil
}

// WebFrame is the main application struct that holds all components, services, models,
// REST groups, and configuration for a web application.
type WebFrame struct {
	mu            sync.Mutex // guards ctxCancelFunc and restart
	ctxCancelFunc context.CancelFunc
	restart       bool

	restGroups  []*core.RestGroup
	modelGroups []core.IModelGroup
	config      config.IConfig
	models      []core.IModel
	services    []core.IService
	rests       []core.IRest
	filters     []core.IFilter
	handles     *web.Handles
}

// Start initializes and runs the web application with a background context.
// Blocks until the application is shut down.
func (w *WebFrame) Start() error {
	return w.Run(context.Background())
}
func (w *WebFrame) GetHandler() http.Handler {
	server, _, err := w.init(context.Background())
	if err != nil {
		log.Error("GetHandler init failed", zap.Error(err))
		return nil
	}
	return server.GetHandler()
}

// Test initializes the application without starting a real HTTP server,
// allowing tests to run against the initialized context.
func (w *WebFrame) Test(f func(ctx *core.Context) error) error {
	_, ctx, err := w.init(context.Background())
	if err != nil {
		return err
	}
	return f(ctx)
}
func (w *WebFrame) init(ctx context.Context) (*core.Server, *core.Context, error) {

	gin.SetMode(gin.ReleaseMode)
	defaultServerConfig := web.DefaultServerConfig()
	if w.config.HasKey(web.ServerConfigKey) {
		err := w.config.UnmarshalKey(web.ServerConfigKey, defaultServerConfig)
		if err != nil {
			return nil, nil, err
		}
	}
	coreContext := core.NewContext(w.config, ctx)
	coreContext.AddService(w.services...)

	iModelGroups := make([]core.IModelGroup, 0)
	iModelGroups = append(iModelGroups, w.modelGroups...)
	if len(w.models) > 0 {
		modelGroupBuilder := core.NewModelGroupBuilder()
		if w.config.HasKey(db2.ConfigKey) {
			db, err := db2.CreateDB(w.config)
			if err != nil {
				log.Error("Failed to initialize the database", zap.Error(err))
				return nil, nil, err
			}
			modelGroupBuilder.DB(db)
		}
		modelGroupBuilder.Model(w.models...)
		modelGroup := modelGroupBuilder.Build()
		iModelGroups = append(iModelGroups, modelGroup)
	}

	if len(iModelGroups) > 0 {
		coreContext.AddModelGroup(iModelGroups...)
		for _, modelGroup := range iModelGroups {
			coreContext.AddModel(modelGroup.GetModel()...)
			err := modelGroup.Init(coreContext)
			if err != nil {
				return nil, nil, errors.WithStackIf(err)
			}
		}
	}
	runners := make([]core.IRunner, 0)
	for _, iService := range w.services {
		if v, ok := iService.(core.IRunner); ok {
			log.Debug("Init", zap.String("runner", util.GetStructFullQualifiedName(iService)))
			runners = append(runners, v)
		} else {
			log.Debug("Init", zap.String("service", util.GetStructFullQualifiedName(iService)))
		}

		err := iService.Init(coreContext)
		if err != nil {
			return nil, nil, errors.WithStackIf(err)
		}
	}
	restGroups := make([]*core.RestGroup, 0)
	restGroups = append(restGroups, w.restGroups...)
	coreContext.AddRunner(runners...)
	if w.config.HasKey(web.ServerConfigKey) || len(restGroups) == 0 || len(w.rests) > 0 || !w.handles.Empty() {
		restGroup := core.NewRestGroupBuilder().
			ServerConfig(defaultServerConfig).
			Rest(w.rests...).
			Filter(w.filters...).
			Handles(w.handles).
			Build()
		restGroups = append(restGroups, restGroup)
	}
	coreServer := core.NewServer(coreContext)
	coreServer.AddIRunner(runners...)
	coreServer.AddRestGroup(restGroups...)
	return coreServer, coreContext, nil

}

// ReStart stops the running application so that Run starts it again with the
// same parent context. It is safe to call from another goroutine (signal
// handler, admin route) and is a no-op when the application is not running yet.
func (w *WebFrame) ReStart() {
	w.mu.Lock()
	w.restart = true
	cancel := w.ctxCancelFunc
	w.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Run initializes the logger, sets up all components, services, models, and REST groups,
// then starts the HTTP servers and background runners. The provided context controls
// the application lifecycle for graceful shutdown. When ReStart is called, Run
// re-initializes and serves again instead of returning.
func (w *WebFrame) Run(pCtx context.Context) error {
	for {
		w.mu.Lock()
		w.restart = false
		w.mu.Unlock()

		err := w.run(pCtx)

		w.mu.Lock()
		restart := w.restart
		w.mu.Unlock()
		if !restart {
			return err
		}
	}
}

// run is a single lifecycle generation of Run: initialize, serve, and return
// when the server stops.
func (w *WebFrame) run(pCtx context.Context) error {
	ctx, cancel := context.WithCancel(pCtx)
	defer cancel()
	w.mu.Lock()
	w.ctxCancelFunc = cancel
	w.mu.Unlock()

	var logConfig = &log.Config{
		Level: "debug",
	}
	err := w.config.UnmarshalKey(log.ConfigKey, &logConfig)
	if err != nil {
		return err
	}
	defer func() {
		err := log.Sync()
		if err != nil {
			log.Error("Failed to close the service", zap.Error(err))
		}
	}()
	log.InitLogger(logConfig)
	server, _, err := w.init(ctx)
	if err != nil {
		return errors.WithStackIf(err)
	}
	err = server.Run()
	if isShutdown(err) {
		// A cancelled context (shutdown or ReStart) is not a failure.
		return nil
	}
	if err != nil {
		log.Error("Start the WebFrame", zap.Error(err))
	}
	return errors.WithStackIf(err)
}

// isShutdown reports whether err is the expected result of stopping the
// application (cancelled context or closed listener) rather than a failure.
func isShutdown(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, http.ErrServerClosed)
}

// Builder provides a fluent API for constructing a WebFrame application.
// Register routes, REST controllers, models, services, filters,
// and runners, then call Build() to create the application.
type Builder struct {
	restGroups  []*core.RestGroup
	modelGroups []core.IModelGroup
	config      config.IConfig
	models      []core.IModel
	services    []core.IService
	rests       []core.IRest
	filters     []core.IFilter
	handles     *web.Handles
}

// NewBuilder creates a new Builder with the given configuration for constructing a WebFrame.
//
// A single config is used as-is so that a file-backed config keeps its write-back support:
// MergeConfig returns a plain *Config with no file to write to, which makes WriteConfig
// fail for every caller that passes a SingleFileConfig.
func NewBuilder(configs ...config.IConfig) *Builder {
	var cfg config.IConfig
	if len(configs) == 1 {
		// Use the config as-is: MergeConfig returns a plain *Config with no file to write
		// to, so a SingleFileConfig passed through it can never WriteConfig.
		cfg = configs[0]
	} else {
		cfg = config.MergeConfig(configs...)
	}

	builder := &Builder{
		models:      make([]core.IModel, 0),
		services:    make([]core.IService, 0),
		restGroups:  make([]*core.RestGroup, 0),
		modelGroups: make([]core.IModelGroup, 0),
		rests:       make([]core.IRest, 0),
		filters:     make([]core.IFilter, 0),
		handles:     web.NewHandles(),
		config:      cfg,
	}
	return builder
}

// Get registers a GET route handler and returns the builder for chaining.
func (b *Builder) Get(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Get(relativePath, handlers...)
	return b
}

// Post registers a POST route handler and returns the builder for chaining.
func (b *Builder) Post(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Post(relativePath, handlers...)
	return b
}

// Delete registers a DELETE route handler and returns the builder for chaining.
func (b *Builder) Delete(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Delete(relativePath, handlers...)
	return b
}

// Put registers a PUT route handler and returns the builder for chaining.
func (b *Builder) Put(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Put(relativePath, handlers...)
	return b
}

// Any registers a route handler for all HTTP methods and returns the builder for chaining.
func (b *Builder) Any(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Any(relativePath, handlers...)
	return b
}

// Rest registers one or more REST controllers and returns the builder for chaining.
func (b *Builder) Rest(rest ...core.IRest) *Builder {
	b.rests = append(b.rests, rest...)
	return b
}

// Runner registers one or more background runners and returns the builder for chaining.
func (b *Builder) Runner(runners ...core.IRunner) *Builder {
	for _, runner := range runners {
		b.services = append(b.services, runner)
	}
	return b
}

// Model registers one or more models and returns the builder for chaining.
func (b *Builder) Model(model ...core.IModel) *Builder {
	b.models = append(b.models, model...)
	return b
}

// Service registers one or more services and returns the builder for chaining.
func (b *Builder) Service(service ...core.IService) *Builder {
	b.services = append(b.services, service...)
	return b
}

// Filter registers one or more filters and returns the builder for chaining.
func (b *Builder) Filter(filters ...core.IFilter) *Builder {
	b.filters = append(b.filters, filters...)
	return b
}

// RestGroup registers one or more REST groups and returns the builder for chaining.
func (b *Builder) RestGroup(restGroups ...*core.RestGroup) *Builder {
	b.restGroups = append(b.restGroups, restGroups...)
	return b
}

// ModelGroup registers one or more model groups and returns the builder for chaining.
func (b *Builder) ModelGroup(modelGroups ...core.IModelGroup) *Builder {
	b.modelGroups = append(b.modelGroups, modelGroups...)
	return b
}

// Build creates a WebFrame from the builder configuration.
// The returned application can be started with Run or Test.
func (b *Builder) Build() *WebFrame {
	w := &WebFrame{
		models:      b.models,
		services:    b.services,
		restGroups:  b.restGroups,
		modelGroups: b.modelGroups,
		rests:       b.rests,
		filters:     b.filters,
		handles:     b.handles,
		config:      b.config,
	}
	return w
}

// NewRestGroupBuilder creates a new REST group builder from the core package.
func NewRestGroupBuilder() *core.RestGroupBuilder {
	return core.NewRestGroupBuilder()
}

// NewModelGroupBuilder creates a new model group builder from the core package.
func NewModelGroupBuilder() *core.ModelGroupBuilder {
	return core.NewModelGroupBuilder()
}
