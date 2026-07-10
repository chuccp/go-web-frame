# 模型

模型层负责数据访问，提供基于 Go 泛型的类型安全 ORM。

## 定义实体

```go
package entity

import "time"

type User struct {
    Id         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name       string    `gorm:"size:255;not null" json:"name"`
    Email      string    `gorm:"size:255;unique" json:"email"`
    Status     int       `gorm:"default:1" json:"status"`
    CreateTime time.Time `json:"createTime"`
    UpdateTime time.Time `json:"updateTime"`
}
```

### 常用字段标签

| 标签 | 说明 |
|---|---|
| `gorm:"primaryKey"` | 主键 |
| `gorm:"autoIncrement"` | 自增 |
| `gorm:"size:255"` | 字段大小 |
| `gorm:"unique"` | 唯一索引 |
| `gorm:"default:1"` | 默认值 |
| `gorm:"index"` | 创建索引 |
| `gorm:"column:xxx"` | 指定数据库列名 |
| `json:"name"` | JSON 序列化名称 |

## 基础模型 Model[T]

使用泛型 `model.Model[T]` 创建数据访问层：

```go
package model

import (
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/model"
    "myapp/entity"
)

type UserModel struct {
    *model.Model[*entity.User]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.Model = model.NewModel[*entity.User](db, "t_user")
    return m.CreateTable()
}
```

### CRUD 操作

```go
userModel := wf.GetModel[*UserModel](ctx)

// 保存（插入或更新）
user := &entity.User{Name: "Alice", Email: "alice@example.com"}
err := userModel.Save(user)

// 查询所有
users, err := userModel.Query().All()

// 查询单条
user, err := userModel.Query().Where("id = ?", 1).One()

// 条件查询
users, err := userModel.Query().Where("status = ?", 1).All()

// 计数
count, err := userModel.Query().Where("status = ?", 1).Count()

// 更新
err := userModel.Update().Where("id = ?", 1).UpdateForMap(map[string]any{
    "name": "Bob",
})

// 更新单列
err := userModel.Update().Where("id = ?", 1).UpdateColumn("status", 0)

// 删除
err := userModel.Delete().Where("id = ?", 1).Delete()
```

## EntryModel[T, PK]

对于包含主键的实体，使用 `EntryModel` 获得更多内置方法。`PK` 支持 `uint`、`int`、`string` 等类型。

```go
type UserModel struct {
    *model.EntryModel[*entity.User, uint]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[*entity.User, uint](db, "t_user")
    return m.CreateTable()
}
```

> **关于 EntryModel 与 GORM 生态**
> 
> `EntryModel` 和 `Model` 是框架提供的辅助封装，旨在减少样板代码。它们底层基于 GORM，你完全可以跳过这些封装，直接使用 GORM 的原生 API：
> 
> ```go
> // 方式一：使用 EntryModel（减少代码量，推荐日常 CRUD）
> user, err := userModel.FindByPK(1)
> 
> // 方式二：直接使用 GORM（更灵活，适合复杂场景）
> var user entity.User
> err := db.DB.First(&user, 1).Error
> ```
> 
> 当 EntryModel 的内置方法无法满足需求时（如复杂 JOIN、子查询、窗口函数等），直接使用 GORM 生态是最佳选择。两者可以混合使用，框架不做限制。
> 
> ### EntryModel 额外方法

```go
// 根据主键查找
user, err := userModel.FindByPK(1)

// 根据主键查找（带预加载）
user, err := userModel.FindByPKWithPreload(1, "Profile", "Role")

// 查找所有
users, err := userModel.FindAll()

// 查找所有（带预加载）
users, err := userModel.FindAllWithPreload("Profile", "Role")

// 根据主键列表批量查找
users, err := userModel.FindAllByPK(1, 2, 3)

// 根据主键列表批量查找（带预加载）
users, err := userModel.FindAllByPKWithPreload([]uint{1, 2, 3}, "Profile")

// 条件查询单条
user, err := userModel.FindOne("email = ?", "test@example.com")

// 根据主键删除
err := userModel.DeleteByPK(1)

// 根据主键更新
err := userModel.UpdateByPK(user)

// 分页查询
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := userModel.Page(page)

// 带条件的分页查询
users, total, err := userModel.QueryPage(page, "status = ?", 1)

// 更新单列
err := userModel.UpdateColumn(1, "status", 0)

// 按主键更新 Map
err := userModel.UpdateForMap(1, map[string]any{"name": "Bob"})

// 批量保存
err := userModel.Saves([]*entity.User{user1, user2})
```

## 查询构建器

### 基本查询

```go
users, err := userModel.Query().All()
user, err := userModel.Query().Where("id = ?", 1).One()
count, err := userModel.Query().Where("status = ?", 1).Count()
```

