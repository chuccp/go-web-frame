# Core API

Core abstractions and the dependency injection container.

## WebFrame

`WebFrame` is the main application entry point.

### Factory Method

The framework uses the `Builder` pattern:

```go
// Load configuration
cfg, err := config.LoadSingleFileConfig("application.yml")
if err != nil {
    // handle error
}

// Create app with Builder and custom config
builder := wf.NewBuilder(cfg)
app := builder.Build()
```

### Registration Methods

Registration methods are called on `Builder`, then `Build()` produces a `WebFrame`:

```go
builder := wf.NewBuilder(config)

// Register REST controllers
builder.Rest(rest ...core.IRest)

// Register services
builder.Service(services ...core.IService)

// Register filters
builder.Filter(filters ...core.IFilter)

// Register models
builder.Model(models ...core.IModel)

// Register model groups
builder.ModelGroup(modelGroups ...core.IModelGroup)

// Register REST groups
builder.RestGroup(restGroups ...*core.RestGroup)

// Register background runners
builder.Runner(runners ...core.IRunner)

// Build the app
app := builder.Build()
```

### Start Methods

```go
// Start app (blocks until shutdown)
err := app.Start()

// Start with context (supports graceful shutdown)
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
err := app.Run(ctx)
```

`Run(ctx)` supports graceful shutdown of HTTP server and background runners via context cancellation.

### Test Method

`Test()` initializes the app without starting the HTTP server — ideal for unit tests. It runs the full initialization flow (config, models, services, components) and provides a usable `*core.Context`.

```go
// Basic usage
err := app.Test(func(ctx *core.Context) error {
    service := wf.GetService[*UserService](ctx)
    user, err := service.GetUserById(1)
    if err != nil {
        return err
    }
    // assertions...
    return nil
})
```

### GetHandler (HTTP Testing)

`GetHandler()` initializes the app and returns an `http.Handler` for use with `httptest.NewServer()`. Unlike `Test()`, it gives you a real HTTP handler to test full request/response cycles:

```go
app := builder.Build()
handler := app.GetHandler()
ts := httptest.NewServer(handler)
defer ts.Close()

// Set Host header with the configured port for correct routing
req, _ := http.NewRequest("GET", ts.URL+"/api/users", nil)
req.Host = "localhost:19009"
resp, _ := http.DefaultClient.Do(req)
```

| Method | Use case |
|--------|----------|
| `Test()` | Unit tests — access services/models directly, no HTTP |
| `GetHandler()` | Integration tests — real HTTP requests via `httptest` |

## Context

`core.Context` is the dependency injection container.

### Get Components

```go
// Get service
service := wf.GetService[*UserService](ctx)

// Get model
model := wf.GetModel[*UserModel](ctx)

// Get component
component := wf.GetService[*CacheComponent](ctx)

// Get runner
runner := wf.GetRunner[*EmailRunner](ctx)

// Get filter
filter := wf.GetFilter[*AuthFilter](ctx)
```

### Register Routes

```go
// Register routes in a controller
func (c *UserController) Init(context *core.Context) error {
    context.Get("/users", c.List)
    context.Get("/users/:id", c.Get)
    context.Post("/users", c.Create)
    context.Put("/users/:id", c.Update)
    context.Delete("/users/:id", c.Delete)
    return nil
}
```

### Get Config

```go
// Get config values
value := ctx.GetConfig().GetString("key")
value := ctx.GetConfig().GetInt("key")
value := ctx.GetConfig().GetBool("key")
```

### Get Transaction

```go
// Get default model group transaction
tx := ctx.GetTransaction()

// Get named model group transaction
tx := ctx.GetTransactionByName("user_group")
```

## Request

`web.Request` provides request handling methods.

### Path Parameters

```go
// Get path parameter (string)
id := req.Param("id")

// Get path parameter (int)
idInt := req.ParamInt("id")

// Get path parameter (uint)
idUint := req.ParamUint("id")
```

### Query Parameters

```go
// Get query parameter
keyword := req.Query("q")

// Query parameter with default
page := req.DefaultQuery("page", "1")

// Get JSON body
jsonObject, err := req.Json()

// Bind JSON to struct
var data MyStruct
err := req.BindJSON(&data)
```

### JsonObject

`req.Json()` returns a `JsonObject` type with convenient value accessors:

```go
jsonObject, err := req.Json()
if err != nil {
    return nil, err
}

// Get values
name := jsonObject.GetString("name")
age := jsonObject.GetInt("age")
price := jsonObject.GetFloat("price")
active := jsonObject.GetBool("active")
tags := jsonObject.GetSlice("tags")
subObj := jsonObject.GetObject("address")
city := subObj.GetString("city")
```

### Headers

```go
// Get header
token := req.GetHeader("Authorization")

// Get client IP
ip := req.ClientIP()
```

### Cookie

```go
// Get cookie helper
cookie := req.Cookie()

// Get cookie value
value := cookie.Get("session_id")

// Set cookie
cookie.Set("token", "xxx", 3600)  // name, value, maxAge
```

### Pagination

```go
// Get pagination parameters
page, err := req.Page()

// page.PageNo  - page number (starts from 1)
// page.PageSize - items per page
// page.LastId  - last record ID (for cursor pagination)
```

## Response

`web.Response` provides response handling methods.

### Set Response

```go
// Set status code
req.Response().WriteHeader(200)

// Set response header
req.Response().Header().Set("Content-Type", "application/json")

// Write response body
req.Response().Write([]byte("Hello"))
```

### Quick Methods

```go
// Return JSON response
req.Response().AbortWithStatusJSON(200, data)

// Return error response
req.Response().AbortWithError(err)
```

## Next Steps

- [Web API](web.md) - Web layer API
- [Model API](model.md) - Model layer API
