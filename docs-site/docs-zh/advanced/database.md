# 数据库高级用法

本文档介绍 Go Web Frame 中数据库的高级用法。

## 事务处理

### 基本事务

```go
func (s *OrderService) CreateOrder(input *CreateOrderInput) (*Order, error) {
    var order *Order

    tx := ctx.GetTransaction()
    err := tx.Exec(func(tx *db.DB) error {
        // 使用 GetReNewModel 在事务中创建模型
        orderModel := wf.GetReNewModel[*OrderModel](tx, ctx)

        // 步骤 1：创建订单
        order = &Order{UserID: input.UserID, Total: input.Total}
        if err := orderModel.Save(order); err != nil {
            return err
        }

        // 步骤 2：创建订单项
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

### 命名事务

```go
func (s *UserService) UpdateUserWithTransaction(user *User) error {
    tx := ctx.GetTransaction()
    return tx.Exec(func(tx *db.DB) error {
        userModel := wf.GetReNewModel[*UserModel](tx, ctx)
        return userModel.Save(user)
    })
}
```

## 模型组

模型组的核心用途是**支持多数据库**。每个模型组绑定一个独立的数据库连接，同组内的模型共享该连接和事务。

### 默认模型组

使用 `builder.Model()` 注册模型时，框架自动创建默认模型组，使用配置文件中 `web.db` 指定的数据库连接：

```go
func main() {
    builder := wf.NewBuilder(cfg)

    // 注册模型到默认模型组（自动使用 web.db 配置的数据库）
    builder.Model(&UserModel{}, &OrderModel{})

    app := builder.Build()
    app.Start()
}
```

### 多数据库模型组

当应用需要连接多个数据库时，使用 `wf.NewModelGroupBuilder()` 创建独立的模型组：

```go
func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)

    // 默认数据库（MySQL）
    builder.Model(&UserModel{}, &OrderModel{})

    // 第二个数据库（SQLite）
    sqliteDB, _ := db.ConnectionSQLite("./logs.db")
    logGroup := wf.NewModelGroupBuilder().
        Name("log_group").
        DB(sqliteDB).
        Model(&LogModel{}).
        AutoCreateTable(true).
        Build()
    builder.ModelGroup(logGroup)

    // 第三个数据库（另一个 MySQL 实例）
    archiveDB, _ := db.ConnectionMysql("archive.internal", 3306, "reader", "pass", "archive", "utf8mb4")
    archiveGroup := wf.NewModelGroupBuilder().
        Name("archive_group").
        DB(archiveDB).
        Model(&ArchiveModel{}).
        AutoCreateTable(true).
        Build()
    builder.ModelGroup(archiveGroup)

    app := builder.Build()
    app.Start()
}
```

也可通过配置文件创建连接：

```go
// 从配置文件读取 MySQL 配置
var mysqlConfig db.MysqlConfig
cfg.UnmarshalKey("archive_db", &mysqlConfig)
archiveDB, err := mysqlConfig.Connection()

// 从配置文件读取 SQLite 配置
var sqliteConfig db.SQLiteConfig
cfg.UnmarshalKey("log_db", &sqliteConfig)
logDB, err := sqliteConfig.Connection()

// 从配置文件读取 PostgreSQL 配置
var pgConfig db.PostgresConfig
cfg.UnmarshalKey("analytics_db", &pgConfig)
pgDB, err := pgConfig.Connection()
```

### 多数据库事务

不同模型组的事务是独立的，通过名称区分：

```go
// 默认模型组的事务
tx := ctx.GetTransaction()
err := tx.Exec(func(tx *db.DB) error {
    userModel := wf.GetReNewModel[*UserModel](tx, ctx)
    return userModel.Save(user)
})

// 指定模型组的事务
logTx := ctx.GetTransactionByName("log_group")
err := logTx.Exec(func(tx *db.DB) error {
    logModel := wf.GetReNewModel[*LogModel](tx, ctx)
    return logModel.Save(logEntry)
})

// 归档数据库的事务
archiveTx := ctx.GetTransactionByName("archive_group")
err := archiveTx.Exec(func(tx *db.DB) error {
    archiveModel := wf.GetReNewModel[*ArchiveModel](tx, ctx)
    return archiveModel.Save(archiveRecord)
})
```

> **注意**：跨模型组不支持分布式事务。如需跨库一致性，请在应用层自行处理补偿逻辑。

### 动态切换数据库

模型组支持运行时切换数据库连接，所有模型会自动重新初始化：

```go
newDB, err := db.ConnectionMysql("new-host", 3306, "user", "pass", "newdb", "utf8mb4")
modelGroup := ctx.GetModelGroup("log_group")
err = modelGroup.SwitchDB(newDB, ctx)
```

## 查询构建器高级用法

### 复杂查询

```go
// 多条件查询（Where 可多次调用）
users, err := userModel.Query().
    Where("status = ?", 1).
    Where("age > ?", 18).
    Order("create_time desc").
    List(10)
