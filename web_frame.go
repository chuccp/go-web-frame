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

func GetService[T core.IService](c *core.Context) T {
	return core.GetService[T](c)
}

func GetModel[T core.IModel](c *core.Context) T {
	return core.GetModel[T](c)
}
func GetReNewModel[T core.IModel](db *db2.DB, c *core.Context) T {
	return core.GetReNewModel[T](db, c)
}
func GetComponent[T core.IComponent](c *core.Context) T {
	return core.GetComponent[T](c)
}

func GetRunner[T core.IRunner](c *core.Context) T {
	return core.GetRunner[T](c)
}
func GetFilter[T core.IFilter](c *core.Context) T {
	return core.GetFilter[T](c)
}
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

func (w *WebFrame) Start() error {
	return w.Run(context.Background())
}
func (w *WebFrame) Test(f func(ctx *core.Context) error) error {
	_, ctx, err := w.init(context.Background())
	if err != nil {
		return err
	}
	return f(ctx)
}
func (w *WebFrame) init(ctx context.Context) (*core.Server, *core.Context, error) {

	gin.SetMode(gin.ReleaseMode)

	for _, component := range w.component {
		log.Debug("Init", zap.String("component", util.GetStructFullQualifiedName(component)))
		err := errors.WithStackIf(component.Init(ctx, w.config))
		if err != nil {
			log.Error("Failed to initialize the component", zap.Error(err))
			return nil, nil, err
		}
	}

	coreContext := core.NewContext(w.config)
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
func (w *WebFrame) Run(ctx context.Context) error {
	var logConfig log.Config
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
	log.InitLogger(&logConfig)

	server, _, err := w.init(ctx)
	if err != nil {
		return errors.WithStackIf(err)
	}
	return errors.WithStackIf(server.Run(ctx))
}

type Builder struct {
	component         []core.IComponent
	restGroups        []*core.RestGroup
	modelGroup        []core.IModelGroup
	config            config2.IConfig
	models            []core.IModel
	services          []core.IService
	rests             []core.IRest
	runners           []core.IRunner
	filters           []core.IFilter
	defaultModelGroup core.IModelGroup
	handles           *web.Handles
}

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
		//defaultModelGroup: core.DefaultModelGroup(),
		handles: web.NewHandles(),
		config:  config,
	}
	return builder
}

func (b *Builder) Get(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Handle(http.MethodGet, relativePath, handlers...)
	return b
}

func (b *Builder) Post(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Handle(http.MethodPost, relativePath, handlers...)
	return b
}

func (b *Builder) Delete(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Handle(http.MethodDelete, relativePath, handlers...)
	return b
}

func (b *Builder) Put(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Handle(http.MethodPut, relativePath, handlers...)
	return b
}

func (b *Builder) Any(relativePath string, handlers ...web.HandlerFunc) *Builder {
	b.handles.Handle(http.MethodGet, relativePath, handlers...)
	return b
}

func (b *Builder) Rest(rest ...core.IRest) *Builder {
	b.rests = append(b.rests, rest...)
	return b
}

func (b *Builder) Component(component ...core.IComponent) *Builder {
	b.component = append(b.component, component...)
	return b
}

func (b *Builder) Runner(runner ...core.IRunner) *Builder {
	b.runners = append(b.runners, runner...)
	return b
}

func (b *Builder) Model(model ...core.IModel) *Builder {
	b.models = append(b.models, model...)
	return b
}

func (b *Builder) Service(service ...core.IService) *Builder {
	b.services = append(b.services, service...)
	return b
}

func (b *Builder) Filter(filters ...core.IFilter) *Builder {
	b.filters = append(b.filters, filters...)
	return b
}

func (b *Builder) RestGroup(restGroups ...*core.RestGroup) *Builder {
	b.restGroups = append(b.restGroups, restGroups...)
	return b
}

func (b *Builder) ModelGroup(modelGroups ...core.IModelGroup) *Builder {
	b.modelGroup = append(b.modelGroup, modelGroups...)
	return b
}
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

func NewRestGroupBuilder() *core.RestGroupBuilder {
	return core.NewRestGroupBuilder()
}
func NewModelGroupBuilder() *core.ModelGroupBuilder {
	return core.NewModelGroupBuilder()
}
