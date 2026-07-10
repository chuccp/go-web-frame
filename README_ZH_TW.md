# Go Web Frame

**用 Go 寫 CRUD 後端，零 ORM 樣板程式碼。定義 struct，嵌入泛型 Model——型別安全的查詢、分頁、context 傳播都包含在內。**

---

**🌐 Language / 語言**
[English](README.md) • [中文](README_ZH.md) • [繁體中文](README_ZH_TW.md) • [日本語](README_JA.md)

---

## 框架概述

Go Web Frame 是一個整合好的後端工具箱。路由、ORM、快取等組件已預先整合，不需要分別選型再手動組裝。

核心是**一個消除 CRUD 樣板程式碼的泛型 Model 層。** 定義實體 struct，嵌入 `Model[T]`，編譯器從資料庫到 handler 全程檢查資料型別。沒有 `interface{}`，不用程式碼生成。

```go
// 定義一次，到處使用
type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement"`
    Name string
}

type UserModel struct {
    *model.EntryModel[*User, uint]   // ← 所有 CRUD 方法都來自這裡
}

func (m *UserModel) Init(db *db.DB, c *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](db, "t_user")
    return m.CreateTable()
}

// 就這些。現在可以在任何地方用了：
user, _ := userModel.FindByPK(1)          // → *User
users, _ := userModel.FindAll()            // → []*User
users, _ := userModel.Query().Where("age > ?", 18).Order("id desc").All()
total, _ := userModel.Query().Where("status = ?", 1).Count()
users, total, _ := userModel.Page(&web.Page{PageNo: 1, PageSize: 10})
userModel.Update().Where("id = ?", 1).UpdateColumn("name", "新名字")
userModel.DeleteByPK(1)
```

還附帶：路由（Gin）、認證過濾器、WebSocket、SSE、CORS、限流、定時任務、校驗、快取、Let's Encrypt HTTPS、多資料庫（MySQL / PostgreSQL / SQLite / Redis）——一個 YAML 配好全跑起來。

---

## 快速開始

### 30 秒：Hello World

```bash
go get github.com/chuccp/go-web-frame
```

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })
    builder.Build().Run(context.Background())
}
```

```bash
go run main.go
# → http://localhost:19009
```

### 5 分鐘：REST + 資料庫

建立 `application.yml`：

```yaml
web:
  db:
    type: sqlite
    file_path: ./data.db
```

建立 `main.go`：

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/model"
    "github.com/chuccp/go-web-frame/web"
)

// ── 實體 ──
type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement"`
    Name string `gorm:"size:255"`
}

// ── Model（零 CRUD 樣板程式碼）──
type UserModel struct {
    *model.EntryModel[*User, uint]
}

func (m *UserModel) Init(database *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](database, "t_user")
    return m.CreateTable()
}

// ── Controller ──
type UserController struct {
    core.IService
    userModel *UserModel
}

func (c *UserController) Init(ctx *core.Context) error {
    c.userModel = wf.GetModel[*UserModel](ctx)

    ctx.Get("/users", c.List)
    ctx.Get("/users/:id", c.Get)
    ctx.Post("/users", c.Create)
    ctx.Put("/users/:id", c.Update)
    ctx.Delete("/users/:id", c.Delete)
    return nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    return c.userModel.FindAll()
}

func (c *UserController) Get(req *web.Request) (any, error) {
    return c.userModel.FindByPK(req.ParamUint("id"))
}

func (c *UserController) Create(req *web.Request) (any, error) {
    var user User
    if err := req.BindJSON(&user); err != nil {
        return nil, err
    }
    return &user, c.userModel.Save(&user)
}

func (c *UserController) Update(req *web.Request) (any, error) {
    var user User
    if err := req.BindJSON(&user); err != nil {
        return nil, err
    }
    user.Id = req.ParamUint("id")
    return nil, c.userModel.UpdateByPK(&user)
}

func (c *UserController) Delete(req *web.Request) (any, error) {
    return nil, c.userModel.DeleteByPK(req.ParamUint("id"))
}

