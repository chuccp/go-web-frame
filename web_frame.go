package wf

import (
	"sync"

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

func UnmarshalConfig[T any](key string, c *core.Context) T {
	return core.UnmarshalConfig[T](key, c)
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
	middlewareFunc    []core.MiddlewareFunc
	authentication    web.Authentication
	db                *gorm.DB
	schedule          *core.Schedule
	server            *core.Server
	lock              *sync.Mutex
	defaultModelGroup core.IModelGroup
	isClose           bool
}

func New(config config2.IConfig) *WebFrame {
	//ctx2, cancel := context.WithCancel(context.Background())
	w := &WebFrame{
		models:            make([]core.IModel, 0),
		services:          make([]core.IService, 0),
		restGroups:        make([]*core.RestGroup, 0),
		modelGroup:        make([]core.IModelGroup, 0),
		rests:             make([]core.IRest, 0),
		component:         make([]core.IComponent, 0),
		runners:           make([]core.IRunner, 0),
		config:            config,
		schedule:          core.NewSchedule(),
		lock:              new(sync.Mutex),
		defaultModelGroup: core.DefaultModelGroup(),
		isClose:           false,
	}
	return w
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
func (w *WebFrame) GetRestGroup(serverConfig *web.ServerConfig) *core.RestGroup {
	groupGroup := core.NewRestGroup(serverConfig)
	w.restGroups = append(w.restGroups, groupGroup)
	return groupGroup
}
func (w *WebFrame) AddMiddleware(middlewareFunc ...core.MiddlewareFunc) {
	w.middlewareFunc = append(w.middlewareFunc, middlewareFunc...)
}

func (w *WebFrame) Close() error {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.isClose = true
	errs := make([]error, 0)
	err := w.server.Destroy()
	errs = append(errs, err)
	err = log.Sync()
	errs = append(errs, err)
	err = w.schedule.Destroy()
	errs = append(errs, err)
	for _, component := range w.component {
		err = component.Destroy()
		errs = append(errs, err)
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Combine(errs...)
}
func (w *WebFrame) Start() error {
	err := w.init()
	if err != nil {
		return err
	}
	if w.isClose {
		return errors.New("The service has been closed")
	}
	return w.server.Run()
}
func (w *WebFrame) init() error {
	w.lock.Lock()
	defer w.lock.Unlock()
	gin.SetMode(gin.ReleaseMode)
	var logConfig log.Config
	err := w.config.Unmarshal(logConfig.Key(), &logConfig)
	if err != nil {
		return err
	}
	log.InitLogger(&logConfig)

	for _, component := range w.component {
		err := errors.WithStackIf(component.Init(w.config))
		if err != nil {
			log.Error("Failed to initialize the component", zap.Error(err))
			return err
		}
	}

	coreContext := core.NewContext(w.config, w.schedule, w.defaultModelGroup)
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
	w.server = core.NewServer(w.restGroups, w.runners)
	err = w.server.Init(coreContext)
	if err != nil {
		return err
	}
	err = w.schedule.Init(w.config)
	if err != nil {
		log.Error("Failed to initialize the scheduled task", zap.Error(err))
		return err
	}
	return nil
}

func (w *WebFrame) Daemon(svcConfig *service.Config) {
	RunDaemon(w, svcConfig)
}

func (w *WebFrame) Authentication(authentication web.Authentication) {
	w.authentication = authentication
}
