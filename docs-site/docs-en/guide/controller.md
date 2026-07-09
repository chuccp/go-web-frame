# Controller

REST controllers in Go Web Frame.

## Define Controller

```go
type UserController struct {
    core.IService
}

func (u *UserController) Init(ctx *core.Context) error {
    ctx.Get("/users", u.List)
    ctx.Post("/users", u.Create)
    return nil
}
```

## Register Controller

```go
builder := wf.NewBuilder(config.LoadAutoConfig())
builder.Rest(&UserController{})
```

## Request Handling

```go
func (u *UserController) List(c *web.Request) (any, error) {
    // Path params
    id := c.Param("id")
    
    // Query params
    page := c.Query("page")
    
    // JSON body
    var data struct {
        Name string `json:"name"`
    }
    if err := c.BindJSON(&data); err != nil {
        return nil, err
    }
    
    // Get service/model
    userModel := wf.GetModel[*UserModel](c.Context())
    
    return userModel.Query().All()
}
```

## Request Methods

| Method | Description |
|--------|-------------|
| `Param(key)` | Path parameter (string) |
| `ParamInt(key)` | Path parameter (int) |
| `Query(key)` | Query parameter (string) |
| `QueryInt(key)` | Query parameter (int) |
| `BindJSON(v)` | Bind JSON body |
| `GetHeader(key)` | Get header |
| `Cookie()` | Get cookie helper |
| `Context()` | Get core.Context |
| `Page()` | Get pagination info |

## Response Types

```go
// JSON (default)
return map[string]any{"id": 1}, nil

// String
return "hello", nil

// File download
return &web.File{Path: "/path/file.pdf", FileName: "doc.pdf"}, nil

// Redirect
return web.Redirect("/new-url"), nil

// Error
return nil, errors.New("something wrong")
```

## File Upload

```go
func (u *UserController) Upload(c *web.Request) (any, error) {
    file, header, err := c.File("file")  // form field name
    if err != nil {
        return nil, err
    }
    defer file.Close()

    dst := "./uploads/" + header.Filename
    if err := web.SaveUploadedFile(header, dst); err != nil {
        return nil, err
    }
    return map[string]string{"path": dst}, nil
}
```