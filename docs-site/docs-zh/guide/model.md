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
    gwfmodel "github.com/chuccp/go-web-frame/model"
    "myapp/entity"
)

type UserModel struct {
    *gwfmodel.Model[*entity.User]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.Model = gwfmodel.NewModel[*entity.User](db, "t_user")
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
// 注意：记录不存在时 One() 返回零值和 nil error（不是 gorm.ErrRecordNotFound）
// 应检查返回值判断记录是否存在：
// if user == nil { /* 记录不存在 */ }

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

!!! info "关于 EntryModel 与 GORM 生态"
    `EntryModel` 和 `Model` 是框架提供的辅助封装，旨在减少样板代码。它们底层基于 [GORM](https://gorm.io/zh_CN/docs/)，你完全可以跳过这些封装，直接使用 GORM 的原生 API：

    ```go
    // 方式一：使用 EntryModel（减少代码量，推荐日常 CRUD）
    user, err := userModel.FindByPK(1)

    // 方式二：直接使用 GORM（更灵活，适合复杂场景）
    // 通过 GetGorm() 获取 *gorm.DB，直接操作 GORM 生态
    var user entity.User
    err := db.GetGorm().First(&user, 1).Error
    ```

    当 EntryModel 的内置方法无法满足需求时（如复杂 JOIN、子查询、窗口函数等），直接使用 GORM 生态是最佳选择。两者可以混合使用，框架不做限制。

## EntryModel 额外方法

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

分页参数会自动规范化：`PageNo` 默认 1，`PageSize` 默认 10（通过 `util.DefaultPage` 统一处理，无需手动校验）。

```go
page := &web.Page{PageNo: 1, PageSize: 10}

// 返回列表和总数
users, total, err := userModel.Query().Where("status = ?", 1).Page(page)

// 返回 PageAble 结构
pageAble, err := userModel.Query().Where("status = ?", 1).PageForWeb(page)

// 只返回列表（不分页查询总数）
users, err := userModel.Query().Where("status = ?", 1).ListPage(page)
```

### 排序和限制

```go
users, err := userModel.Query().Order("id desc").All()
users, err := userModel.Query().Order("id desc").List(100)
users, total, err := userModel.Query().Order("id desc").Size(100)
```

### 预加载关联

使用 `Preload` 加载关联数据。框架会自动设置 GORM 的 `Model()` 子句以确保关联正确解析：

```go
users, err := userModel.Query().Preload("Profile").Preload("Role").All()
user, err := userModel.Query().Where("id = ?", 1).Preload("Profile").One()
```

### JOIN 查询

使用 `Joins` 进行 JOIN 查询。与 `Preload` 一样，框架会自动设置 `Model()` 子句（v1.0.14 修复）：

```go
// 关联 JOIN（GORM 自动解析关联名）
users, err := userModel.Query().Joins("Profile").Where("status = ?", 1).All()

// 原生 JOIN
users, err := userModel.Query().
    Joins("JOIN orders ON orders.user_id = t_user.id").
    Where("status = ?", 1).
    All()
```

> **注意**：`Preload` 和 `Joins` 都会触发 `Table.Model(entry)` 调用，确保 GORM 能正确解析关联。无需手动调用。

### 聚合查询

使用 `Aggregate()` 构建聚合查询，支持 `SUM`、`COUNT`、`AVG`、`MAX`、`MIN` 等 SQL 聚合函数：

```go
// 标量聚合 — 结果扫描到单个值
var total float64
err := orderModel.Aggregate().
    Select("SUM(amount)").
    Where("status = ?", 1).
    Aggregate(&total)

var count int
err := orderModel.Aggregate().
    Select("COUNT(*)").
    Aggregate(&count)

var avg float64
err := orderModel.Aggregate().
    Select("AVG(amount)").
    Aggregate(&avg)
```

#### GROUP BY 分组聚合

分组聚合将结果扫描到自定义结构体切片：

```go
type CategoryStat struct {
    Category string  `json:"category"`
    Total    float64 `json:"total"`
    Count    int     `json:"count"`
    AvgAmt   float64 `json:"avgAmt"`
}

var stats []CategoryStat
err := orderModel.Aggregate().
    Select("category, SUM(amount) as total, COUNT(*) as count, AVG(amount) as avg_amt").
    Group("category").
    Aggregate(&stats)
```

#### HAVING 过滤

```go
var stats []CategoryStat
err := orderModel.Aggregate().
    Select("category, SUM(amount) as total").
    Group("category").
    Having("SUM(amount) > ?", 200).
    Aggregate(&stats)
```

#### DISTINCT 去重

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

#### 带参数的表达式

```go
var result float64
err := orderModel.Aggregate().
    Select("SUM(amount) * ? + ?", 2, 100).
    Aggregate(&result)
```

> **提示**：`Aggregate(result)` 自动判断结果类型——标量值（如 `*float64`）添加 `LIMIT 1`，切片值（如 `*[]Stat`）返回所有匹配行。

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
```

> **弃用提示**：`ExecPage()` 已标记为弃用，推荐使用 `Page()` 或 `PageForWeb()` 进行分页查询。

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
