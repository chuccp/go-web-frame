package wf

import (
	"context"
	"net/http"

	"emperror.dev/errors"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	db2 "github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetService retrieves a service of the specified type from the context.
func GetService[T core.IService](c *core.Context) T {
	return core.GetService[T](c)
}

// GetModel retrieves a model of the specified type from the context.
func GetModel[T core.IModel](c *core.Context) T {
	return core.GetModel[T](c)
}

// GetReNewModel retrieves a model and creates a fresh instance with the given database connection.
func GetReNewModel[T core.IModel](db *db2.DB, c *core.Context) T {
	return core.GetReNewModel[T](db, c)
}

// GetComponent retrieves a component of the specified type from the context.
func GetComponent[T core.IComponent](c *core.Context) T {
	return core.GetComponent[T](c)
}

// GetRunner retrieves a runner of the specified type from the context.
func GetRunner[T core.IRunner](c *core.Context) T {
	return core.GetRunner[T](c)
}

// GetFilter retrieves a filter of the specified type from the context.
func GetFilter[T core.IFilter](c *core.Context) T {
	return core.GetFilter[T](c)
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
	component  []core.IComponent
	restGroups []*core.RestGroup
	modelGroup []core.IModelGroup
	config     config2.IConfig
	models     []core.IModel
	services   []core.IService
	rests      []core.IRest
	runners    []core.IRunner
	filters    []core.IFilter
	//defaultModelGroup core.IModelGroup
	handles *web.Handles
}

// Start initializes and runs the web application with a background context.
// Blocks until the application is shut down.
func (w *WebFrame) Start() error {
	return w.Run(context.Background())
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

	coreContext := core.NewContext(w.handles, w.config, ctx)
	coreContext.AddComponent(w.component...)
	coreContext.AddService(w.services...)
	coreContext.AddRunner(w.runners...)

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
		w.modelGroup = append(w.modelGroup, modelGroup)
	}

	if len(w.modelGroup) > 0 {
		coreContext.AddModelGroup(w.modelGroup...)
		for _, modelGroup := range w.modelGroup {
			coreContext.AddModel(modelGroup.GetModel()...)
			err := modelGroup.Init(coreContext)
			if err != nil {
				return nil, nil, errors.WithStackIf(err)
			}
		}
	}

	for _, component := range w.component {
		log.Debug("Init", zap.String("component", util.GetStructFullQualifiedName(component)))
		err := errors.WithStackIf(component.Init(ctx, w.config))
		if err != nil {
			log.Error("Failed to initialize the component", zap.Error(err))
			return nil, nil, err
		}
	}

	for _, iService := range w.services {
		log.Debug("Init", zap.String("service", util.GetStructFullQualifiedName(iService)))
		err := iService.Init(coreContext)
		if err != nil {
			return nil, nil, errors.WithStackIf(err)
		}
	}

	if w.config.HasKey(web.ServerConfigKey) || len(w.restGroups) == 0 || len(w.rests) > 0 || !w.handles.Empty() {
		var serverConfig = web.DefaultServerConfig()
		err := w.config.UnmarshalKey(web.ServerConfigKey, &serverConfig)
		if err != nil {
			return nil, nil, errors.WithStackIf(err)
		}
		restGroup := core.NewRestGroupBuilder().
			ServerConfig(serverConfig).
			Handles(w.handles).
			Rest(w.rests...).
			Filter(w.filters...).
			Build()
		w.restGroups = append(w.restGroups, restGroup)
	}
	server := core.NewServer(w.restGroups, w.runners)
	err := server.Init(coreContext)
	if err != nil {
		return nil, nil, errors.WithStackIf(err)
	}
	return server, coreContext, nil

}
// Run initializes the logger, sets up all components, services, models, and REST groups,
// then starts the HTTP servers and background runners. The provided context controls
// the application lifecycle for graceful shutdown.
func (w *WebFrame) Run(ctx context.Context) error {
	var logConfig = &log.Config{
		Level: "debug",
	}
	err := w.config.UnmarshalKey(logConfig.Key(), &logConfig)
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
	return errors.WithStackIf(server.Run())
}

// Builder provides a fluent API for constructing a WebFrame application.
// Register routes, REST controllers, models, services, filters, components,
// and runners, then call Build() to create the application.
type Builder struct {
	component  []core.IComponent
	restGroups []*core.RestGroup
	modelGroup []core.IModelGroup
	config     config2.IConfig
	models     []core.IModel
	services   []core.IService
	rests      []core.IRest
	runners    []core.IRunner
	filters    []core.IFilter
	handles    *web.Handles
}

// NewBuilder creates a new Builder with the given configuration for constructing a WebFrame.
func NewBuilder(config config2.IConfig) *Builder {

	builder := &Builder{
		models:     make([]core.IModel, 0),
		services:   make([]core.IService, 0),
		restGroups: make([]*core.RestGroup, 0),
		modelGroup: make([]core.IModelGroup, 0),
		rests:      make([]core.IRest, 0),
		component:  make([]core.IComponent, 0),
		runners:    make([]core.IRunner, 0),
		filters:    make([]core.IFilter, 0),
		handles:    web.NewHandles(),
		config:     config,
	}
	return builder
}

// Get registers a GET route handler and returns the builder for chaining.
func (b *Builder) Get(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Handle(http.MethodGet, relativePath, handlers...)
	return b
}

// Post registers a POST route handler and returns the builder for chaining.
func (b *Builder) Post(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Handle(http.MethodPost, relativePath, handlers...)
	return b
}

// Delete registers a DELETE route handler and returns the builder for chaining.
func (b *Builder) Delete(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Handle(http.MethodDelete, relativePath, handlers...)
	return b
}

// Put registers a PUT route handler and returns the builder for chaining.
func (b *Builder) Put(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Handle(http.MethodPut, relativePath, handlers...)
	return b
}

// Any registers a route handler for all HTTP methods and returns the builder for chaining.
func (b *Builder) Any(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Handle(http.MethodGet, relativePath, handlers...)
	return b
}

// Rest registers one or more REST controllers and returns the builder for chaining.
func (b *Builder) Rest(rest ...core.IRest) *Builder {
	b.rests = append(b.rests, rest...)
	return b
}

// Component registers one or more independent components and returns the builder for chaining.
func (b *Builder) Component(component ...core.IComponent) *Builder {
	b.component = append(b.component, component...)
	return b
}

// Runner registers one or more background runners and returns the builder for chaining.
func (b *Builder) Runner(runner ...core.IRunner) *Builder {
	b.runners = append(b.runners, runner...)
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
	b.modelGroup = append(b.modelGroup, modelGroups...)
	return b
}

// Build creates a WebFrame from the builder configuration.
// The returned application can be started with Run or Test.
func (b *Builder) Build() *WebFrame {
	w := &WebFrame{
		models:     b.models,
		services:   b.services,
		restGroups: b.restGroups,
		modelGroup: b.modelGroup,
		rests:      b.rests,
		component:  b.component,
		runners:    b.runners,
		filters:    b.filters,
		config:     b.config,
		handles:    b.handles,
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
