# 数据库高级用法

本文档介绍 Go Web Frame 中数据库的高级用法。

## 事务处理

### 基本事务

```go
func (s *OrderService) CreateOrder(input *CreateOrderInput) (*Order, error) {
    var order *Order
    
    tx := s.GetTransaction()
    err := tx.Exec(func(db *gorm.DB) error {
        // 步骤 1：创建订单
        order = &Order{UserID: input.UserID, Total: input.Total}
        if err := db.Create(order).Error; err != nil {
            return err
        }
        
        // 步骤 2：创建订单项
        for _, item := range input.Items {
            orderItem := &OrderItem{OrderID: order.Id, ProductID: item.ProductID}
            if err := db.Create(orderItem).Error; err != nil {
                return err
            }
        }
        
        // 步骤 3：更新库存
        for _, item := range input.Items {
            if err := db.Model(&Product{}).
                Where("id = ?", item.ProductID).
                Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
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
    tx := s.GetTransaction()
    return tx.Exec(func(db *gorm.DB) error {
        return db.Save(user).Error
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
// 多条件查询
users, err := userModel.Query().
    Where("status = ?", 1).
    Where("age > ?", 18).
    Order("create_time desc").
    Limit(10).
    Find()
```

### 关联查询

```go
// 预加载关联
user, err := userModel.Query().
    Preload("Profile").
    Preload("Roles").
    Where("id = ?", 1).
    First()
```

### 原生 SQL

```go
// 原生 SQL 查询
var users []User
err := userModel.Query().
    Raw("SELECT * FROM t_user WHERE status = ?", 1).
    Scan(&users).Error
```

## 数据库配置

### SQLite

```ini
[core]
dbType = sqlite

[sqlite]
filename = data.db
```

### MySQL

```ini
[core]
dbType = mysql

[mysql]
host     = localhost
port     = 3306
dbname   = mydb
charset  = utf8
username = root
password = password
```

### 连接池配置

```go
// 在代码中配置连接池
db.DB().SetMaxOpenConns(100)
db.DB().SetMaxIdleConns(10)
db.DB().SetConnMaxLifetime(time.Hour)
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