// ── Main ──
func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Model(&UserModel{})
    builder.Rest(&UserController{})
    builder.Build().Run(context.Background())
}
```

```bash
curl http://localhost:19009/users              # → [{"Id":1,"Name":"alice"}]
curl http://localhost:19009/users/1            # → {"Id":1,"Name":"alice"}
curl -X POST ... -d '{"Name":"bob"}'          # → {"Id":2,"Name":"bob"}
```

資料表自動建立，CRUD 全通。不需要寫 SQL，不需要寫 ORM 樣板程式碼。

---

## 操作資料庫

### Model 分兩級，按需選擇

| 類型 | 提供的方法 | 適用場景 |
|---|---|---|
| `Model[T]` | `Save`、`Query()`、`Update()`、`Delete()`、`CreateTable()`、`WithContext()` | 需要完全控制，用流式建構器 |
| `EntryModel[T, PK]` | Model 全部 + `FindByPK`、`FindAll`、`DeleteByPK`、`UpdateByPK`、`UpdateColumn`、`Page` | 實體有主鍵（最常見） |

`PK` 可以是 `uint`、`int`、`string` 等滿足 `~uint | ~int | ~string` 約束的任意型別。

### 日常查詢

```go
m := userModel.WithContext(req.Ctx())   // context 自動傳播到所有 DB 呼叫

// 取得
user, err := m.FindByPK(1)                          // 按主鍵
users, err := m.FindAll()                            // 全表
user, err := m.Query().Where("email = ?", email).One()
users, err := m.Query().Where("status = ?", 1).Order("id desc").List(100)

// 分頁
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := m.Query().Where("age > ?", 18).Page(page)
pageAble, err := m.Query().Where("age > ?", 18).PageForWeb(page)  // 回傳 PageAble[*User]

// 計數
count, err := m.Query().Where("status = ?", 1).Count()

// 關聯查詢（GORM Preload）
users, err := m.Query().Preload("Orders").Preload("Profile").All()
user, err := m.Query().Where("id = ?", 1).Preload("Orders").One()

// Join
users, err := m.Query().Joins("JOIN orders ON orders.user_id = t_user.id").All()
```

### 日常寫入

```go
// 插入
err := m.Save(&User{Name: "alice"})

// 按主鍵更新
user.Name = "新名字"
err := m.UpdateByPK(user)

// 更新單列
err := m.UpdateColumn(1, "status", 0)

// 條件更新
err := m.Update().Where("status = ?", 0).UpdateForMap(map[string]any{"status": 1})

// 刪除
err := m.DeleteByPK(1)
err := m.Delete().Where("status = ?", -1).Delete()
```

### 對比原生 GORM

```go
// Go Web Frame — 表名綁定一次，ctx 傳播一次，分頁內建
m := userModel.WithContext(req.Ctx())
user, _ := m.FindByPK(1)
users, total, _ := m.Query().Where("age > ?", 18).Page(&web.Page{PageNo: 1, PageSize: 10})

// 原生 GORM — 透過 GetGorm() 取得 *gorm.DB，直接使用 GORM 生態
var user User
db.GetGorm().WithContext(ctx).Table("t_user").Where("id = ?", 1).First(&user)
var users []User
db.GetGorm().WithContext(ctx).Table("t_user").Where("age > ?", 18).Offset(0).Limit(10).Find(&users)
var total int64
db.GetGorm().WithContext(ctx).Table("t_user").Where("age > ?", 18).Count(&total)
```

`EntryModel` 是輔助封裝——減少樣板程式碼，而非限制能力。內建方法不夠用時（複雜 JOIN、子查詢、視窗函式等），呼叫 `db.GetGorm()` 直接使用 [GORM](https://gorm.io/zh_CN/docs/) 生態，兩種方式可自由混用。

原生 GORM 每次都要重複表名、context 和型別。Go Web Frame 在建構時綁定一次，開發時差異巨大——有 20 個 Model 的時候最明顯。

---

## 組織專案

### Builder：一個入口註冊所有組件

```go
builder := wf.NewBuilder(config.LoadAutoConfig())

// 基礎設施（最先初始化）
builder.Filter(&cors.Filter{})        // CORS 標頭
builder.Filter(&AuthFilter{})         // 認證檢查

// 資料層
builder.Model(&UserModel{})
builder.Model(&OrderModel{})
builder.Model(&ProductModel{})

