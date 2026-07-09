# Database

Advanced database usage.

## Transactions

```go
err := db.Transaction(func(tx *db.DB) error {
    userModel := model.NewModel[*User](tx, "t_user")
    orderModel := model.NewModel[*Order](tx, "t_order")
    
    user := &User{Name: "alice"}
    if err := userModel.Save(user); err != nil {
        return err
    }
    
    order := &Order{UserId: user.Id}
    return orderModel.Save(order)
})
```

## Model Groups

Share database connection across models:

```go
group := builder.NewModelGroup(db, "user_group")
group.AddModel(&UserModel{}, &ProfileModel{})
```

## Multiple Databases

```go
db1, err := db.NewDB(config1)
db2, err := db.NewDB(config2)

builder.ModelGroup("users").WithDB(db1).Register(&UserModel{})
builder.ModelGroup("orders").WithDB(db2).Register(&OrderModel{})
```

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

    app := wf.NewWithAutoConfig()
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

## Raw SQL

```go
result, err := db.Raw("SELECT * FROM users WHERE id = ?", 1)
```