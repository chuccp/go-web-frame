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

// ReNew 创建新实例（用于事务）
newModel := userModel.NewEntryModel(tx)

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

### 链式更新

```go
// 使用 Set 链式更新多个字段
err := userModel.Update().
    Where("id = ?", 1).
    Set("name", "Alice").
    Set("status", 1).
    Exec()

// 也可以使用 UpdateForMap 批量更新
err := userModel.Update().Where("id = ?", 1).UpdateForMap(map[string]interface{}{
    "name":   "Bob",
    "status": 1,
})
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

模型组（ModelGroup）的核心用途是**支持多数据库**。每个模型组绑定一个独立的数据库连接，同组内的模型共享该连接和事务。通过模型组，可以在一个应用中同时使用多个不同的数据库（如 MySQL + SQLite、多个 MySQL 实例等）。

### 默认模型组

使用 `builder.Model()` 注册模型时，框架会自动创建默认模型组，并使用配置文件中 `web.db` 指定的数据库连接：

```go
func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")

    builder := wf.NewBuilder(cfg)

    // 注册模型到默认模型组（自动使用 web.db 配置的数据库）
    builder.Model(&UserModel{}, &OrderModel{})

    app := builder.Build()
    app.Start()
}
```

默认模型组的名称为 `ModelDefaultName`，事务通过 `ctx.GetTransaction()` 获取。

### 多数据库模型组

当应用需要连接多个数据库时，使用 `wf.NewModelGroupBuilder()` 创建独立的模型组，每个组绑定不同的数据库连接：

```go
package main

import (
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/core"
)

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")

    builder := wf.NewBuilder(cfg)

    // === 默认数据库（MySQL） ===
    // UserModel 和 OrderModel 使用配置文件中的 web.db 数据库
    builder.Model(&UserModel{}, &OrderModel{})

    // === 第二个数据库（SQLite） ===
    // 手动创建 SQLite 数据库连接
    sqliteDB, err := db.ConnectionSQLite("./logs.db")
    if err != nil {
        panic(err)
    }

    // 创建独立的模型组，绑定 SQLite 连接
    logGroup := wf.NewModelGroupBuilder().
        Name("log_group").          // 模型组名称（用于获取事务）
        DB(sqliteDB).               // 绑定数据库连接
        Model(&LogModel{}).         // 添加模型
        AutoCreateTable(true).      // 自动建表
        Build()

    builder.ModelGroup(logGroup)

    // === 第三个数据库（另一个 MySQL 实例） ===
    archiveDB, err := db.ConnectionMysql("archive.internal", 3306, "reader", "reader_pass", "archive", "utf8mb4")
    if err != nil {
        panic(err)
    }

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

也可以通过配置文件创建数据库连接：

```go
// 从配置文件读取 MySQL 配置并创建连接
var mysqlConfig db.MysqlConfig
cfg.UnmarshalKey("archive_db", &mysqlConfig)
archiveDB, err := mysqlConfig.Connection()

// 从配置文件读取 SQLite 配置并创建连接
var sqliteConfig db.SQLiteConfig
cfg.UnmarshalKey("log_db", &sqliteConfig)
logDB, err := sqliteConfig.Connection()

// 从配置文件读取 PostgreSQL 配置并创建连接
var pgConfig db.PostgresConfig
cfg.UnmarshalKey("analytics_db", &pgConfig)
pgDB, err := pgConfig.Connection()
```

对应配置文件（`application.yml`）：

```yaml
# 默认数据库（MySQL） — 由 builder.Model() 自动使用
web:
  db:
    type: mysql
    host: localhost
    port: 3306
    database: mydb
    user: root
    password: your_password

# 日志数据库（SQLite） — 手动创建连接
log_db:
  file_path: ./logs.db

# 归档数据库（另一个 MySQL 实例） — 手动创建连接
archive_db:
  host: archive.internal
  port: 3306
  database: archive
  user: reader
  password: reader_pass
```

### ModelGroupBuilder API

| 方法 | 说明 |
|------|------|
| `Name(name string)` | 设置模型组名称（用于通过名称获取事务） |
| `DB(db *db.DB)` | 绑定数据库连接 |
| `Model(model ...IModel)` | 添加模型到组 |
| `AutoCreateTable(auto bool)` | 是否自动创建表 |
| `Build()` | 构建模型组 |

### 多数据库事务

不同模型组使用不同的数据库连接，因此事务也是独立的：

```go
// 默认模型组的事务（MySQL）
tx := ctx.GetTransaction()
err := tx.Exec(func(tx *db.DB) error {
    userModel := wf.GetReNewModel[*UserModel](tx, ctx)
    return userModel.Save(user)
})

// 按名称获取指定模型组的事务（SQLite）
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

> **注意**：跨模型组的事务不支持分布式事务。如果业务需要跨库一致性，需要在应用层自行处理补偿逻辑。

### 动态切换数据库

模型组支持在运行时切换数据库连接，所有模型会自动重新初始化：

```go
// 切换模型组的数据库连接
newDB, err := db.CreateDBWithConfig(cfg, "web.new_db")
if err != nil {
    panic(err)
}

modelGroup := ctx.GetModelGroup("log_group")
err = modelGroup.SwitchDB(newDB, ctx)
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
