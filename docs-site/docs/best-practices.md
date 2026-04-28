# 最佳实践

本文档提供使用 Go Web Frame 的推荐模式和最佳实践。

## 项目结构

### 推荐的目录布局

```
myapp/
├── main.go                 # 应用入口
├── go.mod                  # Go 模块文件
├── go.sum
├── config/                 # 配置文件目录
│   └── application.yml
├── controller/             # HTTP 处理器 / REST 控制器
│   ├── user_controller.go
│   └── order_controller.go
├── service/                # 业务逻辑层
│   ├── user_service.go
│   └── order_service.go
├── model/                  # 数据访问层
│   ├── user.go
│   ├── user_model.go
│   └── order_model.go
├── entity/                 # 领域实体
│   ├── user.go
│   └── order.go
├── filter/                 # HTTP 过滤器 / 中间件
│   ├── auth_filter.go
│   └── logging_filter.go
├── runner/                 # 后台任务
│   └── cleanup_runner.go
├── util/                   # 应用工具函数
├── certs/                  # SSL 证书（自动生成）
├── logs/                   # 日志文件
└── www/                    # 静态文件（可选）
```

## 分层分离

### 控制器层（HTTP）

控制器只处理 HTTP 相关的关注点：

```go
type UserController struct {
    core.IService
    userService *UserService
}

func (c *UserController) Init(context *core.Context) error {
    c.userService = wf.GetService[*UserService](context)
    
    context.Get("/users", c.List)
    context.Get("/users/:id", c.Get)
    context.Post("/users", c.Create)
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
    return c.userService.GetUserById(id)
}
```

### 服务层（业务逻辑）

服务包含业务逻辑并编排模型：

```go
type UserService struct {
    core.IService
    userModel    *UserModel
    emailService *EmailService
}

func (s *UserService) Init(context *core.Context) error {
    s.userModel = wf.GetModel[*UserModel](context)
    s.emailService = wf.GetService[*EmailService](context)
    return nil
}

func (s *UserService) CreateUser(input *CreateUserInput) (*User, error) {
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
    user := &User{
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

### 模型层（数据访问）

模型处理数据库操作：

```go
type UserModel struct {
    *model.EntryModel[*User]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User](db, "t_user")
    return m.CreateTable()
}

// 添加自定义查询
func (m *UserModel) FindActiveUsers() ([]*User, error) {
    return m.Query().Where("status = ?", 1).Order("create_time desc").All()
}

func (m *UserModel) FindByEmail(email string) (*User, error) {
    return m.Query().Where("email = ?", email).One()
}
```

## 错误处理

### 定义自定义错误

```go
var (
    ErrUserNotFound    = errors.New("user not found")
    ErrInvalidInput    = errors.New("invalid input")
    ErrUnauthorized    = errors.New("unauthorized")
    ErrForbidden       = errors.New("forbidden")
)
```

### 在处理器中使用错误

```go
func (c *UserController) GetUser(req *web.Request) (any, error) {
    user, err := c.userService.GetUserById(id)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, ErrUserNotFound
    }
    return user, nil
}
```

### 自定义响应转换器

```go
type APIConverter struct{}

func (c *APIConverter) Request(fc web.FilterChain, req *web.Request) {
    result, err := fc.Next()
    if err != nil {
        // 返回标准错误响应
        req.Response().AbortWithStatusJSON(200, web.ErrorMessage(err.Error()))
        return
    }
    // 返回标准成功响应
    req.Response().AbortWithStatusJSON(200, web.Data(result))
}
```

通过注册转换器过滤器，可以统一所有 API 的响应格式：

```go
builder := wf.NewBuilder(cfg)
builder.Filter(&APIConverter{})
```

## 认证与授权

### 使用路由元数据

```go
// 定义元数据选项
func Public() web.MetaOption {
    return web.WithValue("public", true)
}

func RequireAuth() web.MetaOption {
    return web.WithValue("auth", true)
}

func RequireRole(role string) web.MetaOption {
    return web.WithValue("role", role)
}

// 注册带元数据的路由
func (c *AdminController) Init(context *core.Context) error {
    context.Get("/login", c.Login).WithMeta(Public())
    context.Get("/dashboard", c.Dashboard).WithMeta(RequireAuth())
    context.Delete("/users/:id", c.DeleteUser).WithMeta(RequireAuth(), RequireRole("admin"))
    return nil
}
```

### 认证过滤器

```go
type AuthFilter struct {
    core.IFilter
    jwtSecret string
}

func (f *AuthFilter) Init(context *core.Context) error {
    f.jwtSecret = context.GetConfig().GetString("jwt.secret")
    return nil
}

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    meta := req.HandlerMeta()
    
    // 跳过公开路由
    if meta.Has("public") {
        return fc.Next()
    }
    
    // 检查认证
    if meta.Has("auth") {
        token := req.GetHeader("Authorization")
        if token == "" {
            return nil, ErrUnauthorized
        }
        
        claims, err := f.validateToken(token)
        if err != nil {
            return nil, ErrUnauthorized
        }
        
        // 存储用户信息到请求上下文
        req.Set("user", claims)
        
        // 检查角色
        requiredRole, _ := meta.Get("role").(string)
        if requiredRole != "" && claims.Role != requiredRole {
            return nil, ErrForbidden
        }
    }
    
    return fc.Next()
}
```

## 配置管理

### 使用环境变量

```yaml
# application.yml
web:
  db:
    type: ${DB_TYPE:mysql}
    host: ${DB_HOST:localhost}
    port: ${DB_PORT:3306}
    user: ${DB_USER:root}
    password: ${DB_PASSWORD:}
    database: ${DB_NAME:myapp}
