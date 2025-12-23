package core

import (
	"errors"
	"sync"

	config2 "github.com/chuccp/go-web-frame/config"
	db2 "github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/sourcegraph/conc/panics"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WebFrame struct {
	component      []IComponent
	restGroups     []*RestGroup
	config         config2.IConfig
	httpServers    []*web.HttpServer
	context        *Context
	models         []IModel
	services       []IService
	configs        []IValueConfig
	rests          []IRest
	middlewareFunc []MiddlewareFunc
	authentication web.Authentication
	db             *gorm.DB
	certManager    *web.CertManager
	schedule       *Schedule
	serverConfig   *web.ServerConfig
}

func New(config config2.IConfig) *WebFrame {
	w := &WebFrame{
		httpServers:  make([]*web.HttpServer, 0),
		models:       make([]IModel, 0),
		services:     make([]IService, 0),
		restGroups:   make([]*RestGroup, 0),
		rests:        make([]IRest, 0),
		component:    make([]IComponent, 0),
		certManager:  web.NewCertManager(),
		schedule:     NewSchedule(),
		config:       config,
		serverConfig: web.DefaultServerConfig(),
	}
	return w
}
func (w *WebFrame) GetServerConfig() *web.ServerConfig {
	return w.serverConfig
}
func (w *WebFrame) AddRest(rest ...IRest) {
	w.rests = append(w.rests, rest...)
}
func (w *WebFrame) AddComponent(component ...IComponent) {
	w.component = append(w.component, component...)
}

func (w *WebFrame) AddModel(model ...IModel) {
	w.models = append(w.models, model...)
	for _, iModel := range model {
		w.addService(iModel)
	}
}
func (w *WebFrame) RegisterConfig(configs ...IValueConfig) {
	for _, config := range configs {
		w.configs = append(w.configs, config)
	}
}
func (w *WebFrame) addService(service IService) {
	w.services = append(w.services, service)
}
func (w *WebFrame) AddService(service ...IService) {
	w.services = append(w.services, service...)
}
func (w *WebFrame) GetRestGroup(serverConfig *web.ServerConfig) *RestGroup {

	for _, group := range w.restGroups {
		if group.port == serverConfig.Port {
			return group
		}
	}
	groupGroup := newRestGroup(serverConfig)
	w.restGroups = append(w.restGroups, groupGroup)
	return groupGroup
}
func (w *WebFrame) AddMiddleware(middlewareFunc ...MiddlewareFunc) {
	w.middlewareFunc = append(w.middlewareFunc, middlewareFunc...)
}
func (w *WebFrame) getHttpServer(serverConfig *web.ServerConfig) *web.HttpServer {
	for _, httpServer := range w.httpServers {
		if httpServer.Port() == serverConfig.Port {
			return httpServer
		}
	}
	httpServer := web.NewHttpServer(serverConfig, w.certManager)
	w.httpServers = append(w.httpServers, httpServer)
	return httpServer
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
	return errors.Join(errs...)
}
func (w *WebFrame) Start() error {
	gin.SetMode(gin.ReleaseMode)
	var logConfig log.Config
	err := w.config.Unmarshal(logConfig.Key(), &logConfig)
	if err != nil {
		return err
	}
	log.InitLogger(&logConfig)
	for _, config := range w.configs {
		err := w.config.Unmarshal(config.Key(), config)
		if err != nil {
			log.Error("加载配置文件失败:", zap.Any(config.Key(), config), zap.Error(err))
		}
	}

	db, err := db2.InitDB(w.config)
	if err != nil && !errors.Is(err, db2.NoConfigDBError) {
		log.Error("初始化数据库失败:", zap.Error(err))
		return err
	}
	for _, component := range w.component {
		err := component.Init(w.config)
		if err != nil {
			log.Error("初始化组件失败:", zap.NamedError(component.Name(), err))
			return err
		}
	}
	err = w.schedule.Init(w.config)
	if err != nil {
		log.Error("初始化计划任务失败:", zap.Error(err))
		return err
	}
	w.db = db
	w.context = &Context{
		rLock:        new(sync.RWMutex),
		config:       w.config,
		restMap:      make(map[string]IRest),
		modelMap:     make(map[string]IModel),
		serviceMap:   make(map[string]IService),
		componentMap: make(map[string]IComponent),
		db:           db,
		transaction:  model.NewTransaction(db),
		configMap:    make(map[string]IValueConfig),
	}
	contextGroup := newContextGroup(w.context)
	w.context.contextGroup = contextGroup
	w.context.addComponent(w.component...)
	w.context.addModel(w.models...)
	w.context.addConfig(w.configs...)
	w.context.AddService(w.services...)
	for _, iService := range w.services {
		iService.Init(w.context)
	}
	err = w.config.Unmarshal(w.serverConfig.Key(), w.serverConfig)
	if err != nil {
		return err
	}
	rootGroup := newRestGroup(w.serverConfig).AddRest(w.rests...).Authentication(w.authentication).AddMiddlewares(w.middlewareFunc...)
	hasRootGroup := false
	for _, group := range w.restGroups {
		if group.port == 0 || group.port == w.serverConfig.Port {
			group.merge(rootGroup)
			hasRootGroup = true
			break
		}
	}
	if !hasRootGroup {
		w.restGroups = append(w.restGroups, rootGroup)
	}
	for _, group := range w.restGroups {
		group.httpServer = w.getHttpServer(group.serverConfig)
		for _, rest := range group.rests {
			w.context.AddRest(rest)
		}
	}
	for _, group := range w.restGroups {
		context := w.context.Copy(group.digestAuth, group.httpServer)
		group.UseMiddleware(context)
		for _, rest := range group.rests {
			rest.Init(context)
		}

	}
	var wg = pool.New()
	wg.WithMaxGoroutines(len(w.httpServers))
	errorsPool := wg.WithErrors()
	if len(w.httpServers) > 0 {
		for _, engine := range w.httpServers {
			errorsPool.Go(func() error {
				var catcher panics.Catcher
				catcher.Try(func() {
					err := engine.Run()
					if err != nil {
						log.PanicErrors("启动服务失败:", err)
					}
				})
				return catcher.Recovered().AsError()
			})
		}
	}
	w.certManager.Start()
	return errorsPool.Wait()
}

func (w *WebFrame) Daemon(svcConfig *service.Config) {
	RunDaemon(w, svcConfig)
}

func (w *WebFrame) Authentication(authentication web.Authentication) {
	w.authentication = authentication
}
