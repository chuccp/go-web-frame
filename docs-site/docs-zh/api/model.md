# 模型 API 参考

本文档介绍 Go Web Frame 的模型层 API。

## Model[T]

`model.Model[T]` 是泛型模型基类。

### 创建模型

```go
type UserModel struct {
    *model.Model[entity.User]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.Model = model.NewModel[entity.User](db, "t_user")
    return m.CreateTable()
}
```

### CRUD 操作

```go
// 保存（插入或更新）
user := &entity.User{Name: "Alice", Email: "alice@example.com"}
err := userModel.Save(user)

// 查询所有
users, err := userModel.Query().All()

// 查询单条记录
// 注意：记录不存在时返回零值和 nil error，不返回 gorm.ErrRecordNotFound
// 应检查返回值是否为零值来判断记录是否存在
user, err := userModel.Query().Where("id = ?", 1).One()

// 条件查询
users, err := userModel.Query().Where("status = ?", 1).All()

// 计数
count, err := userModel.Query().Where("status = ?", 1).Count()

// 更新
err := userModel.Update().Where("id = ?", 1).UpdateForMap(map[string]interface{}{
    "name": "Bob",
})

// 更新单列
err := userModel.Update().Where("id = ?", 1).UpdateColumn("status", 0)

// 删除
err := userModel.Delete().Where("id = ?", 1).Delete()
```

## EntryModel[T, PK]

`model.EntryModel[T, PK]` 为包含主键的实体提供额外方法。`PK` 是主键类型，支持 `uint`、`int`、`string` 等（满足 `PKConstraint` 约束）。

### 创建 EntryModel

```go
type UserModel struct {
    *model.EntryModel[entity.User, uint]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[entity.User, uint](db, "t_user")
    return m.CreateTable()
}
```

### EntryModel 额外方法

```go
// 根据主键查找
user, err := userModel.FindByPK(1)

// 根据主键查找（带预加载）
user, err := userModel.FindByPKWithPreload(1, "Profile", "Role")

// 查找所有记录
users, err := userModel.FindAll()

// 查找所有记录（带预加载）
users, err := userModel.FindAllWithPreload("Profile", "Role")

// 根据主键列表批量查找
users, err := userModel.FindAllByPK(1, 2, 3)

// 根据主键列表批量查找（带预加载）
users, err := userModel.FindAllByPKWithPreload([]uint{1, 2, 3}, "Profile")

// 条件查询单条（找不到返回 nil 而不是错误）
user, err := userModel.FindOne("email = ?", "test@example.com")

// 条件查询单条（带预加载）
user, err := userModel.FindOneWithPreload("status = ?", []interface{}{1}, "Profile")

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
err := userModel.UpdateForMap(1, map[string]interface{}{"name": "Bob"})

// 批量保存
err := userModel.Saves([]entity.User{user1, user2})

// 保存 Map 并返回主键
id, err := userModel.SaveForMapWithPk(map[string]interface{}{"name": "Alice"}, "id")

// 保存 Map 并返回 uint 主键
id, err := userModel.SaveForMapWithUintPk(mapValue, "id")

// 创建记录并返回主键
id, err := userModel.CreateWithPk(user, "id", reflect.Uint)

// ReNew 创建新实例（用于事务）
newModel := userModel.NewEntryModel(tx)
```

## Query[T]

`model.Query[T]` 是查询构建器。

### 基本查询

```go
// 所有记录
users, err := userModel.Query().All()

// 单条记录（记录不存在时返回零值 + nil error）
user, err := userModel.Query().Where("id = ?", 1).One()

// 计数
count, err := userModel.Query().Where("status = ?", 1).Count()
```

### 分页

分页参数自动规范化（`PageNo` 默认 1，`PageSize` 默认 10，通过 `util.DefaultPage` 统一处理）：

```go
// 分页查询（返回列表和总数）
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := userModel.Query().Where("status = ?", 1).Page(page)

// Web 分页（返回 PageAble 结构）
pageAble, err := userModel.Query().Where("status = ?", 1).PageForWeb(page)

// 分页列表（不返回总数）
users, err := userModel.Query().Where("status = ?", 1).ListPage(page)
```

### 排序和限制