// 業務邏輯
builder.Service(&UserService{})       // 共享業務邏輯
builder.Service(&PaymentService{})

// HTTP 層
builder.Rest(&UserController{})
builder.Rest(&OrderController{})

// 背景任務
builder.Runner(&CleanupTask{})

// 啟動
app := builder.Build()
app.Run(ctx)
```

註冊順序決定初始化順序：Filter → Model → Service → Controller → Runner。依賴透過 `Init()` 自動注入。

### 執行時取得依賴

```go
// 在任何 Init() 或 handler 中，按型別取得：
userModel := wf.GetModel[*UserModel](ctx)
userService := wf.GetService[*UserService](ctx)
authFilter := wf.GetFilter[*AuthFilter](ctx)
cleanupTask := wf.GetRunner[*CleanupTask](ctx)
```

沒有字串 key，不做型別斷言。泛型函式直接處理。

### Service 層：共享業務邏輯

多個 Controller 共用邏輯時，抽出一個 Service：

```go
type UserService struct {
    core.IService
    userModel *UserModel
}

func (s *UserService) Init(ctx *core.Context) error {
    s.userModel = wf.GetModel[*UserModel](ctx)
    return nil
}

func (s *UserService) GetActiveUsers() ([]*User, error) {
    return s.userModel.Query().Where("status = ?", 1).All()
}

// 註冊：
builder.Service(&UserService{})

// 在 Controller 裡用：
userService := wf.GetService[*UserService](ctx)
users, _ := userService.GetActiveUsers()
```

### Model Group：多資料庫

```go
// 預設資料庫
builder.Model(&UserModel{}, &LogModel{})

// 分析系統用獨立資料庫
analyticsGroup := wf.NewModelGroupBuilder().
    Name("analytics").
    DB(analyticsDB).
    Model(&ReportModel{}).
    AutoCreateTable(true).
    Build()
builder.ModelGroup(analyticsGroup)
```

---

## 認證和中間件

### 路由級元資料（WithMeta）

在路由上宣告，Filter 統一處理——不用在每個 handler 裡寫認證邏輯：

```go
func (c *ApiController) Init(ctx *core.Context) error {
    // 公開介面
    ctx.Get("/api/login", login).WithMeta(SkipAuth())

    // 需要登入
    ctx.Get("/api/profile", profile).WithMeta(RequireAuth())

    // 需要登入 + 特定權限
    ctx.Post("/api/admin/users", createUser).
        WithMeta(RequireAuth(), RequirePermission("admin:create_user"))
    return nil
}
```

```go
// 定義元資料工廠
func RequireAuth() web.MetaOption      { return web.WithValue("require_auth", true) }
func SkipAuth() web.MetaOption          { return web.WithValue("skip_auth", true) }
func RequirePermission(p string) web.MetaOption { return web.WithValue("require_permission", p) }
```

```go
// 一個 Filter 處理所有認證邏輯
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if !req.HasMeta(RequireAuth()) || req.HasMeta(SkipAuth()) {
        return fc.Next()
    }

    token := req.Request().Header.Get("Authorization")
    if token == "" {
        return nil, errors.New("未登入")
    }

    // 驗證 token，檢查權限...
    return fc.Next()
}
```

### 全域 Filter

透過 `builder.Filter()` 註冊的 Filter 對所有請求生效：

```go
// 日誌
type LoggingFilter struct { core.IFilter }

func (f *LoggingFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    start := time.Now()
    result, err := fc.Next()
    log.Info("請求", zap.String("path", req.FullPath()), zap.Duration("耗時", time.Since(start)))
    return result, err
}

// CORS（內建，直接註冊即可）
builder.Filter(&cors.Filter{})
```

### RestGroup：按路由分組應用 Filter

```go
apiGroup := core.NewRestGroupBuilder().
    ServerConfig(web.DefaultServerConfig()).
    ContextPath("/api/v1").
    Build()

apiGroup.AddFilter(&AuthFilter{})     // 只影響這個 Group 裡的路由
apiGroup.AddRest(&UserController{})   // 這個 Controller 的所有路由都需要認證

