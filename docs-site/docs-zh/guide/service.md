# 服务

服务层包含业务逻辑，编排模型和其他服务。

## 创建服务

### 基本结构

```go
package service

import (
    "github.com/chuccp/go-web-frame/core"
    wf "github.com/chuccp/go-web-frame"
    "myapp/entity"
    "myapp/model"
)

type UserService struct {
    core.IService
    userModel *model.UserModel
}

func (s *UserService) Init(ctx *core.Context) error {
    s.userModel = wf.GetModel[*model.UserModel](ctx)
    return nil
}

func (s *UserService) GetUserById(id uint) (*entity.User, error) {
    return s.userModel.FindByPK(id)
}

func (s *UserService) GetAllUsers() ([]*entity.User, error) {
    return s.userModel.FindAll()
}
```

### 注册服务

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "myapp/model"
    "myapp/service"
)

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)

    builder.Model(&model.UserModel{})
    builder.Service(&service.UserService{})

    builder.Build().Run(context.Background())
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

func (s *UserService) Init(ctx *core.Context) error {
    s.userModel = wf.GetModel[*model.UserModel](ctx)
    s.profileModel = wf.GetModel[*model.ProfileModel](ctx)
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

func (s *OrderService) Init(ctx *core.Context) error {
    s.userService = wf.GetService[*service.UserService](ctx)
    s.orderModel = wf.GetModel[*model.OrderModel](ctx)
    return nil
}
```

### 获取配置

```go
type EmailService struct {
    core.IService
    apiKey string
}

func (s *EmailService) Init(ctx *core.Context) error {
    s.apiKey = ctx.GetConfig().GetString("email.api_key")
    return nil
}
```

## 业务逻辑示例

### 创建用户

```go
func (s *UserService) CreateUser(input *CreateUserInput) (*entity.User, error) {
    if !s.isValidEmail(input.Email) {
        return nil, errors.New("invalid email format")
    }

    exists, _ := s.userModel.Query().Where("email = ?", input.Email).Count()
    if exists > 0 {
        return nil, errors.New("email already registered")
    }

    user := &entity.User{Name: input.Name, Email: input.Email}
    if err := s.userModel.Save(user); err != nil {
        return nil, err
    }

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
        order = &entity.Order{UserID: input.UserID, Total: input.Total}
        if err := orderModel.Save(order); err != nil {
            return err
        }
        for _, item := range input.Items {
            orderItem := &entity.OrderItem{OrderID: order.Id, ProductID: item.ProductID}
            if err := orderModel.Save(orderItem); err != nil {
                return err
            }
        }
        return nil
    })

    return order, err
}
```

## 在控制器中获取服务

```go
type UserController struct {
    core.IService
    userService *service.UserService
}

func (c *UserController) Init(ctx *core.Context) error {
    c.userService = wf.GetService[*service.UserService](ctx)
    ctx.Get("/users", c.List)
    ctx.Get("/users/:id", c.Get)
    return nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    return c.userService.GetAllUsers()
}

func (c *UserController) Get(req *web.Request) (any, error) {
    return c.userService.GetUserById(req.ParamUint("id"))
}
```

## 完整示例

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "myapp/controller"
    "myapp/model"
    "myapp/service"
)

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)

    builder.Model(&model.UserModel{})
    builder.Service(&service.UserService{})

    restGroup := wf.NewRestGroupBuilder().
        Rest(&controller.UserController{}).
        Port(8081).
        Build()
    builder.RestGroup(restGroup)

    builder.Build().Run(context.Background())
}
```

## 下一步

- [过滤器/中间件](filter.md) - 添加横切关注点
- [配置](configuration.md) - 了解配置管理
- [后台任务](runner.md) - Runner 和定时任务