```go
// 排序
users, err := userModel.Query().Order("id desc").All()

// 限制数量
users, err := userModel.Query().Order("id desc").List(100)

// 限制数量并返回总数
users, total, err := userModel.Query().Order("id desc").Size(100)
```

### 预加载关联

使用 `Preload` 进行关联预加载。框架自动设置 GORM `Model()` 子句确保关联正确解析：

```go
// 预加载关联（GORM eager loading）
users, err := userModel.Query().Preload("Profile").Preload("Role").All()

user, err := userModel.Query().Where("id = ?", 1).Preload("Profile").One()
```

### JOIN 查询

`Joins` 同样会自动设置 `Model()` 子句（v1.0.14 修复，确保 JOIN 查询中关联正确解析）：

```go
// 关联 JOIN（GORM 自动解析关联名）
users, err := userModel.Query().Joins("Profile").Where("status = ?", 1).All()

// 原生 JOIN
users, err := userModel.Query().
    Joins("JOIN orders ON orders.user_id = t_user.id").
    Where("status = ?", 1).
    All()
```

### 链式更新

```go
// 使用 Set 链式更新多个字段
err := userModel.Update().
    Where("id = ?", 1).
    Set("name", "Alice").
    Set("status", 1).
    Exec()
```

### 原生 SQL

```go
// 原生 SQL 查询（Where 条件自动合并）
users, err := userModel.Query().Where("status = ?", 1).Exec("SELECT * FROM t_user")
```

> **弃用**：`ExecPage()` 已弃用，推荐使用 `Page()` 或 `PageForWeb()` 进行分页查询。

## Aggregate[T]

`model.Aggregate[T]` 是聚合查询构建器，通过 `model.Aggregate()` 创建。

### 方法

| 方法 | 说明 |
|------|------|
| `Select(query, args...)` | 指定聚合表达式（如 `"SUM(amount)"`、`"category, COUNT(*) as cnt"`） |
| `Where(query, args...)` | 添加 WHERE 条件 |
| `Group(name)` | 添加 GROUP BY 子句 |
| `Having(query, args...)` | 添加 HAVING 条件（配合 Group 使用） |
| `Order(query)` | 添加 ORDER BY 子句 |
| `Joins(query)` | 添加 JOIN 子句 |
| `Distinct(args...)` | 添加 DISTINCT（无参数=全行去重，传列名=指定列去重） |
| `Aggregate(result)` | 执行查询并扫描结果（标量自动加 LIMIT 1，切片返回全部行） |
| `WithContext(ctx)` | 设置上下文（返回浅拷贝） |

### 标量聚合

```go
// SUM
var total float64
err := orderModel.Aggregate().Select("SUM(amount)").Where("status = ?", 1).Aggregate(&total)

// COUNT
var count int
err := orderModel.Aggregate().Select("COUNT(*)").Aggregate(&count)

// AVG
var avg float64
err := orderModel.Aggregate().Select("AVG(amount)").Aggregate(&avg)

// 带参数的表达式
var result float64
err := orderModel.Aggregate().Select("SUM(amount) * ? + ?", 2, 100).Aggregate(&result)
```

### 分组聚合

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

### DISTINCT

```go
// 去重查询
var rows []struct{ Category string }
err := orderModel.Aggregate().Distinct("category").Aggregate(&rows)

// COUNT(DISTINCT ...)
var cnt int
err := orderModel.Aggregate().Select("COUNT(DISTINCT category)").Aggregate(&cnt)
```

## Transaction

`model.Transaction` 提供事务支持。

### 基本事务

```go
tx := ctx.GetTransaction()
err := tx.Exec(func(tx *db.DB) error {
    // 在事务中执行操作
    userModel := wf.GetReNewModel[*UserModel](tx, ctx)
    return userModel.Save(user)
})
```

### 命名事务

```go
tx := ctx.GetTransactionByName("user_group")
err := tx.Exec(func(tx *db.DB) error {
    // 在命名事务中执行操作
    userModel := wf.GetReNewModel[*UserModel](tx, ctx)
    return userModel.Save(user)
})
```

## 下一步

- [核心 API](../api/core.md) - 了解核心 API
- [Web API](../api/web.md) - 了解 Web 层 API
