# Best Practices

This guide covers recommended patterns and practices when building applications with Go Web Frame.

## Project Structure

### Recommended Directory Layout

```
myapp/
├── main.go                 # Application entry point
├── config/                 # Configuration files
│   └── application.yml
├── controller/             # HTTP handlers / REST controllers
│   ├── user_controller.go
│   └── order_controller.go
├── service/                # Business logic layer
│   ├── user_service.go
│   └── order_service.go
├── model/                  # Data access layer
│   ├── user.go
│   ├── user_model.go
│   └── order_model.go
├── entity/                 # Domain entities
│   ├── user.go
│   └── order.go
├── filter/                 # HTTP filters / middleware
│   ├── auth_filter.go
│   └── logging_filter.go
├── runner/                 # Background tasks
│   └── cleanup_runner.go
├── util/                   # Application utilities
│   └── helper.go
├── certs/                  # SSL certificates (auto-generated)
├── logs/                   # Log files
└── www/                    # Static files (optional)
```

## Layer Separation

### Controller Layer (HTTP)

Controllers handle HTTP-specific concerns only:

```go
type UserController struct {
    core.IService
    userService *UserService
}

func (c *UserController) Init(ctx *core.Context) error {
    c.userService = wf.GetService[*UserService](ctx)
    ctx.Get("/users/:id", c.GetUser)
    ctx.Post("/users", c.CreateUser)
    return nil
}

func (c *UserController) GetUser(req *web.Request) (any, error) {
    id := req.ParamUint("id")
    return c.userService.GetUserById(id)
}

func (c *UserController) CreateUser(req *web.Request) (any, error) {
    var input CreateUserInput
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }
    return c.userService.CreateUser(&input)
}
```

### Service Layer (Business Logic)

Services contain business rules and orchestrate models:

```go
type UserService struct {
    core.IService
    userModel    *UserModel
    emailService *EmailService
}

func (s *UserService) Init(ctx *core.Context) error {
    s.userModel = wf.GetModel[*UserModel](ctx)
    s.emailService = wf.GetService[*EmailService](ctx)
    return nil
}

func (s *UserService) CreateUser(input *CreateUserInput) (*User, error) {
    // Business validation
    if !s.isValidEmail(input.Email) {
        return nil, errors.New("invalid email format")
    }
    
    // Check uniqueness
    exists, _ := s.userModel.Query().Where("email = ?", input.Email).Count()
    if exists > 0 {
        return nil, errors.New("email already registered")
    }
    
    // Create user
    user := &User{
        Name:  input.Name,
        Email: input.Email,
    }
    if err := s.userModel.Save(user); err != nil {
        return nil, err
    }
    
    // Send welcome email (fire and forget)
    go s.emailService.SendWelcome(user.Email)
    
    return user, nil
}
```

### Model Layer (Data Access)

Models handle database operations only:

```go
type UserModel struct {
    *model.EntryModel[*User, uint]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](db, "t_user")
    return m.CreateTable()
}

// Add custom queries
func (m *UserModel) FindActiveUsers() ([]*User, error) {
    return m.Query().Where("status = ?", StatusActive).Order("created_at desc").All()
}

func (m *UserModel) FindByEmail(email string) (*User, error) {
    return m.Query().Where("email = ?", email).One()
}
```

## Error Handling

### Define Custom Errors

```go
var (
    ErrUserNotFound    = errors.New("user not found")
    ErrInvalidInput    = errors.New("invalid input")
    ErrUnauthorized    = errors.New("unauthorized")
)

type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string {
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}
```

### Use Errors in Handlers

```go
func (c *UserController) GetUser(req *web.Request) (any, error) {
    user, err := c.userService.GetUserById(id)
    if err != nil {
        return nil, &AppError{
            Code:    "USER_NOT_FOUND",
            Message: "User not found",
            Err:     err,
        }
    }
    return user, nil
}
```

### Custom Converter for Error Responses

