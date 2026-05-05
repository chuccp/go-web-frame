package core

import (
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/web"
)

// IService 是所有需要初始化的服务的基接口
//
// 实现此接口的服务会自动被框架的依赖注入容器管理，并在应用启动时调用 Init 方法进行初始化。
//
// 使用示例：
//
//	type UserService struct {
//	    core.IService
//	    // 添加依赖
//	}
//
//	func (s *UserService) Init(ctx *Context) error {
//	    // 初始化逻辑
//	    return nil
//	}
type IService interface {
	// Init 在应用启动时被调用，用于初始化服务
	// ctx: 依赖注入上下文，可以从中获取其他服务
	// 返回 error: 初始化失败时返回错误
	Init(ctx *Context) error
}

// IModel 是数据访问层的接口，提供数据库操作能力
//
// 实现此接口的模型会自动被注册到依赖注入容器中，并支持数据库表管理。
//
// 使用示例：
//
//	type UserModel struct {
//	    *model.Model[*User]
//	}
//
//	func (m *UserModel) Init(db *db.DB, c *Context) error {
//	    m.Model = model.NewModel[*User](db, "t_user")
//	    return m.CreateTable()
//	}
type IModel interface {
	// Init 初始化模型，绑定数据库和上下文
	// db: 数据库连接
	// c: 依赖注入上下文
	Init(db *db.DB, c *Context) error

	// IsExist 检查表是否存在
	// 返回 (bool, error): 表是否存在，以及可能的错误
	IsExist() (bool, error)

	// CreateTable 创建数据库表
	// 返回 error: 创建失败时返回错误
	CreateTable() error

	// DeleteTable 删除数据库表
	// 返回 error: 删除失败时返回错误
	DeleteTable() error

	// GetTableName 获取表名
	// 返回 string: 表名
	GetTableName() string

	// ReNew 创建一个新的模型实例，用于事务处理
	// db: 新的数据库连接
	// c: 依赖注入上下文
	// 返回 IModel: 新的模型实例
	ReNew(db *db.DB, c *Context) IModel
}

// IRest 是 REST 控制器的接口，继承自 IService
//
// 实现此接口的控制器可以方便地注册 RESTful API 路由。
//
// 使用示例：
//
//	type UserController struct {
//	    core.IRest
//	}
//
//	func (c *UserController) Init(ctx *Context) error {
//	    ctx.Get("/users", c.GetUsers)
//	    ctx.Get("/users/:id", c.GetUser)
//	    ctx.Post("/users", c.CreateUser)
//	    return nil
//	}
type IRest interface {
	IService
}

// IRunner 是后台任务的接口，继承自 IService
//
// 实现此接口的任务会在应用启动后在后台运行，框架通过 context pool 管理生命周期。
// Run 方法不再接收参数，生命周期由 server 的 context pool 控制。
// 如需依赖查找，可在 Init 中缓存 *Context。
// 适用于定时任务、消息消费者、数据同步等场景。
//
// 使用示例：
//
//	type BackgroundTask struct {
//	    core.IRunner
//	    ctx *core.Context
//	}
//
//	func (t *BackgroundTask) Init(ctx *Context) error {
//	    t.ctx = ctx
//	    return nil
//	}
//
//	func (t *BackgroundTask) Run() error {
//	    ticker := time.NewTicker(5 * time.Second)
//	    defer ticker.Stop()
//
//	    for {
//	        select {
//	        case <-ticker.C:
//	            // 执行任务，需要依赖查找时使用 t.ctx
//	        }
//	    }
//	}
type IRunner interface {
	IService
	// Run 运行后台任务，生命周期由 server 的 context pool 控制
	// 返回 error: 任务失败时返回错误
	Run() error
}

// IModelGroup 是模型组的接口，用于管理多个相关的模型
//
// 模型组可以批量管理多个模型，支持自动创建表、事务管理等。
//
// 使用示例：
//
//	userModel := &UserModel{}
//	group := app.NewModelGroup(db, "user_group")
//	group.AddModel(userModel)
//	group.AutoCreateTable(true)
type IModelGroup interface {
	IService
	// AddModel 添加模型到组中
	AddModel(model ...IModel)

	// GetModel 获取组中的所有模型
	// 返回 []IModel: 模型列表
	GetModel() []IModel

	// AutoCreateTable 设置是否自动创建表
	// autoCreateTable: true 表示自动创建表
	AutoCreateTable(autoCreateTable bool)

	// SwitchDB 切换数据库连接
	// db: 新的数据库连接
	// context: 依赖注入上下文
	// 返回 error: 切换失败时返回错误
	SwitchDB(db *db.DB, context *Context) error

	// SetDefaultDB 设置默认数据库连接
	// db: 数据库连接
	SetDefaultDB(db *db.DB)

	// Name 获取模型组名称
	// 返回 string: 组名
	Name() string

	// GetTransaction 获取事务管理器
	// 返回 *model.Transaction: 事务管理器
	GetTransaction() *model.Transaction
}

// IFilter 是 HTTP 过滤器/中间件的接口，继承自 IService
//
// 实现此接口的过滤器会在请求处理链中执行，用于横切关注点如认证、日志、限流等。
//
// 使用示例：
//
//	type AuthFilter struct {
//	    core.IFilter
//	}
//
//	func (f *AuthFilter) Init(ctx *Context) error {
//	    return nil
//	}
//
//	func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
//	    token := req.GetHeader("Authorization")
//	    if token == "" {
//	        return nil, errors.New("unauthorized")
//	    }
//	    return fc.Next()
//	}
type IFilter interface {
	IService
	web.Filter
}

// IConverter 是响应转换器的接口，继承自 IService
//
// 实现此接口的转换器负责将服务返回的数据转换为 HTTP 响应。
// 可以自定义响应格式，如 JSON、XML、Protocol Buffers 等。
//
// 使用示例：
//
//	type CustomConverter struct {
//	    core.IConverter
//	}
//
//	func (c *CustomConverter) Init(ctx *Context) error {
//	    return nil
//	}
//
//	func (c *CustomConverter) Request(fc web.FilterChain, req *web.Request) {
//	    result, err := fc.Next()
//	    // 自定义响应处理
//	}
type IConverter interface {
	IService
	web.Converter
}