```

### 在代码中访问配置

```go
type UserService struct {
    core.IService
    maxRetries int
    timeout    time.Duration
}

func (s *UserService) Init(context *core.Context) error {
    s.maxRetries = context.GetConfig().GetInt("service.max_retries")
    s.timeout = context.GetConfig().GetDuration("service.timeout")
    return nil
}
```

## 数据库最佳实践

### 使用事务处理多步操作

```go
func (s *OrderService) CreateOrder(input *CreateOrderInput) (*Order, error) {
    var order *Order

    tx := s.ctx.GetTransaction()
    err := tx.Exec(func(tx *db.DB) error {
        // 使用 GetReNewModel 在事务中创建模型
        orderModel := wf.GetReNewModel[*OrderModel](tx, s.ctx)

        // 步骤 1：创建订单
        order = &Order{UserID: input.UserID, Total: input.Total}
        if err := orderModel.Save(order); err != nil {
            return err
        }

        // 步骤 2：创建订单项
        itemModel := wf.GetReNewModel[*OrderItemModel](tx, s.ctx)
        for _, item := range input.Items {
            orderItem := &OrderItem{OrderID: order.Id, ProductID: item.ProductID}
            if err := itemModel.Save(orderItem); err != nil {
                return err
            }
        }

        return nil
    })

    return order, err
}
```

### 为频繁查询的字段添加索引

```go
type User struct {
    Id        uint      `gorm:"primaryKey"`
    Email     string    `gorm:"size:255;uniqueIndex"`  // 唯一索引
    Status    int       `gorm:"index"`                  // 常规索引
    CompanyID uint      `gorm:"index:idx_company"`      // 命名索引
    CreatedAt time.Time `gorm:"index:idx_created"`      // 用于排序
}
```

## 后台任务

### 优雅关闭

```go
type EmailRunner struct {
    core.IRunner
    emailQueue chan *Email
}

func (r *EmailRunner) Init(context *core.Context) error {
    r.emailQueue = make(chan *Email, 100)
    return nil
}

func (r *EmailRunner) Run() error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-r.ctx.Done():
            // 在退出前处理剩余邮件
            r.drainQueue()
            return nil
        case email := <-r.emailQueue:
            if err := r.sendEmail(email); err != nil {
                log.Error("failed to send email", zap.Error(err))
            }
        }
    }
}
```

## 日志

### 结构化日志

```go
func (s *UserService) CreateUser(input *CreateUserInput) (*User, error) {
    log.Info("creating user",
        zap.String("email", input.Email),
        zap.String("name", input.Name),
    )
    
    user, err := s.doCreateUser(input)
    if err != nil {
        log.Error("failed to create user",
            zap.String("email", input.Email),
            zap.Error(err),
        )
        return nil, err
    }
    
    log.Info("user created",
        zap.Uint("id", user.Id),
        zap.String("email", user.Email),
    )
    
    return user, nil
}
```

## 测试

### 单元测试服务

```go
func TestUserService_CreateUser(t *testing.T) {
    // 设置
    mockModel := &MockUserModel{}
    service := &UserService{userModel: mockModel}
    
    // 执行
    user, err := service.CreateUser(&CreateUserInput{
        Name:  "Alice",
        Email: "alice@example.com",
    })
    
    // 断言
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

### 集成测试

使用 `WebFrame.Test()` 方法进行集成测试（不启动 HTTP 服务器）：

```go
func TestUserController_GetUser(t *testing.T) {
    // 加载测试配置
    cfg, err := config.LoadSingleFileConfig("test.yml")
    if err != nil {
        t.Fatal(err)
    }

    // 构建应用
    builder := wf.NewBuilder(cfg)
    builder.Model(&model.UserModel{})
    builder.Service(&service.UserService{})
    builder.Rest(&controller.UserController{})
    app := builder.Build()

    // 使用 Test 方法初始化并运行测试
    err = app.Test(func(ctx *core.Context) error {
        userService := wf.GetService[*service.UserService](ctx)
        user, err := userService.GetUserById(1)
        if err != nil {
            return err
        }
        assert.Equal(t, "test_user", user.Name)
        return nil
    })
    assert.NoError(t, err)
}
```

## 性能技巧

1. **使用连接池** - 配置合适的池大小
2. **为数据库字段添加索引** - 为频繁查询的列添加索引
3. **缓存频繁访问的数据** - 使用缓存组件
4. **避免 N+1 查询** - 使用 GORM 的 Preload
5. **使用后台任务** - 卸载繁重操作

```go
// 好：使用 Preload 避免 N+1
users, err := userModel.Query().
    Preload("Profile").
    Preload("Orders").
    All()

// 坏：N+1 问题
users, _ := userModel.Query().All()
for _, u := range users {
    profile, _ := profileModel.FindByUserId(u.Id)  // N 次查询！
}
```

## 下一步

- [API 参考](../api/core.md) - 查看完整的 API 文档
- [高级主题](../advanced/database.md) - 了解更多高级功能
