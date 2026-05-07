# Model

Type-safe generic ORM in Go Web Frame.

## Define Entity

```go
type User struct {
    Id         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name       string    `gorm:"size:255" json:"name"`
    CreateTime time.Time `json:"createTime"`
}
```

## Define Model

```go
type UserModel struct {
    *model.Model[*User]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.Model = model.NewModel[*User](db, "t_user")
    return m.CreateTable()
}
```

## Register Model

```go
builder.Model(&UserModel{})
```

## CRUD Operations

```go
userModel := wf.GetModel[*UserModel](ctx)

// Query
users, err := userModel.Query().Where("id > ?", 10).All()
user, err := userModel.Query().Where("id = ?", 1).One()
count, err := userModel.Query().Count()

// Save
user := &User{Name: "alice"}
err := userModel.Save(user)

// Update
err := userModel.Update().Where("id = ?", 1).UpdateColumn("name", "bob")

// Delete
err := userModel.Delete().Where("id = ?", 1).Delete()
```

## EntryModel

For entities with `Id`, `CreateTime`, `UpdateTime` fields:

```go
type UserModel struct {
    *model.EntryModel[*User]
}

// Additional methods:
user, err := userModel.FindById(1)
users, err := userModel.FindAll()
err := userModel.DeleteById(1)
pageAble, err := userModel.Query().PageForWeb(page)
```

## Query Builder

```go
userModel.Query().
    Where("status = ?", 1).
    Order("id desc").
    Limit(10).
    All()
```

## Pagination

```go
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := userModel.Query().Page(page)

// Web pagination
pageAble, err := userModel.Query().PageForWeb(page)
```