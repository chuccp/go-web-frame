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

For entities with a primary key. `EntryModel` accepts two type parameters: `T` (entity type) and `PK` (primary key type).

```go
type UserModel struct {
    *model.EntryModel[*User, uint]
}

// Additional methods:
user, err := userModel.FindByPK(1)
users, err := userModel.FindAll()
err := userModel.DeleteByPK(1)
pageAble, err := userModel.Query().PageForWeb(page)
```

> **EntryModel vs Raw GORM**
> 
> `EntryModel` and `Model` are convenience wrappers that reduce boilerplate code. They are built on top of GORM — you can skip these wrappers entirely and use GORM's native API directly:
> 
> ```go
> // Option 1: Use EntryModel (recommended for common CRUD, less boilerplate)
> user, err := userModel.FindByPK(1)
> 
> // Option 2: Use GORM directly (more flexible for complex queries)
> var user User
> err := db.DB.First(&user, 1).Error
> ```
> 
> When EntryModel's built-in methods are insufficient (complex JOINs, subqueries, window functions, etc.), using the GORM ecosystem directly is the best choice. You can mix both approaches freely — the framework imposes no restrictions.

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