```

### 关联查询

`Preload` 和 `Joins` 都会自动设置 GORM 的 `Model()` 子句，确保关联正确解析（v1.0.14 修复了 Joins 场景）：

```go
// 预加载关联
user, err := userModel.Query().
    Preload("Profile").
    Preload("Roles").
    Where("id = ?", 1).
    One()

// JOIN 查询（关联名由 GORM 自动解析）
users, err := userModel.Query().
    Joins("Profile").
    Where("status = ?", 1).
    All()

// 原生 JOIN
users, err := userModel.Query().
    Joins("JOIN orders ON orders.user_id = t_user.id").
    Where("t_user.status = ?", 1).
    All()
```

> **记录未找到处理**：`Query.One()` 在记录不存在时返回零值和 `nil` error（不返回 `gorm.ErrRecordNotFound`）。应检查返回值是否为零值来判断记录是否存在。如需使用 `errors.Is(err, gorm.ErrRecordNotFound)`，请通过 `db.GetGorm()` 使用原生 GORM。

### 聚合查询

使用 `Aggregate()` 构建器执行 SUM、COUNT、AVG、GROUP BY、HAVING、DISTINCT 等聚合查询：

```go
// 标量聚合
var total float64
err := orderModel.Aggregate().
    Select("SUM(amount)").
    Where("status = ?", 1).
    Aggregate(&total)

// 分组聚合
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

> `Aggregate(result)` 自动判断结果类型：标量值（如 `*float64`）添加 `LIMIT 1`，切片值（如 `*[]Stat`）返回全部行。

### 原生 SQL

```go
// 原生 SQL 查询
var users []User
users, err := userModel.Query().
    Exec("SELECT * FROM t_user WHERE status = ?", 1)
```

## 自定义数据库驱动

框架内置了 MySQL、PostgreSQL 和 SQLite 三种数据库驱动。如果需要使用其他数据库（如 SQL Server、ClickHouse 等），可以通过 `db.RegisterDB` 注册自定义驱动。

> **注意**：框架底层使用 GORM 作为 ORM 引擎，因此自定义数据库必须有对应的 GORM 驱动支持。可以在 [GORM 官方文档](https://gorm.io/docs/connecting_to_the_database.html) 查看支持的数据库列表。

### 注册自定义驱动

```go
// 1. 定义配置结构体，实现 db.IConfig 接口
type ClickHouseConfig struct {
    Host     string
    Port     int
    Database string
    User     string
    Password string
}

func (c *ClickHouseConfig) Connection() (*db.DB, error) {
    // 实现连接逻辑（需要引入对应的 GORM 驱动）
    dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.Database)
    gormDB, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return &db.DB{DB: gormDB}, nil
}

// 2. 在应用启动前注册
func main() {
    db.RegisterDB("clickhouse", &ClickHouseConfig{})

    builder := wf.NewBuilder(config.LoadAutoConfig())
    app := builder.Build()
    app.Start()
}
```

### 配置文件

注册后，在配置文件中指定 `type` 为注册的类型名即可：

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

## 数据库配置

### SQLite

```yaml
web:
  db:
    type: sqlite
    path: ./data.db
```

### MySQL

```yaml
web:
  db:
    type: mysql
    host: localhost
    port: 3306
    database: mydb
    user: root
    password: your_password
    max_open_conns: 100      # 最大打开连接数
    max_idle_conns: 10        # 最大空闲连接数
    conn_max_lifetime: 3600   # 连接最大生命周期（秒）
```

### PostgreSQL

```yaml
web:
  db:
    type: postgres
    host: localhost
    port: 5432
    database: mydb
    user: postgres
    password: your_password
    max_open_conns: 100      # 最大打开连接数
    max_idle_conns: 10        # 最大空闲连接数
    conn_max_lifetime: 3600   # 连接最大生命周期（秒）
```

### 连接池配置

在 YAML 配置文件中设置连接池参数：

```yaml
web:
  db:
    type: mysql
    host: localhost
    port: 3306
    database: mydb
    user: root
    password: your_password
    max_open_conns: 100      # 最大打开连接数
    max_idle_conns: 10        # 最大空闲连接数
    conn_max_lifetime: 3600   # 连接最大生命周期（秒）
```

## 数据库迁移

### 自动迁移

```go
func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.Model = model.NewModel[entity.User](db, "t_user")
    // 自动迁移表结构
    return m.AutoMigrate()
}
```

### 手动迁移

```go
func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.Model = model.NewModel[entity.User](db, "t_user")
    
    // 检查表是否存在
    exists, err := m.IsExist()
    if err != nil {
        return err
    }
    
    // 如果不存在则创建
    if !exists {
        return m.CreateTable()
    }
    
    return nil
}
```

## 下一步

- [组件](../guide/components.md) - 了解缓存组件等
- [最佳实践](../best-practices.md) - 推荐的使用模式