builder.RestGroup(apiGroup)
```

---

## 常用場景

### 檔案上傳

```go
ctx.Post("/upload", func(req *web.Request) (any, error) {
    file, header, err := req.File("file")
    if err != nil {
        return nil, err
    }
    defer file.Close()

    dst := "./uploads/" + header.Filename
    if err := web.SaveUploadedFile(header, dst); err != nil {
        return nil, err
    }
    return map[string]string{"path": dst}, nil
})
```

### WebSocket

```go
ctx.WebSocket("/ws", func(stream *web.WebSocketStream) error {
    defer stream.Close()
    for {
        typ, msg, err := stream.Read(stream.Context())
        if err != nil {
            return err
        }
        stream.Write(stream.Context(), typ, msg)  // echo
    }
})
```

### Server-Sent Events

```go
ctx.SSE("/events", func(stream *web.SSEStream) error {
    defer stream.Close()
    stream.SendRetry(3000)

    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for range ticker.C {
        stream.Send("update", fmt.Sprintf("time: %s", time.Now()))
    }
    return nil
})
```

### 靜態檔案 & SPA

```go
ctx.Static("/assets", "./public")
ctx.Static("/", "./frontend/dist")   // SPA 建構產物
```

```yaml
# 或透過設定——多目錄查找，自動 404 回退：
web:
  server:
    locations:
      - view/dist
      - www
    page404: 404.html
```

### 反向代理

```go
ctx.ReverseProxy("/api/legacy", "http://127.0.0.1:8081")
```

### 背景任務

```go
type CleanupTask struct { core.IRunner }

func (t *CleanupTask) Init(ctx *core.Context) error { return nil }

func (t *CleanupTask) Run() error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        // 清理過期會話...
    }
    return nil
}

builder.Runner(&CleanupTask{})
```

### 參數校驗

```go
type CreateUserInput struct {
    Name  string `validate:"required,min=2,max=50"`
    Email string `validate:"required,email"`
}

func (c *Controller) Create(req *web.Request) (any, error) {
    var input CreateUserInput
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }

    validator := wf.GetService[*validator.Validator](req.Context())
    if err := validator.Validate(input); err != nil {
        return nil, web.NewValidationError().WithError(err)
    }

    // input 已校驗，繼續業務...
}
```

### 自訂 HTTP 回應

```go
// 回傳普通 struct → 自動包裝：{"code":200, "data":{...}, "msg":"ok"}
return &User{Id: 1, Name: "alice"}, nil

// 回傳字串 → 純文字回應
return "ok", nil

// 指定狀態碼
return web.DataCode(http.StatusCreated, &user), nil

// 回傳錯誤
return nil, errors.New("出錯了")

// 回傳業務錯誤碼
return nil, web.NewValidationError().WithDetail("名字不能為空")

// 重新導向
return web.Redirect("/new-url"), nil

// 檔案下載
return &web.File{Path: "/path/to/report.pdf", FileName: "report.pdf"}, nil
```

---

## 設定檔

一個 YAML 覆蓋所有設定。自動從 `./config/`、`~/.<appname>/`、`/etc/<appname>/` 載入。

```yaml
web:
  server:
    port: 8080
    context_path: /api            # 全域路由前綴
    ssl:                          # HTTPS：自動憑證或本地憑證
      enabled: true
      hosts:                      # Let's Encrypt 自動申請憑證
        - example.com
      # certs:                    # 或使用本地憑證檔案（可選）
      #   - host: example.com
      #     cert-file: /path/to/fullchain.pem
      #     key-file: /path/to/privkey.pem

  db:
    type: mysql                   # mysql | postgres | sqlite
    host: localhost
    port: 3306
    user: root
    password: your_password
    database: mydb
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600

  log:
    level: info                   # debug | info | warn | error
    path: ./logs/app.log
    max_size: 500                 # 單檔案最大 MB，觸發滾動
    max_backups: 3
    max_age: 30                   # 保留天數
    compress: true

  redis:
    addr: localhost:6379
    password: ""
    db: 0
```

PostgreSQL 和 SQLite：

```yaml
# PostgreSQL
db:
  type: postgres
  host: localhost
  port: 5432
  user: postgres
  password: ""
  database: mydb
  sslmode: disable

# SQLite
db:
  type: sqlite
  file_path: ./data.db
