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

### 创建模型组

```go
func main() {
    builder := wf.NewBuilder(cfg)

    // 注册模型到默认模型组
    builder.Model(&UserModel{}, &ProfileModel{})

    app := builder.Build()
    app.Start()
}
```

### 默认模型组

模型通过 `builder.Model()` 注册后会自动加入默认模型组，共享同一个数据库连接和事务：

```go
func main() {
    builder := wf.NewBuilder(cfg)

    // 添加模型到默认模型组
    builder.Model(&UserModel{}, &OrderModel{})

    app := builder.Build()
    app.Start()
}
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

```go
// 预加载关联
user, err := userModel.Query().
    Preload("Profile").
    Preload("Roles").
    Where("id = ?", 1).
    One()
```

### 原生 SQL

```go
// 原生 SQL 查询
var users []User
users, err := userModel.Query().
    Exec("SELECT * FROM t_user WHERE status = ?", 1)
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

- [缓存](../advanced/cache.md) - 了解缓存用法
- [最佳实践](../best-practices.md) - 推荐的使用模式
