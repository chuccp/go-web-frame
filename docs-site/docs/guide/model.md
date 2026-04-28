# 模型

模型层负责数据访问操作，提供类型安全的 ORM 功能。

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

### 字段标签

| 标签 | 说明 |
|------|------|
| `gorm:"primaryKey"` | 主键 |
| `gorm:"autoIncrement"` | 自增 |
| `gorm:"size:255"` | 字段大小 |
| `gorm:"unique"` | 唯一索引 |
| `gorm:"default:1"` | 默认值 |
| `gorm:"index"` | 创建索引 |
| `json:"name"` | JSON 序列化名称 |

## 基本模型

使用泛型 `model.Model[T]` 创建数据访问：

```go
package model

import (
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/model"
    "myapp/entity"
)

type UserModel struct {
    *model.Model[entity.User]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    // 创建表
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

## EntryModel

对于包含 `Id`、`CreateTime`、`UpdateTime` 字段的实体，使用 `EntryModel` 获得更多内置方法：

### 实体必须实现 IEntry 接口

```go
package entity

import "time"

type User struct {
    Id         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name       string    `gorm:"size:255" json:"name"`
    CreateTime time.Time `json:"createTime"`
    UpdateTime time.Time `json:"updateTime"`
}

func (u *User) SetCreateTime(t time.Time) { u.CreateTime = t }
func (u *User) SetUpdateTime(t time.Time) { u.UpdateTime = t }
func (u *User) GetId() uint                { return u.Id }
func (u *User) SetId(id uint)              { u.Id = id }
```

### 创建 EntryModel

```go
package model

import (
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/model"
    "github.com/chuccp/go-web-frame/core"
    "myapp/entity"
)

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

// 根据主键查找（带预加载）
user, err := userModel.FindByIdWithPreload(1, "Profile", "Role")

// 查找所有记录
users, err := userModel.FindAll()

// 查找所有记录（带预加载）
users, err := userModel.FindAllWithPreload("Profile", "Role")

// 根据 ID 列表批量查找
users, err := userModel.FindAllByIds(1, 2, 3)

// 根据 ID 列表批量查找（带预加载）
users, err := userModel.FindAllByIdsWithPreload([]uint{1, 2, 3}, "Profile")

// 条件查询单条（找不到返回 nil 而不是错误）
user, err := userModel.FindOne("email = ?", "test@example.com")

// 条件查询单条（带预加载）
user, err := userModel.FindOneWithPreload("status = ?", []interface{}{1}, "Profile")

// 根据 ID 删除
err := userModel.DeleteById(1)

// 根据 ID 更新
err := userModel.UpdateById(user)

// 分页查询
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := userModel.Page(page)

// 带条件的分页查询
users, total, err := userModel.QueryPage(page, "status = ?", 1)

// 更新单列
err := userModel.UpdateColumn(1, "status", 0)

// 按 ID 更新 Map
err := userModel.UpdateForMap(1, map[string]interface{}{"name": "Bob"})

// 批量保存
err := userModel.Saves([]entity.User{user1, user2})

// 保存 Map 并返回主键
id, err := userModel.SaveForMapWithPk(map[string]interface{}{"name": "Alice"}, "id")

// 保存 Map 并返回 uint 主键
id, err := userModel.SaveForMapWithUintPk(mapValue, "id")

// 创建记录并返回主键
id, err := userModel.CreateWithPk(user, "id", reflect.Uint)
```

## 查询构建器

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

```go
// 预加载关联（GORM eager loading）
users, err := userModel.Query().Preload("Profile").Preload("Role").All()

user, err := userModel.Query().Where("id = ?", 1).Preload("Profile").One()
```

### JOIN 查询

```go
// JOIN 查询
users, err := userModel.Query().Joins("Profile").Where("status = ?", 1).All()
```

### 原生 SQL

```go
// 原生 SQL 查询（支持 Where 条件自动合并）
users, err := userModel.Query().
    Where("status = ?", 1).
    Exec("SELECT * FROM t_user WHERE status = ?")

// 原生 SQL 分页查询
users, total, err := userModel.Query().
    Where("status = ?", 1).
    Order("id desc").
    ExecPage(page, "SELECT * FROM t_user WHERE status = ?")
```

## 模型组

将模型分组以共享数据库连接和事务：

```go
func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")

    builder := wf.NewBuilder(cfg)

    // 注册模型到默认模型组
    builder.Model(&UserModel{}, &OrderModel{})

    app := builder.Build()
    app.Start()
}
```

## 完整示例

```go
package main

import (
    "time"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/model"
    "github.com/chuccp/go-web-frame/core"
    "myapp/entity"
)

// 实体
type User struct {
    Id         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name       string    `gorm:"size:255;not null" json:"name"`
    Email      string    `gorm:"size:255;unique" json:"email"`
    Status     int       `gorm:"default:1" json:"status"`
    CreateTime time.Time `json:"createTime"`
    UpdateTime time.Time `json:"updateTime"`
}

func (u *User) SetCreateTime(t time.Time) { u.CreateTime = t }
func (u *User) SetUpdateTime(t time.Time) { u.UpdateTime = t }
func (u *User) GetId() uint                { return u.Id }
func (u *User) SetId(id uint)              { u.Id = id }

// 模型
type UserModel struct {
    *model.EntryModel[entity.User]
}

func (m *UserModel) Init(db *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[entity.User](db, "t_user")
    return m.CreateTable()
}

// 自定义查询
func (m *UserModel) FindActiveUsers() ([]entity.User, error) {
    return m.Query().Where("status = ?", 1).Order("create_time desc").All()
}

func (m *UserModel) FindByEmail(email string) (*entity.User, error) {
    return m.Query().Where("email = ?", email).One()
}

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Model(&UserModel{})
    app := builder.Build()
    app.Start()
}
```

## 下一步

- [服务](service.md) - 业务逻辑层
- [数据库高级用法](../advanced/database.md) - 事务、迁移等
- [后台任务](runner.md) - Runner 和定时任务