```

支援 JSON、YAML、TOML 格式。

### 自訂資料庫驅動

註冊自訂資料庫驅動，用於框架未內建的資料庫（如 SQL Server、ClickHouse 等）。框架底層使用 GORM 作為 ORM 引擎，因此自訂資料庫必須有對應的 [GORM 驅動](https://gorm.io/docs/connecting_to_the_database.html)。

```go
// 1. 實作 db.IConfig 介面
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

// 2. 在應用啟動前註冊
func main() {
    db.RegisterDB("clickhouse", &ClickHouseConfig{})
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Build().Start()
}
```

然後在設定檔中指定 `type` 為註冊的類型名稱：

```yaml
web:
  db:
    type: clickhouse
    host: localhost
    port: 9000
    database: mydb
```

---

## 技術棧

預整合、生產驗證過的組件：

| 層次 | 庫 | 作用 |
|---|---|---|
| HTTP | Gin | 路由、中介層鏈、參數綁定 |
| ORM | GORM | SQL 驅動、遷移、join/preload |
| 設定 | Viper | 多格式、多路徑自動載入 |
| 日誌 | Zap | 結構化、分級、滾動 |
| JSON | Sonic | 高效能序列化 |
| 快取 | Otter | 本地記憶體快取 |
| 併發 | Conc | 結構化併發池，管理服務生命週期 |
| Redis | go-redis | 發布訂閱、快取 |
| SQLite | modernc/sqlite | 純 Go，零 CGO |
| 校驗 | go-playground/validator | struct tag 校驗 |
| WebSocket | coder/websocket | 連線升級 + 讀寫 |
| 定時任務 | robfig/cron | 表示式排程器 |

---

## 專案結構

```
├── web_frame.go        # Builder、WebFrame 工廠
├── core/               # 介面（IService, IModel, IRest, IRunner, IFilter）、DI context
├── web/                # Request、Response、路由、Filter、SSE、WebSocket、靜態檔案
├── model/              # Model[T]、EntryModel[T, PK]、Query[T]、Update[T]、Delete[T]
├── db/                 # MySQL、PostgreSQL、SQLite 連線管理
├── redis/              # Redis 客戶端封裝
├── config/             # Viper 自動載入
├── log/                # Zap + lumberjack 滾動
├── component/          # cors、cache、rate-limit、captcha、qrcode、cron、validator
├── util/               # 加密、檔案、字串工具
└── example/            # 可執行的範例
    ├── helloworld/     # 最小應用
    ├── rest/           # REST Controller
    ├── model/          # 泛型 ORM 用法
    ├── filter/         # 認證 + 日誌 Filter
    ├── withmeta/       # 路由元資料
    └── background/     # 背景任務
```

---

## 可選模組

重型依賴拆分為獨立子模組，按需安裝：

<!-- component-badges -->
[![cache](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/cache/*&label=cache&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/cache)
[![captcha](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/captcha/*&label=captcha&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/captcha)
[![schedule](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/schedule/*&label=schedule&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/schedule)
[![qrcode](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/qrcode/*&label=qrcode&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/qrcode)
[![ratelimit](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/ratelimit/*&label=ratelimit&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/ratelimit)
<!-- /component-badges -->

```bash
# 核心框架（不含 captcha/qrcode/cron/otter）
go get github.com/chuccp/go-web-frame

# 按需安裝
go get github.com/chuccp/go-web-frame/component/captcha@v1.0.7
go get github.com/chuccp/go-web-frame/component/schedule@v1.0.7
go get github.com/chuccp/go-web-frame/component/qrcode@v1.0.7
go get github.com/chuccp/go-web-frame/component/cache@v1.0.7
go get github.com/chuccp/go-web-frame/component/ratelimit@v1.0.7
```

### 使用說明

#### Cache — 高效能記憶體快取

```go
import "github.com/chuccp/go-web-frame/component/cache"

// 1. 註冊
builder.Service(&cache.Cache{})

// 2. 在控制器或服務中使用
c := core.GetService[*cache.Cache](ctx)
c.Set("user:1", user)
val, ok := c.Get("user:1")
c.SetNX("lock:task", true, 30*time.Second) // 不存在則設定，帶過期時間
```

#### Captcha — 滑動拼圖驗證碼

```yaml
# application.yml
captcha:
  code_key: "your-32-character-key-here!!"  # 32 個字元，必填
  code_iv: "your-16-char-iv"                # 16 個字元，必填