```go
type APIConverter struct{}

func (c *APIConverter) Request(fc web.FilterChain, req *web.Request) {
    result, err := fc.Next()
    if err != nil {
        var appErr *AppError
        if errors.As(err, &appErr) {
            req.Response().AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                "code":    appErr.Code,
                "message": appErr.Message,
            })
            return
        }
        req.Response().AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "code":    "INTERNAL_ERROR",
            "message": err.Error(),
        })
        return
    }
    req.Response().AbortWithStatusJSON(http.StatusOK, gin.H{
        "code": "SUCCESS",
        "data": result,
    })
}
```

## Authentication & Authorization

### Using Route Metadata

```go
// Define metadata options
func RequireAuth() web.MetaOption {
    return web.WithValue("auth", true)
}

func RequireRole(role string) web.MetaOption {
    return web.WithValue("role", role)
}

func Public() web.MetaOption {
    return web.WithValue("public", true)
}

// Register routes with metadata
func (c *AdminController) Init(ctx *core.Context) error {
    ctx.Get("/login", c.Login).WithMeta(Public())
    ctx.Get("/dashboard", c.Dashboard).WithMeta(RequireAuth())
    ctx.Delete("/users/:id", c.DeleteUser).WithMeta(RequireAuth(), RequireRole("admin"))
    return nil
}
```

### Auth Filter

```go
type AuthFilter struct {
    core.IFilter
    jwtSecret string
}

func (f *AuthFilter) Init(ctx *core.Context) error {
    f.jwtSecret = ctx.GetConfig().GetString("jwt.secret")
    return nil
}

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    // Skip public routes
    if req.HasMeta(Public()) {
        return fc.Next()
    }

    // Check authentication
    if req.HasMeta(RequireAuth()) {
        token := req.GetHeader("Authorization")
        if token == "" {
            return nil, ErrUnauthorized
        }

        claims, err := f.validateToken(token)
        if err != nil {
            return nil, ErrUnauthorized
        }

        // Store user info in request context
        req.Set("user", claims)

        // Check role — use HandlerMeta().Get() for value retrieval
        requiredRole, _ := req.HandlerMeta().Get("role").(string)
        if requiredRole != "" && claims.Role != requiredRole {
            return nil, ErrForbidden
        }
    }

    return fc.Next()
}
```

## Configuration Management

### Environment-Specific Configs

```
config/
├── application.yml          # Base configuration
├── application-dev.yml      # Development overrides
├── application-prod.yml     # Production overrides
└── application-test.yml     # Test configuration
```

### Use Environment Variables

```yaml
# application.yml
db:
  type: mysql
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:3306}
  user: ${DB_USER:root}
  password: ${DB_PASSWORD:}
  database: ${DB_NAME:myapp}
```

### Access Config in Code

```go
func (s *MyService) Init(ctx *core.Context) error {
    s.maxRetries = ctx.GetConfig().GetInt("service.max_retries")
    s.timeout = ctx.GetConfig().GetDuration("service.timeout")
    return nil
}
```

## Context Propagation

### Propagate Request Context to Database

Always propagate the request context to database operations for cancellation and tracing:

```go
func (c *UserController) GetUser(req *web.Request) (any, error) {
    id := req.ParamUint("id")
    return c.userService.GetUserById(req, id)
}

func (s *UserService) GetUserById(req *web.Request, id uint) (*User, error) {
    // Inject request context once, all subsequent DB operations carry it
    return s.userModel.WithContext(req.Ctx()).FindByPK(id)
}
```

### Custom Context with Timeout

For slow queries, set a custom timeout:

```go
func (s *UserService) SearchUsers(req *web.Request, keyword string) ([]*User, error) {
    ctx, cancel := context.WithTimeout(req.Ctx(), 5*time.Second)
    defer cancel()
    return s.userModel.WithContext(ctx).Query().Where("name LIKE ?", "%"+keyword+"%").All()
}
```

### Builder-Level Context

Inject context directly on query/update/delete builders:

