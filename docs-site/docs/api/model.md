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

## EntryModel[T]

`model.EntryModel[T]` 为包含 `Id`、`CreateTime`、`UpdateTime` 字段的实体提供额外方法。

### 创建 EntryModel

```go
type UserModel struct {
    *model.EntryModel[entity.User]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[entity.User](db, "t_user")
    return m.CreateTable()
}
```

### EntryModel 额外方法

```go
// 根据主键查找
user, err := userModel.FindById(1)

// 查找所有记录
users, err := userModel.FindAll()

// 根据 ID 删除
err := userModel.DeleteById(1)

// 根据 ID 更新
err := userModel.UpdateById(user)

// 分页查询
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := userModel.Page(page)

// 更新单列
err := userModel.UpdateColumn(1, "status", 0)
```

## Query[T]

`model.Query[T]` 是查询构建器。

### 基本查询

```go
// 所有记录
users, err := userModel.Query().All()

// 单条记录
user, err := userModel.Query().Where("id = ?", 1).One()

// 计数
count, err := userModel.Query().Where("status = ?", 1).Count()
```

### 分页

```go
// 分页查询
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := userModel.Query().Where("status = ?", 1).Page(page)

// Web 分页（返回 PageAble 结构）
pageAble, err := userModel.Query().Where("status = ?", 1).PageForWeb(page)
```

### 排序和限制

```go
// 排序
users, err := userModel.Query().Order("id desc").All()

// 限制数量
users, err := userModel.Query().Order("id desc").List(100)
```

### 预加载关联

```go
// 预加载关联（GORM eager loading）
users, err := userModel.Query().Preload("Profile").Preload("Role").All()

user, err := userModel.Query().Where("id = ?", 1).Preload("Profile").One()
```

## Transaction

`model.Transaction` 提供事务支持。

### 基本事务

```go
tx := ctx.GetTransaction()
err := tx.Exec(func(db *gorm.DB) error {
    // 在事务中执行操作
    if err := db.Create(user).Error; err != nil {
        return err
    }
    return nil
})
```

### 命名事务

```go
tx := ctx.GetTransactionByName("user_group")
err := tx.Exec(func(db *gorm.DB) error {
    // 在命名事务中执行操作
    return nil
})
```

## 下一步

- [核心 API](../api/core.md) - 了解核心 API
- [Web API](../api/web.md) - 了解 Web 层 API