### 分页

```go
page := &web.Page{PageNo: 1, PageSize: 10}

// 返回列表和总数
users, total, err := userModel.Query().Where("status = ?", 1).Page(page)

// 返回 PageAble 结构
pageAble, err := userModel.Query().Where("status = ?", 1).PageForWeb(page)

// 只返回列表
users, err := userModel.Query().Where("status = ?", 1).ListPage(page)
```

### 排序和限制

```go
users, err := userModel.Query().Order("id desc").All()
users, err := userModel.Query().Order("id desc").List(100)
users, total, err := userModel.Query().Order("id desc").Size(100)
```

### 预加载关联

```go
users, err := userModel.Query().Preload("Profile").Preload("Role").All()
user, err := userModel.Query().Where("id = ?", 1).Preload("Profile").One()
```

### JOIN 查询

```go
users, err := userModel.Query().Joins("Profile").Where("status = ?", 1).All()
```

### 链式更新

```go
err := userModel.Update().
    Where("id = ?", 1).
    Set("name", "Alice").
    Set("status", 1).
    Exec()

// 或使用 UpdateForMap
err := userModel.Update().Where("id = ?", 1).UpdateForMap(map[string]any{
    "name":   "Bob",
    "status": 1,
})
```

### 原生 SQL

```go
users, err := userModel.Query().
    Where("status = ?", 1).
    Exec("SELECT * FROM t_user WHERE status = ?")

users, total, err := userModel.Query().
    Where("status = ?", 1).
    Order("id desc").
    ExecPage(page, "SELECT * FROM t_user WHERE status = ?")
```

## 请求上下文传播

所有模型和构建器都支持 `WithContext(ctx)`，将请求上下文传播到数据库操作：

```go
// 在 Handler 中使用请求上下文
func (c *UserController) List(req *web.Request) (any, error) {
    return c.userModel.WithContext(req.Ctx()).FindAll()
}

// 在查询构建器中使用
users, err := userModel.Query().WithContext(req.Ctx()).Where("age > ?", 18).All()
```

`WithContext` 返回浅拷贝，原模型不变，可并发安全使用。

## 模型组

模型组用于支持多数据库。每个模型组绑定独立的数据库连接，同组模型共享连接和事务。

### 默认模型组

使用 `builder.Model()` 注册模型时，框架会自动创建默认模型组，使用配置文件中 `web.db` 指定的数据库：

```go
func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Model(&UserModel{}, &OrderModel{})
    builder.Build().Run(context.Background())
}
```

### 多数据库模型组

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/db"
)

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)

    // 默认数据库
    builder.Model(&UserModel{}, &OrderModel{})

    // 第二个数据库（SQLite）
    sqliteDB, err := db.ConnectionSQLite("./logs.db")
    if err != nil {
        panic(err)
    }
    logGroup := wf.NewModelGroupBuilder().
        Name("log_group").
        DB(sqliteDB).
        Model(&LogModel{}).
        AutoCreateTable(true).
        Build()
    builder.ModelGroup(logGroup)

    builder.Build().Run(context.Background())
}
```

### 事务

```go
// 默认模型组事务
tx := ctx.GetTransaction()
err := tx.Exec(func(tx *db.DB) error {
    userModel := wf.GetReNewModel[*UserModel](tx, ctx)
    return userModel.Save(user)
})

// 指定模型组事务
logTx := ctx.GetTransactionByName("log_group")
err := logTx.Exec(func(tx *db.DB) error {
    logModel := wf.GetReNewModel[*LogModel](tx, ctx)
    return logModel.Save(logEntry)
})
```

> 跨模型组事务不支持分布式事务，跨库一致性需在应用层处理。

## 完整示例

```go
package main

import (
    "context"
    "time"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/model"
)

type User struct {
    Id         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name       string    `gorm:"size:255;not null" json:"name"`
    Email      string    `gorm:"size:255;unique" json:"email"`
    Status     int       `gorm:"default:1" json:"status"`
    CreateTime time.Time `json:"createTime"`
    UpdateTime time.Time `json:"updateTime"`
}

type UserModel struct {
    *model.EntryModel[*User, uint]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](db, "t_user")
    return m.CreateTable()
}

func (m *UserModel) FindActiveUsers() ([]*User, error) {
    return m.Query().Where("status = ?", 1).Order("create_time desc").All()
}

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Model(&UserModel{})
    builder.Build().Run(context.Background())
}
```

## 下一步

- [服务](service.md) - 业务逻辑层
- [数据库高级用法](../advanced/database.md) - 事务、迁移等
- [后台任务](runner.md) - Runner 和定时任务