```go
users, err := userModel.Query().WithContext(req.Ctx()).Where("status = ?", 1).All()
err := userModel.Update().WithContext(req.Ctx()).Where("id = ?", 1).UpdateColumn("status", 0)
err := userModel.Delete().WithContext(req.Ctx()).Where("id = ?", 1).Delete()
```

## Database Best Practices

### Use Transactions for Multi-Step Operations

```go
func (s *OrderService) CreateOrder(input *CreateOrderInput) (*Order, error) {
    var order *Order
    
    tx := s.ctx.GetTransaction()
    err := tx.Execute(func(db *gorm.DB) error {
        // Step 1: Create order
        order = &Order{UserID: input.UserID, Total: input.Total}
        if err := db.Create(order).Error; err != nil {
            return err
        }
        
        // Step 2: Create order items
        for _, item := range input.Items {
            orderItem := &OrderItem{OrderID: order.Id, ProductID: item.ProductID}
            if err := db.Create(orderItem).Error; err != nil {
                return err
            }
        }
        
        // Step 3: Update inventory
        for _, item := range input.Items {
            if err := db.Model(&Product{}).
                Where("id = ?", item.ProductID).
                Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
                return err
            }
        }
        
        return nil
    })
    
    return order, err
}
```

### Index frequently queried fields

```go
type User struct {
    Id        uint      `gorm:"primaryKey"`
    Email     string    `gorm:"size:255;uniqueIndex"`  // Unique index
    Status    int       `gorm:"index"`                  // Regular index
    CompanyID uint      `gorm:"index:idx_company"`      // Named index
    CreatedAt time.Time `gorm:"index:idx_created"`      // For sorting
}
```

## Background Tasks

### Graceful Shutdown

```go
type EmailRunner struct {
    core.IRunner
    emailQueue chan *Email
}

func (r *EmailRunner) Run(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            // Drain remaining emails before exit
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

## Logging

### Structured Logging

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

## Testing

### Unit Test Services

```go
func TestUserService_CreateUser(t *testing.T) {
    // Setup
    mockModel := &MockUserModel{}
    service := &UserService{userModel: mockModel}
    
    // Execute
    user, err := service.CreateUser(&CreateUserInput{
        Name:  "Alice",
        Email: "alice@example.com",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

### Integration Test with HTTP Server

```go
func TestUserController_GetUser(t *testing.T) {
    // Setup
    app := setupTestApp()
    ts := httptest.NewServer(app.Engine())
    defer ts.Close()
    
    // Execute
    resp, err := http.Get(ts.URL + "/users/1")
    assert.NoError(t, err)
    defer resp.Body.Close()
    
    // Assert
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    var result map[string]any
    json.NewDecoder(resp.Body).Decode(&result)
    assert.Equal(t, float64(1), result["id"])
}
```

## Security

### Input Validation

```go
type CreateUserInput struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gte=0,lte=150"`
}

func (c *UserController) CreateUser(req *web.Request) (any, error) {
    var input CreateUserInput
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }
    
    // Use validator component
    validator := wf.GetService[*validator.Validator](req.Context())
    if err := validator.Validate(input); err != nil {
        return nil, err
    }
    
    return c.userService.CreateUser(&input)
}
```

### Rate Limiting

```go
func main() {
    app := wf.NewWithAutoConfig()
    
    // Add rate limit filter
    rateLimiter := ratelimit.NewRateLimit()
    app.AddFilter(rateLimiter)
    
    app.Start()
}
```

## Performance Tips

1. **Use connection pooling** - Configure appropriate pool sizes
2. **Index database fields** - Add indexes for frequently queried columns
3. **Cache frequently accessed data** - Use the cache component
4. **Avoid N+1 queries** - Use GORM's Preload
5. **Use background tasks** - Offload heavy operations

```go
// Good: Use Preload to avoid N+1
users, err := userModel.Query().
    Preload("Profile").
    Preload("Orders").
    All()

// Bad: N+1 problem
users, _ := userModel.Query().All()
for _, u := range users {
    profile, _ := profileModel.FindByUserId(u.Id)  // N queries!
}
```