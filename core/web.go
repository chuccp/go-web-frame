package core

import (
	"errors"
	"log"
	"sync"

	config2 "github.com/chuccp/go-web-frame/config"
	db2 "github.com/chuccp/go-web-frame/db"
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type webEngine struct {
	engine *gin.Engine
	port   int
	log    *log2.Logger
}

func (e *webEngine) run() error {
	e.log.Info("启动服务", zap.String("serving run", "http://127.0.0.1:"+cast.ToString(e.port)))
	return e.engine.Run(":" + cast.ToString(e.port))
}

func defaultEngine(port int, log *log2.Logger) *webEngine {
	engine := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = false
	config.AllowCredentials = true
	config.AllowOriginFunc = func(origin string) bool {
		return true
	}
	engine.Use(cors.New(config))
	return &webEngine{
		engine: engine,
		port:   port,
		log:    log,
	}
}

type Web struct {
	component      []IComponent
	restGroups     []*RestGroup
	log            *log2.Logger
	config         *config2.Config
	engines        []*webEngine
	context        *Context
	models         []IModel
	services       []IService
	rests          []IRest
	authentication web.Authentication
	db             *gorm.DB
}

func CreateWeb(configFiles ...string) *Web {
	w := &Web{
		engines:    make([]*webEngine, 0),
		models:     make([]IModel, 0),
		services:   make([]IService, 0),
		restGroups: make([]*RestGroup, 0),
		rests:      make([]IRest, 0),
		component:  make([]IComponent, 0),
	}
	loadConfig, err := config2.LoadConfig(configFiles...)
	if err != nil {
		log.Panic("加载配置文件失败:", err)
		return nil
	}
	w.Configure(loadConfig)
	return w
}
func (w *Web) Configure(config *config2.Config) {
	w.config = config
}

func (w *Web) AddRest(rest ...IRest) {
	w.rests = append(w.rests, rest...)
}
func (w *Web) AddComponent(component ...IComponent) {
	w.component = append(w.component, component...)
}

func (w *Web) AddModel(model ...IModel) {
	w.models = append(w.models, model...)
	for _, iModel := range model {
		w.addService(iModel)
	}
}
func (w *Web) addService(service IService) {
	w.services = append(w.services, service)
}
func (w *Web) AddService(service ...IService) {
	w.services = append(w.services, service...)
}
func (w *Web) GetRestGroup(port ...int) *RestGroup {
	if len(port) > 1 {
		log.Panic("参数错误:", "port的数量不能大于1")
	}
	_port_ := 0
	if len(port) == 1 {
		_port_ = port[0]
	}
	for _, group := range w.restGroups {
		if group.port == _port_ {
			return group
		}
	}
	groupGroup := newRestGroup(_port_)
	w.restGroups = append(w.restGroups, groupGroup)
	return groupGroup
}
func (w *Web) getEngine(port int, log *log2.Logger) *webEngine {
	for _, engine := range w.engines {
		if engine.port == port {
			return engine
		}
	}
	engine := defaultEngine(port, log)
	w.engines = append(w.engines, engine)
	return engine
}

func (w *Web) Start() error {
	debug := w.config.GetBoolOrDefault("web.server.debug", true)
	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	logZap := log2.InitLogger(w.config)

	db, err := db2.InitDB(w.config, logZap)
	if err != nil && !errors.Is(err, db2.NoConfigDBError) {
		log.Panic("初始化数据库失败:", err)
		return err
	}
	for _, component := range w.component {
		err := component.Init(w.config)
		if err != nil {
			log.Panic("初始化组件失败:", component.Name(), err)
			return err
		}
	}
	w.db = db
	w.log = logZap
	w.context = &Context{
		rLock:        new(sync.RWMutex),
		config:       w.config,
		log:          logZap,
		restMap:      make(map[string]IRest),
		modelMap:     make(map[string]IModel),
		serviceMap:   make(map[string]IService),
		componentMap: make(map[string]IComponent),
		db:           db,
		transaction:  NewTransaction(db),
	}
	w.context.addComponent(w.component...)
	w.context.addModel(w.models...)
	w.context.AddService(w.services...)
	for _, service := range w.services {
		service.Init(w.context)
	}
	port := cast.ToInt(w.config.GetStringOrDefault("web.server.port", "9009"))
	rootGroup := newRestGroup(port).AddRest(w.rests...).Authentication(w.authentication)
	hasRootGroup := false
	for _, group := range w.restGroups {
		if group.port == 0 || group.port == port {
			group.merge(rootGroup)
			hasRootGroup = true
			break
		}
	}
	if !hasRootGroup {
		w.restGroups = append(w.restGroups, rootGroup)
	}
	for _, group := range w.restGroups {
		group.engine = w.getEngine(group.port, logZap)
		for _, rest := range group.rests {
			w.context.AddRest(rest)
		}
	}
	for _, group := range w.restGroups {
		context := w.context.Copy(group.digestAuth, group.engine.engine)
		for _, rest := range group.rests {
			rest.Init(context)
		}
	}

	if len(w.engines) > 1 {
		for _, engine := range w.engines[1:] {
			go func() {
				err := engine.run()
				if err != nil {
					log.Panic("启动服务失败:", err)
					return
				}
			}()
		}
	}

	return w.engines[0].run()
}
