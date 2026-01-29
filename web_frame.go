package wf

import (
	"context"
	"net/http"
	"os"
	"path"
	"strings"

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
func UnmarshalKeyConfig[T any](key string, c *core.Context) T {
	return core.UnmarshalKeyConfig[T](key, c)
}

type DefaultConverter struct {
	ctx *core.Context
}

func (receiver *DefaultConverter) Init(ctx *core.Context) error {
	receiver.ctx = ctx
	return nil
}

func (receiver *DefaultConverter) Request(filterChain web.FilterChain, request *web.Request) {
	value, err := filterChain.Next()
	resp := request.Response()
	if err != nil {
		err0 := web.Errors(value, err)
		resp.JSON(err0.Code, err0)
		resp.Abort()
	} else {
		if value != nil {
			switch t := value.(type) {
			case *web.Message:
				if t.Code == http.StatusMovedPermanently {
					resp.Redirect(http.StatusMovedPermanently, t.Data.(string))
					resp.Abort()
					return
				}
				resp.JSON(t.Code, value)
			case string:
				_, err2 := resp.Write([]byte(t))
				if err2 != nil {
					resp.Abort()
					return
				}
			case *web.File:
				if len(t.FileName) == 0 {
					_, filename := path.Split(t.Path)
					t.FileName = filename
				}
				if util.IsNotBlank(t.Suffix) && !strings.HasSuffix(t.FileName, t.Suffix) {
					if !strings.HasPrefix(t.Suffix, ".") {
						t.Suffix = "." + t.Suffix
					}
					t.FileName = t.FileName + t.Suffix
				}
				resp.FileAttachment(t.Path, t.FileName)
			case *os.File:
				resp.FileAttachment(t.Name(), t.Name())
			default:
				resp.JSON(200, web.Data(value))
			}
		}
	}
}

type DefaultRest struct {
	ctx *core.Context
}

func (receiver *DefaultRest) Init(ctx *core.Context) error {
	receiver.ctx = ctx
	return nil
}

type WebFrame struct {
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

func NewWithAutoConfig() *WebFrame {
	return New(LoadAutoConfig())
}

func New(config config2.IConfig) *WebFrame {
	w := &WebFrame{
		models:            make([]core.IModel, 0),
		services:          make([]core.IService, 0),
		restGroups:        make([]*core.RestGroup, 0),
		modelGroup:        make([]core.IModelGroup, 0),
		rests:             make([]core.IRest, 0),
		component:         make([]core.IComponent, 0),
		runners:           make([]core.IRunner, 0),
		filters:           make([]core.IFilter, 0),
		config:            config,
		defaultModelGroup: core.DefaultModelGroup(),
		handles:           web.NewHandles(),
	}
	return w
}

func (w *WebFrame) Get(relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return w.handles.Handle(http.MethodGet, relativePath, handlers...)
}

func (w *WebFrame) Post(relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return w.handles.Handle(http.MethodPost, relativePath, handlers...)
}
func (w *WebFrame) Delete(relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return w.handles.Handle(http.MethodDelete, relativePath, handlers...)

}
func (w *WebFrame) Put(relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return w.handles.Handle(http.MethodPut, relativePath, handlers...)
}

func (w *WebFrame) Any(relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return w.handles.Handle(http.MethodGet, relativePath, handlers...)

}

func (w *WebFrame) AddRest(rest ...core.IRest) {
	w.rests = append(w.rests, rest...)
}
func (w *WebFrame) AddComponent(component ...core.IComponent) {
	w.component = append(w.component, component...)
}
func (w *WebFrame) AddRunner(runner ...core.IRunner) {
	w.runners = append(w.runners, runner...)
}
func (w *WebFrame) AddModel(model ...core.IModel) {
	w.models = append(w.models, model...)
}
func (w *WebFrame) addService(service core.IService) {
	w.services = append(w.services, service)
}
func (w *WebFrame) AddService(service ...core.IService) {
	w.services = append(w.services, service...)
}
func (w *WebFrame) NewRestGroup(serverConfig *web.ServerConfig) *core.RestGroup {
	groupGroup := core.NewRestGroup(serverConfig, &DefaultConverter{}, web.NewHandles())
	w.restGroups = append(w.restGroups, groupGroup)
	return groupGroup
}
func (w *WebFrame) AddFilter(filters ...core.IFilter) {
	w.filters = append(w.filters, filters...)
}
func (w *WebFrame) Start() error {
	return w.Run(context.Background())
}

func (w *WebFrame) Run(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
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

	for _, component := range w.component {
		err := errors.WithStackIf(component.Init(ctx, w.config))
		if err != nil {
			log.Error("Failed to initialize the component", zap.Error(err))
			return err
		}
	}

	coreContext := core.NewContext(w.config, w.defaultModelGroup)
	coreContext.AddComponent(w.component...)
	coreContext.AddService(w.services...)
	coreContext.AddRunner(w.runners...)

	if len(w.models) > 0 {
		coreContext.AddModel(w.models...)
		w.defaultModelGroup.AddModel(w.models...)
	}
	if w.config.HasKey(db2.ConfigKey) {
		db, err := db2.CreateDB(w.config)
		if err != nil {
			log.Error("Failed to initialize the database", zap.Error(err))
			return err
		}
		err = w.defaultModelGroup.SwitchDB(db, coreContext)
		if err != nil {
			log.Error("Failed to switch the database", zap.Error(err))
			return err
		}
	}

	if len(w.modelGroup) > 0 {
		coreContext.AddModelGroup(w.modelGroup...)
		for _, modelGroup := range w.modelGroup {
			coreContext.AddModel(modelGroup.GetModel()...)
			err := modelGroup.Init(coreContext)
			if err != nil {
				return errors.WithStackIf(err)
			}
		}
	}

	for _, iService := range w.services {
		err := iService.Init(coreContext)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}

	if w.config.HasKey(web.ServerConfigKey) || len(w.restGroups) == 0 || len(w.rests) > 0 {
		var serverConfig = web.DefaultServerConfig()
		err = w.config.UnmarshalKey(web.ServerConfigKey, &serverConfig)
		if err != nil {
			return err
		}
		rootGroup := core.NewRestGroup(serverConfig, &DefaultConverter{}, w.handles)
		rootGroup.AddRest(w.rests...)
		rootGroup.AddFilter(w.filters...)
		w.restGroups = append(w.restGroups, rootGroup)
	}
	server := core.NewServer(w.restGroups, w.runners)
	err = server.Init(coreContext)
	if err != nil {
		return errors.WithStackIf(err)
	}
	return errors.WithStackIf(server.Run(ctx))
}