```

```go
import "github.com/chuccp/go-web-frame/component/captcha"

// 1. 註冊
builder.Service(&captcha.Captcha{})

// 2. 生成驗證碼（返回給前端）
c := core.GetService[*captcha.Captcha](ctx)
data, _ := c.GetCaptchaData()
// → 返回 SlideCaptchaData，包含 TileImage、MasterImage、ThumbCode

// 3. 校驗使用者滑動結果
result, ok := c.ValidateThumb(data.ThumbCode, userXOffset)
if ok {
    valid, _ := c.ValidateCode(result.CaptchaCode)
}
```

#### Schedule — Cron 定時任務

```go
import (
    "github.com/chuccp/go-web-frame/component/schedule"
    "github.com/robfig/cron/v3"
)

// 1. 註冊為 Runner（注意不是 Service！）
sched := schedule.NewSchedule(cron.WithSeconds())
builder.Runner(sched)

// 2. 新增任務 — Build() 前後均可
sched.AddKeyFunc("cleanup", "0 0 * * *", func(ctx *core.Context) {
    // 每天午夜執行
})
sched.AddKeyFunc("report", "*/30 * * * * *", func(ctx *core.Context) {
    // 每 30 秒執行
})
```

#### QRCode — QR 碼產生器

```go
import "github.com/chuccp/go-web-frame/component/qrcode"

// 生成到檔案（條紋樣式）
qrcode.GenerateStripeQRCode("https://example.com", "qr.png")

// 生成到記憶體，自訂樣式
buf := qrcode.CreateBufferWriteCloser()
qrcode.GenerateQrcode("hello", buf, qrcode.WithCircleShape())
pngBytes := buf.Bytes()

// 自訂顏色
s := qrcode.NewStripeQRCode().WithModuleSize(10)
s.GenerateFile("https://example.com", "custom.png")
```

> QRCode 是獨立工具套件，無需註冊，直接 import 呼叫即可。

#### RateLimit — 令牌桶限流器

```yaml
# application.yml（可選，以下為預設值）
rate_limit:
  limit: 10     # 每秒生成令牌數
  burst: 5      # 最大突發容量
```

```go
import "github.com/chuccp/go-web-frame/component/ratelimit"

// 1. 註冊
builder.Service(&ratelimit.RateLimit{})

// 2. 使用 — 通常在 Filter 中
r := core.GetService[*ratelimit.RateLimit](ctx)
if r.Allow(req.ClientIP()) {
    return fc.Next()
}
return nil, errors.New("請求過於頻繁")
```

### 發布子模組

每個子模組透過 tag 前綴獨立版本管理：

```bash
# 1. 發布主模組
git tag v1.0.1
git push origin v1.0.1

# 2. 更新各子模組對主模組的依賴
for mod in captcha schedule qrcode cache ratelimit; do
  cd component/$mod
  go get github.com/chuccp/go-web-frame@v1.0.1
  go mod edit -dropreplace github.com/chuccp/go-web-frame
  go mod tidy
  cd ../..
done

# 3. 為每個子模組打 tag（格式：component/<name>/vX.Y.Z）
git tag component/captcha/v1.0.1
git tag component/schedule/v1.0.1
git tag component/qrcode/v1.0.1
git tag component/cache/v1.0.1
git tag component/ratelimit/v1.0.1

# 4. 推送所有 tag
git push origin --tags
```

Go proxy 會根據 `component/captcha/v1.0.1` tag 前綴自動找到對應目錄的 `go.mod`。

---

## 取得幫助

- **[使用手冊 (中文)](./docs-site/docs-zh/index.md)** — 完整使用文件
- **[使用手冊 (英文)](./docs-site/docs-en/index.md)** — 英文文件
- **[架構設計](./ARCHITECTURE.md)** — 設計決策和內部實現
- **[CLAUDE.md](./CLAUDE.md)** — AI 程式助手使用指南
- **[Go Reference](https://pkg.go.dev/github.com/chuccp/go-web-frame)** — 套件 API 文件

---

## 貢獻

歡迎提交 PR。提交前請執行 `go test ./...`。

## License

MIT — 詳見 [LICENSE](./LICENSE)
