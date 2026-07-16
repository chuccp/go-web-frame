# Database

Advanced database usage.

## Transactions

### Basic Transaction

```go
func (s *OrderService) CreateOrder(input *CreateOrderInput) (*Order, error) {
    var order *Order

    tx := ctx.GetTransaction()
    err := tx.Exec(func(tx *db.DB) error {
        // Use GetReNewModel to create a model bound to the transaction
        orderModel := wf.GetReNewModel[*OrderModel](tx, ctx)

        // Step 1: create the order
        order = &Order{UserID: input.UserID, Total: input.Total}
        if err := orderModel.Save(order); err != nil {
            return err
        }

        // Step 2: create order items
        itemModel := wf.GetReNewModel[*OrderItemModel](tx, ctx)
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

### Named Transaction

```go
func (s *UserService) UpdateUserWithTransaction(user *User) error {
    tx := ctx.GetTransaction()
    return tx.Exec(func(tx *db.DB) error {
        userModel := wf.GetReNewModel[*UserModel](tx, ctx)
        return userModel.Save(user)
    })
}
```

## Model Groups

Share database connection across models:

```go
group := wf.NewModelGroupBuilder().
    Name("user_group").
    DB(db).
    Model(&UserModel{}, &ProfileModel{}).
    AutoCreateTable(true).
    Build()
builder.ModelGroup(group)
```

## Multiple Databases

```go
// First database (MySQL)
builder.Model(&UserModel{}, &OrderModel{})

// Second database (SQLite)
sqliteDB, _ := db.ConnectionSQLite("./logs.db")
logGroup := wf.NewModelGroupBuilder().
    Name("log_group").
    DB(sqliteDB).
    Model(&LogModel{}).
    AutoCreateTable(true).
    Build()
builder.ModelGroup(logGroup)
```

### Multi-Database Transactions

Transactions for different model groups are independent:

```go
// Default model group transaction
tx := ctx.GetTransaction()
err := tx.Exec(func(tx *db.DB) error {
    userModel := wf.GetReNewModel[*UserModel](tx, ctx)
    return userModel.Save(user)
})

// Named model group transaction
logTx := ctx.GetTransactionByName("log_group")
err = logTx.Exec(func(tx *db.DB) error {
    logModel := wf.GetReNewModel[*LogModel](tx, ctx)
    return logModel.Save(logEntry)
})
```

> **Note**: Cross-model-group transactions do not support distributed transactions. Handle cross-database consistency at the application level.

## Custom Database Driver

The framework includes MySQL, PostgreSQL, and SQLite drivers. To use other databases (e.g., SQL Server, ClickHouse), register a custom driver with `db.RegisterDB`.

> **Note**: The framework uses GORM as the underlying ORM engine, so custom databases must have a corresponding GORM driver. See [GORM official documentation](https://gorm.io/docs/connecting_to_the_database.html) for supported databases.

### Register Custom Driver

```go
// 1. Define config struct implementing db.IConfig interface
type ClickHouseConfig struct {
    Host     string
    Port     int
    Database string
    User     string
    Password string
}

func (c *ClickHouseConfig) Connection() (*db.DB, error) {
    dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.Database)
    gormDB, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return &db.DB{DB: gormDB}, nil
}

// 2. Register before app starts
func main() {
    db.RegisterDB("clickhouse", &ClickHouseConfig{})

    builder := wf.NewBuilder(config.LoadAutoConfig())
    app := builder.Build()
    app.Start()
}
```

### Configuration

After registration, set `type` to the registered name:

```yaml
web:
  db:
    type: clickhouse
    host: localhost
    port: 9000
    database: mydb
    user: default
    password: ""
```

## Connection Pool

```yaml
web:
  db:
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600  # seconds
```

## Associations (Preload & Joins)

Both `Preload` and `Joins` automatically set the GORM `Model()` clause to ensure associations resolve correctly (v1.0.14 fix for Joins):

```go
// Preload (eager loading)
user, err := userModel.Query().
    Preload("Profile").
    Preload("Roles").
    Where("id = ?", 1).
    One()

// Joins (association name resolved by GORM)
users, err := userModel.Query().
    Joins("Profile").
    Where("status = ?", 1).
    All()

// Raw JOIN
users, err := userModel.Query().
    Joins("JOIN orders ON orders.user_id = t_user.id").
    Where("t_user.status = ?", 1).
    All()
```

> **Not-found handling**: `Query.One()` returns a zero value and `nil` error when no record is found (not `gorm.ErrRecordNotFound`). Check the return value to determine if a record exists. To use `errors.Is(err, gorm.ErrRecordNotFound)`, use raw GORM via `db.GetGorm()`.

## Aggregate Queries

Use `Aggregate()` for SUM, COUNT, AVG, GROUP BY, HAVING, DISTINCT:

```go
// Scalar aggregate
var total float64
err := orderModel.Aggregate().
    Select("SUM(amount)").
    Where("status = ?", 1).
    Aggregate(&total)

// Grouped aggregate
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

// DISTINCT
var cnt int
err := orderModel.Aggregate().
    Select("COUNT(DISTINCT category)").
    Aggregate(&cnt)
```

> `Aggregate(result)` auto-detects the result type: scalar pointers (e.g. `*float64`) get `LIMIT 1`; slice pointers (e.g. `*[]Stat`) return all rows.

## Raw SQL

```go
// Raw SQL query (Where conditions are auto-merged)
users, err := userModel.Query().
    Where("status = ?", 1).
    Exec("SELECT * FROM t_user WHERE status = ?", 1)
```

> **Deprecated**: `ExecPage()` is deprecated. Use `Page()` or `PageForWeb()` for pagination instead.