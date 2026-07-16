# Model API

Generic ORM API.

## Model[T]

Base generic model.

```go
type UserModel struct {
    *model.Model[*User]
}

// Initialize
m.Model = model.NewModel[*User](db, "t_user")
```

## Query Operations

```go
// All records
users, err := m.Query().All()

// Single record
// Note: returns zero value + nil error when not found (not gorm.ErrRecordNotFound)
user, err := m.Query().Where("id = ?", 1).One()

// Count
count, err := m.Query().Count()

// With conditions
users, err := m.Query().Where("status = ?", 1).Order("id desc").All()

// Pagination (PageNo defaults to 1, PageSize defaults to 10 via util.DefaultPage)
users, total, err := m.Query().Page(page)

// Limit
users, err := m.Query().List(10)

// Preload (GORM eager loading — auto-sets Model() clause)
users, err := m.Query().Preload("Profile").All()

// Joins (also auto-sets Model() clause, v1.0.14)
users, err := m.Query().Joins("JOIN orders ON orders.user_id = t_user.id").All()
```

## Update Operations

```go
err := m.Update().Where("id = ?", 1).UpdateColumn("name", "bob")

err := m.Update().Where("id = ?", 1).UpdateForMap(map[string]any{
    "name": "bob",
    "status": 1,
})
```

## Delete Operations

```go
err := m.Delete().Where("id = ?", 1).Delete()
```

## Save

```go
user := &User{Name: "alice"}
err := m.Save(user)
// user.Id is populated after Save
```

## EntryModel[T, PK]

Extended model for entities with a primary key. `PK` is the primary key type (`uint`, `int`, `string`, etc.).

```go
type UserModel struct {
    *model.EntryModel[*User, uint]
}

// Additional methods
user, err := m.FindByPK(1)
users, err := m.FindAll()
err := m.DeleteByPK(1)
err := m.UpdateByPK(user)
pageAble, err := m.Query().PageForWeb(page)
err := m.UpdateColumn(1, "name", "bob")
```

## Aggregate[T]

Aggregate query builder for `SUM`, `COUNT`, `AVG`, `MAX`, `MIN`, `GROUP BY`, `HAVING`, `DISTINCT`. Created via `model.Aggregate()`.

### Methods

| Method | Description |
|--------|-------------|
| `Select(query, args...)` | Aggregate expression (e.g. `"SUM(amount)"`, `"category, COUNT(*) as cnt"`) |
| `Where(query, args...)` | WHERE condition |
| `Group(name)` | GROUP BY clause |
| `Having(query, args...)` | HAVING condition (used with Group) |
| `Order(query)` | ORDER BY clause |
| `Joins(query)` | JOIN clause |
| `Distinct(args...)` | DISTINCT (no args = all columns, pass column names for specific) |
| `Aggregate(result)` | Execute and scan into result (scalar gets LIMIT 1, slice returns all rows) |
| `WithContext(ctx)` | Set context (returns shallow copy) |

### Scalar Aggregates

```go
// SUM
var total float64
err := orderModel.Aggregate().Select("SUM(amount)").Where("status = ?", 1).Aggregate(&total)

// COUNT
var count int
err := orderModel.Aggregate().Select("COUNT(*)").Aggregate(&count)

// Expression with args
var result float64
err := orderModel.Aggregate().Select("SUM(amount) * ? + ?", 2, 100).Aggregate(&result)
```

### Grouped Aggregates

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
    Having("SUM(amount) > ?", 200).
    Order("total desc").
    Aggregate(&stats)
```

## Transaction

`model.Transaction` provides transaction support.

### Basic Transaction

```go
tx := ctx.GetTransaction()
err := tx.Exec(func(tx *db.DB) error {
    // Execute operations within the transaction
    userModel := wf.GetReNewModel[*UserModel](tx, ctx)
    return userModel.Save(user)
})
```

### Named Transaction

```go
tx := ctx.GetTransactionByName("user_group")
err := tx.Exec(func(tx *db.DB) error {
    // Execute operations within a named transaction
    userModel := wf.GetReNewModel[*UserModel](tx, ctx)
    return userModel.Save(user)
})
```

## Next Steps

- [Core API](core.md) - Core API reference
- [Web API](web.md) - Web layer API