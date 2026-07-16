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
// Note: One() returns zero value + nil error when record not found (not gorm.ErrRecordNotFound).
// Check the return value to determine if a record exists:
// if user == nil { /* not found */ }
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

!!! info "EntryModel vs Raw GORM"
    `EntryModel` and `Model` are convenience wrappers that reduce boilerplate code. They are built on top of [GORM](https://gorm.io/docs/) — you can skip these wrappers entirely and use GORM's native API directly:

    ```go
    // Option 1: Use EntryModel (recommended for common CRUD, less boilerplate)
    user, err := userModel.FindByPK(1)

    // Option 2: Use GORM directly (more flexible for complex queries)
    // Get *gorm.DB via GetGorm() and use the full GORM ecosystem
    var user User
    err := db.GetGorm().First(&user, 1).Error
    ```

    When EntryModel's built-in methods are insufficient (complex JOINs, subqueries, window functions, etc.), using the GORM ecosystem directly is the best choice. You can mix both approaches freely — the framework imposes no restrictions.

## Query Builder

```go
userModel.Query().
    Where("status = ?", 1).
    Order("id desc").
    List(10)
```

## Pagination

Pagination parameters are auto-normalized (`PageNo` defaults to 1, `PageSize` defaults to 10 via `util.DefaultPage`):

```go
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := userModel.Query().Page(page)

// Web pagination
pageAble, err := userModel.Query().PageForWeb(page)
```

## Preload (Eager Loading)

`Preload` loads associations. The framework automatically sets the GORM `Model()` clause to ensure associations resolve correctly:

```go
users, err := userModel.Query().Preload("Profile").Preload("Role").All()
user, err := userModel.Query().Where("id = ?", 1).Preload("Profile").One()
```

## Joins

`Joins` also triggers automatic `Model()` clause setup (fixed in v1.0.14):

```go
// Association JOIN (GORM resolves the relation name)
users, err := userModel.Query().Joins("Profile").Where("status = ?", 1).All()

// Raw JOIN
users, err := userModel.Query().
    Joins("JOIN orders ON orders.user_id = t_user.id").
    Where("status = ?", 1).
    All()
```

> **Note**: Both `Preload` and `Joins` automatically call `Table.Model(entry)` — no manual setup needed.

## Aggregate Queries

Use `Aggregate()` for SQL aggregate functions (`SUM`, `COUNT`, `AVG`, `MAX`, `MIN`):

```go
// Scalar aggregate — scan into a single value
var total float64
err := orderModel.Aggregate().
    Select("SUM(amount)").
    Where("status = ?", 1).
    Aggregate(&total)

var count int
err := orderModel.Aggregate().
    Select("COUNT(*)").
    Aggregate(&count)
```

### GROUP BY

Grouped aggregates scan into a slice of custom structs:

```go
type CategoryStat struct {
    Category string  `json:"category"`
    Total    float64 `json:"total"`
    Count    int     `json:"count"`
}

var stats []CategoryStat
err := orderModel.Aggregate().
    Select("category, SUM(amount) as total, COUNT(*) as count").
    Group("category").
    Aggregate(&stats)
```

### HAVING

```go
var stats []CategoryStat
err := orderModel.Aggregate().
    Select("category, SUM(amount) as total").
    Group("category").
    Having("SUM(amount) > ?", 200).
    Aggregate(&stats)
```

### DISTINCT

```go
// SELECT DISTINCT category ...
var categories []struct{ Category string }
err := orderModel.Aggregate().
    Distinct("category").
    Aggregate(&categories)

// COUNT(DISTINCT category)
var cnt int
err := orderModel.Aggregate().
    Select("COUNT(DISTINCT category)").
    Aggregate(&cnt)
```

> **Tip**: `Aggregate(result)` auto-detects the result type — scalar values (e.g. `*float64`) get `LIMIT 1`, slice values (e.g. `*[]Stat`) return all matching rows.