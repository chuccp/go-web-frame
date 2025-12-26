package wf

import (
	"emperror.dev/errors"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	db2 "github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WebFrame struct {
	component      []core.IComponent
	restGroups     []*core.RestGroup
	config         config2.IConfig
	httpServers    []*web.HttpServer
	context        *core.Context
	models         []core.IModel
	services       []core.IService
	rests          []core.IRest
	middlewareFunc []core.MiddlewareFunc
	authentication web.Authentication
	db             *gorm.DB
	schedule       *core.Schedule
}

func New(config config2.IConfig) *WebFrame {
	w := &WebFrame{
		httpServers: make([]*web.HttpServer, 0),
		models:      make([]core.IModel, 0),
		services:    make([]core.IService, 0),
		restGroups:  make([]*core.RestGroup, 0),
		rests:       make([]core.IRest, 0),
		component:   make([]core.IComponent, 0),
		schedule:    core.NewSchedule(),
		config:      config,
	}
	return w
}
func (w *WebFrame) AddRest(rest ...core.IRest) {
	w.rests = append(w.rests, rest...)
}
func (w *WebFrame) AddComponent(component ...core.IComponent) {
	w.component = append(w.component, component...)
}

func (w *WebFrame) AddModel(model ...core.IModel) {
	w.models = append(w.models, model...)
	for _, iModel := range model {
		w.addService(iModel)
	}
}
func (w *WebFrame) addService(service core.IService) {
	w.services = append(w.services, service)
}
func (w *WebFrame) AddService(service ...core.IService) {
	w.services = append(w.services, service...)
}
func (w *WebFrame) GetRestGroup(serverConfig *web.ServerConfig) *core.RestGroup {
	groupGroup := core.NewRestGroup(serverConfig)
	w.restGroups = append(w.restGroups, groupGroup)
	return groupGroup
}
func (w *WebFrame) AddMiddleware(middlewareFunc ...core.MiddlewareFunc) {
	w.middlewareFunc = append(w.middlewareFunc, middlewareFunc...)
}

func (w *WebFrame) Close() error {
	errs := make([]error, 0)
	for _, server := range w.httpServers {
		err := server.Close()
		if err != nil {
			log.Error("关闭服务失败:", zap.Error(err))
			errs = append(errs, err)
		}
	}
	err := log.Sync()
	errs = append(errs, err)
	if len(errs) == 0 {
		return nil
	}
	return errors.Combine(errs...)
}
func (w *WebFrame) Start() error {
	gin.SetMode(gin.ReleaseMode)
	var logConfig log.Config
	err := w.config.Unmarshal(logConfig.Key(), &logConfig)
	if err != nil {
		return err
	}
	log.InitLogger(&logConfig)
	db, err := db2.InitDB(w.config)
	if err != nil && !errors.Is(err, db2.NoConfigDBError) {
		log.Error("Failed to initialize the database", zap.Error(err))
		return err
	}
	for _, component := range w.component {
		err := errors.WithStackIf(component.Init(w.config))
		if err != nil {
			log.Error("Failed to initialize the component", zap.Error(err))
			return err
		}
	}
	err = w.schedule.Init(w.config)
	if err != nil {
		log.Error("Failed to initialize the scheduled task", zap.Error(err))
		return err
	}
	w.context = core.NewContext(w.config, db, w.schedule)
	w.context.AddComponent(w.component...)
	w.context.AddModel(w.models...)
	w.context.AddService(w.services...)
	for _, iService := range w.services {
		err := iService.Init(w.context)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}
	if w.config.HasKey(web.ServerConfigKey) || len(w.restGroups) == 0 || len(w.rests) > 0 {
		var serverConfig = web.DefaultServerConfig()
		err = w.config.Unmarshal(web.ServerConfigKey, &serverConfig)
		if err != nil {
			return err
		}
		rootGroup := core.NewRestGroup(serverConfig)
		rootGroup.AddRest(w.rests...)
		rootGroup.Authentication(w.authentication)
		rootGroup.AddMiddlewares(w.middlewareFunc...)
		w.restGroups = append(w.restGroups, rootGroup)
	}
	server := core.NewServer(w.restGroups)
	err = server.Init(w.context)
	if err != nil {
		return err
	}
	return server.Run()
}

func (w *WebFrame) Daemon(svcConfig *service.Config) {
	core.RunDaemon(w, svcConfig)
}

func (w *WebFrame) Authentication(authentication web.Authentication) {
	w.authentication = authentication
}
