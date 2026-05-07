# Service

Business logic layer and dependency injection.

## Define Service

```go
type UserService struct {
    core.IService
    userModel *UserModel
}

func (s *UserService) Init(ctx *core.Context) error {
    s.userModel = wf.GetModel[*UserModel](ctx)
    return nil
}

func (s *UserService) GetUserById(id uint) (*User, error) {
    return s.userModel.FindById(id)
}
```

## Register Service

```go
builder.Service(&UserService{})
```

## Get Service

```go
userService := wf.GetService[*UserService](ctx)
user, err := userService.GetUserById(1)
```

## Service Types

| Type | Method | Retrieve |
|------|--------|----------|
| Service | `builder.Service(&MyService{})` | `wf.GetService[*MyService](ctx)` |
| Model | `builder.Model(&MyModel{})` | `wf.GetModel[*MyModel](ctx)` |
| REST | `builder.Rest(&MyController{})` | - |
| Filter | `builder.Filter(&MyFilter{})` | - |
| Runner | `builder.Runner(&MyRunner{})` | - |