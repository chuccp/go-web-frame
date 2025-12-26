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
	"github.com/sourcegraph/conc/panics"
	"github.com/sourcegraph/conc/pool"
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
	certManager    *web.CertManager
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
		certManager: web.NewCertManager(),
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

	for _, group := range w.restGroups {
		if group.Port() == serverConfig.Port {
			return group
		}
	}
	groupGroup := core.NewRestGroup(serverConfig, w.certManager)
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
	w.context = core.NewContext(w.config, db, w.schedule, w.certManager)
	w.context.AddComponent(w.component...)
	w.context.AddModel(w.models...)
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
	rootGroup := core.NewRestGroup(serverConfig, w.certManager).AddRest(w.rests...).Authentication(w.authentication).AddMiddlewares(w.middlewareFunc...)
	hasRootGroup := false
	for _, group := range w.restGroups {
		if group.Port() == 0 || group.Port() == serverConfig.Port {
			group.Merge(rootGroup)
			hasRootGroup = true
			break
		}
	}
	if !hasRootGroup {
		w.restGroups = append(w.restGroups, rootGroup)
	}
	for _, group := range w.restGroups {
		context := w.context.Copy(group.DigestAuth(), group.HttpServer())
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
	core.RunDaemon(w, svcConfig)
}

func (w *WebFrame) Authentication(authentication web.Authentication) {
	w.authentication = authentication
}
