# 服务

服务层包含业务逻辑，编排模型和其他服务。

## 创建服务

### 基本结构

```go
package service

import (
	core "github.com/chuccp/go-web-frame/core"
	wf "github.com/chuccp/go-web-frame"
)

type UserService struct {
	core.IService
	userModel *model.UserModel
}

func (s *UserService) Init(context *core.Context) error {
	// 从 Context 获取模型
	s.userModel = wf.GetModel[*model.UserModel](context)
	return nil
}

func (s *UserService) GetUserById(id uint) (*entity.User, error) {
	return s.userModel.FindById(id)
}

func (s *UserService) GetAllUsers() ([]*entity.User, error) {
	return s.userModel.FindAll()
}
```

### 注册服务

在 `main.go` 中注册：

```go
package main

import (
	config "github.com/chuccp/go-web-frame/config"
	wf "github.com/chuccp/go-web-frame"
	"go.uber.org/zap"
	"myapp/service"
)

func createApp() (*wf.WebFrame, error) {
	// 加载配置
	fileConfig, err := config.LoadSingleFileConfig("application.yml")
	if err != nil {
		return nil, err
	}

	// 创建 Builder
	builder := wf.NewBuilder(fileConfig)

	// 注册服务
	builder.Service(&service.UserService{})

	// 注册模型
	builder.Model(&model.UserModel{})

	// 构建应用
	app := builder.Build()
	return app, nil
}

func main() {
	app, err := createApp()
	if err != nil {
		zap.L().Fatal("创建应用失败", zap.Error(err))
		return
	}
	err = app.Start()
	if err != nil {
		zap.L().Fatal("启动应用失败", zap.Error(err))
	}
}
```

## 依赖注入

### 获取模型

```go
type UserService struct {
	core.IService
	userModel    *model.UserModel
	profileModel *model.ProfileModel
}

func (s *UserService) Init(context *core.Context) error {
	s.userModel = wf.GetModel[*model.UserModel](context)
	s.profileModel = wf.GetModel[*model.ProfileModel](context)
	return nil
}
```

### 获取其他服务

```go
type OrderService struct {
	core.IService
	userService *service.UserService
	orderModel  *model.OrderModel
}

func (s *OrderService) Init(context *core.Context) error {
	s.userService = wf.GetService[*service.UserService](context)
	s.orderModel = wf.GetModel[*model.OrderModel](context)
	return nil
}
```

### 获取配置

```go
type EmailService struct {
	core.IService
	apiKey string
}

func (s *EmailService) Init(context *core.Context) error {
	s.apiKey = context.GetConfig().GetString("email.api_key")
	return nil
}
```

## 业务逻辑示例

### 创建用户（带验证）

```go
func (s *UserService) CreateUser(input *CreateUserInput) (*entity.User, error) {
	// 业务验证
	if !s.isValidEmail(input.Email) {
		return nil, errors.New("invalid email format")
	}
	
	// 检查唯一性
	exists, _ := s.userModel.Query().Where("email = ?", input.Email).Count()
	if exists > 0 {
		return nil, errors.New("email already registered")
	}
	
	// 创建用户
	user := &entity.User{
		Name:  input.Name,
		Email: input.Email,
	}
	if err := s.userModel.Save(user); err != nil {
		return nil, err
	}
	
	// 发送欢迎邮件（异步）
	go s.emailService.SendWelcome(user.Email)
	
	return user, nil
}
```

### 事务操作

```go
type OrderService struct {
	core.IService
	ctx        *core.Context
	orderModel *model.OrderModel
}

func (s *OrderService) Init(ctx *core.Context) error {
	s.ctx = ctx
	s.orderModel = wf.GetModel[*model.OrderModel](ctx)
	return nil
}

func (s *OrderService) CreateOrder(input *CreateOrderInput) (*entity.Order, error) {
	var order *entity.Order

	err := s.ctx.GetTransaction().Exec(func(tx *db.DB) error {
		orderModel := wf.GetReNewModel[*model.OrderModel](tx, s.ctx)
		// 步骤 1：创建订单
		order = &entity.Order{UserID: input.UserID, Total: input.Total}
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 步骤 2：创建订单项
		for _, item := range input.Items {
			orderItem := &entity.OrderItem{OrderID: order.Id, ProductID: item.ProductID}
			if err := tx.Create(orderItem).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return order, err
}
```

## 获取服务

### 在控制器中获取服务

```go
type UserController struct {
	core.IService
	userService *service.UserService
}

func (c *UserController) Init(context *core.Context) error {
	// 获取服务
	c.userService = wf.GetService[*service.UserService](context)
	
	// 注册路由
	context.Get("/users", c.List)
	context.Get("/users/:id", c.Get)
	
	return nil
}

func (c *UserController) List(req *web.Request) (any, error) {
	users, err := c.userService.GetAllUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (c *UserController) Get(req *web.Request) (any, error) {
	id := req.ParamUint("id")
	user, err := c.userService.GetUserById(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
```

### 在其他服务中获取服务

```go
type PaymentService struct {
	core.IService
	orderService *service.OrderService
	emailService *service.EmailService
}

func (s *PaymentService) Init(context *core.Context) error {
	s.orderService = wf.GetService[*service.OrderService](context)
	s.emailService = wf.GetService[*service.EmailService](context)
	return nil
}
```

## 完整示例

```go
package main

import (
	config "github.com/chuccp/go-web-frame/config"
	core "github.com/chuccp/go-web-frame/core"
	wf "github.com/chuccp/go-web-frame"
	"go.uber.org/zap"
	"myapp/controller"
	"myapp/model"
	"myapp/service"
)

func createApp() (*wf.WebFrame, error) {
	// 加载配置
	fileConfig, err := config.LoadSingleFileConfig("application.yml")
	if err != nil {
		return nil, err
	}

	// 创建 Builder
	builder := wf.NewBuilder(fileConfig)

	// 注册模型
	builder.Model(&model.UserModel{})

	// 注册服务
	builder.Service(&service.UserService{})

	// 注册 REST 控制器
	restGroupBuilder := wf.NewRestGroupBuilder()
	restGroupBuilder.Rest(&controller.UserController{})
	restGroupBuilder.Port(8081)
	restGroup := restGroupBuilder.Build()
	builder.RestGroup(restGroup)

	// 构建应用
	app := builder.Build()
	return app, nil
}

func main() {
	app, err := createApp()
	if err != nil {
		zap.L().Fatal("创建应用失败", zap.Error(err))
		return
	}
	err = app.Start()
	if err != nil {
		zap.L().Fatal("启动应用失败", zap.Error(err))
	}
}
```

## 下一步

- [过滤器/中间件](filter.md) - 添加横切关注点
- [配置](configuration.md) - 了解配置管理
