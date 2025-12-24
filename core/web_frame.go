package core

import (
	"sync"

	"emperror.dev/errors"
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
	rests          []IRest
	middlewareFunc []MiddlewareFunc
	authentication web.Authentication
	db             *gorm.DB
	certManager    *web.CertManager
	schedule       *Schedule
}

func New(config config2.IConfig) *WebFrame {
	w := &WebFrame{
		httpServers: make([]*web.HttpServer, 0),
		models:      make([]IModel, 0),
		services:    make([]IService, 0),
		restGroups:  make([]*RestGroup, 0),
		rests:       make([]IRest, 0),
		component:   make([]IComponent, 0),
		certManager: web.NewCertManager(),
		schedule:    NewSchedule(),
		config:      config,
	}
	return w
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
		schedule:     w.schedule,
		certManager:  w.certManager,
	}
	contextGroup := newContextGroup(w.context)
	w.context.contextGroup = contextGroup
	w.context.addComponent(w.component...)
	w.context.addModel(w.models...)
	w.context.AddService(w.services...)
	for _, iService := range w.services {
		err := iService.Init(w.context)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}
	var serverConfig = web.DefaultServerConfig()
	err = w.config.Unmarshal(serverConfig.Key(), &serverConfig)
	if err != nil {
		return err
	}
	rootGroup := newRestGroup(serverConfig).AddRest(w.rests...).Authentication(w.authentication).AddMiddlewares(w.middlewareFunc...)
	hasRootGroup := false
	for _, group := range w.restGroups {
		if group.port == 0 || group.port == serverConfig.Port {
			group.merge(rootGroup)
			hasRootGroup = true
			break
		}
	}
	if !hasRootGroup {
		w.restGroups = append(w.restGroups, rootGroup)
	}
	for _, group := range w.restGroups {
		for _, rest := range group.rests {
			w.context.AddRest(rest)
		}
	}
	for _, group := range w.restGroups {
		context := w.context.Copy(group.digestAuth, group.httpServer)
		group.Init(context)
	}
	var wg = pool.New()
	wg.WithMaxGoroutines(len(w.restGroups))
	errorsPool := wg.WithErrors()
	if len(w.restGroups) > 0 {
		for _, rg := range w.restGroups {
			errorsPool.Go(func() error {
				var catcher panics.Catcher
				catcher.Try(func() {
					err := rg.Run()
					if err != nil {
						log.PanicErrors("Failed to start the HTTP service", err)
					}
				})
				return catcher.Recovered().AsError()
			})
		}
	}
	w.certManager.Start()
	return errors.WithStackIf(errorsPool.Wait())
}

func (w *WebFrame) Daemon(svcConfig *service.Config) {
	RunDaemon(w, svcConfig)
}

func (w *WebFrame) Authentication(authentication web.Authentication) {
	w.authentication = authentication
}
