package core

import (
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/web"
)

// IService is the base interface for all framework services.
//
// Services implementing this interface are automatically managed by the DI container,
// and Init is called during application startup.
//
// Example:
//
//	type UserService struct {
//	    core.IService
//	}
//
//	func (s *UserService) Init(ctx *Context) error {
//	    // initialization logic
//	    return nil
//	}
type IService interface {
	// Init is called during application startup to initialize the service.
	// ctx: dependency injection context for accessing other services.
	// Returns error if initialization fails.
	Init(ctx *Context) error
}

// IModel is the data access layer interface providing database operations.
//
// Models implementing this interface are automatically registered in the DI container.
//
// Example:
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
	// Init initializes the model with the given database and context.
	Init(db *db.DB, c *Context) error

	// IsExist reports whether the database table exists.
	IsExist() (bool, error)

	// CreateTable creates the database table if it does not exist.
	CreateTable() error

	// DeleteTable drops the database table.
	DeleteTable() error

	// GetTableName returns the database table name.
	GetTableName() string

	// ReNew creates a new model instance with the given database connection,
	// useful for transaction handling.
	ReNew(db *db.DB, c *Context) IModel
}

// IRest is the REST controller interface, extending IService.
//
// Example:
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

// IRunner is the background task interface, extending IService.
//
// Tasks implementing this interface run in the background after application startup.
// The lifecycle is managed by the server's context pool.
// Suitable for scheduled tasks, message consumers, data synchronization, etc.
//
// Example:
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
//	    for {
//	        select {
//	        case <-ticker.C:
//	            // perform task
//	        }
//	    }
//	}
type IRunner interface {
	IService
	// Run executes the background task. Lifecycle is controlled by the server's context pool.
	Run() error
}

// IModelGroup is the interface for managing collections of related models.
//
// Model groups can batch-manage multiple models, supporting auto table creation
// and transaction management.
//
// Example:
//
//	userModel := &UserModel{}
//	group := app.NewModelGroup(db, "user_group")
//	group.AddModel(userModel)
//	group.AutoCreateTable(true)
type IModelGroup interface {
	IService
	// AddModel adds one or more models to the group.
	AddModel(model ...IModel)

	// GetModel returns all models in this group.
	GetModel() []IModel

	// AutoCreateTable sets whether tables should be auto-created during init.
	AutoCreateTable(autoCreateTable bool)

	// SwitchDB replaces the database connection and reinitializes all models.
	SwitchDB(db *db.DB, context *Context) error

	// SetDefaultDB sets the default database connection for this group.
	SetDefaultDB(db *db.DB)

	// Name returns the model group name.
	Name() string

	// GetTransaction returns the transaction manager for this group.
	GetTransaction() *model.Transaction
}

// IFilter is the HTTP filter/middleware interface, extending IService.
//
// Filters are executed in the request processing chain for cross-cutting concerns
// such as authentication, logging, and rate limiting.
//
// Example:
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

// IConverter is the response converter interface, extending IService.
//
// Converters transform service return values into HTTP responses.
// Custom formats like JSON, XML, Protocol Buffers can be implemented.
//
// Example:
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
//	    // custom response handling
//	}
type IConverter interface {
	IService
	web.Converter
}
