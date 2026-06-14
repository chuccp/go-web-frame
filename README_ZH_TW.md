# Go Web Frame

一個現代化 Go Web 框架，提供結構化的方式來構建企業級 Web 應用。

---

**🌐 Language / 語言**
[English](README.md) • [中文](README_ZH.md) • [繁體中文](README_ZH_TW.md) • [日本語](README_JA.md)

---

**📖 [Documentation →](./docs-site/docs-en/index.md)**

---

## 專案概述

Go Web Frame 組合了 Go 生態中的最佳開源組件：Gin（HTTP）、GORM（ORM）、Viper（配置）、Zap（日誌）、Sonic（JSON）、Otter（快取）等。提供宣告式路由元資料、類型安全泛型 ORM、依賴注入，開箱即用。

**核心特性：**
- **WithMeta**：為每個路由宣告元資料（認證、權限、限流標記），在 Filter 中統一處理
- **Builder 模式**：顯式註冊、可控的初始化順序，無隱式掃描
- **泛型 ORM**：零樣板程式碼 CRUD，基於 `Model[T]`，無需程式碼生成
- **透明 Context**：所有依賴透過 `GetService[T]` 注入，除錯友好

## 🧩 技術棧

本框架精選並整合了以下優秀的開源組件，經過深度整合和最佳實踐配置：

### 核心框架
| 組件 | 說明 |
|------|------|
| [Gin](https://github.com/gin-gonic/gin) | 高效能 HTTP Web 框架，API 效能優異 |
| [GORM](https://gorm.io/) | 強大的 ORM 函式庫，支援多種資料庫 |
| [Viper](https://github.com/spf13/viper) | 完整的設定解決方案，支援多格式 |
| [Zap](https://go.uber.org/zap) | Uber 出品的高效能結構化日誌函式庫 |

### 資料儲存
| 組件 | 說明 |
|------|------|
| [go-redis](https://github.com/redis/go-redis) | Redis 官方推薦客戶端 |
| [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | 純 Go 實現的 SQLite，無 CGO 依賴 |
| [gorm-driver/mysql](https://gorm.io/docs/connecting_to_the_database.html) | MySQL 資料庫驅動 |
| [gorm-driver/postgres](https://gorm.io/docs/connecting_to_the_database.html) | PostgreSQL 資料庫驅動 |

### 快取與效能
| 組件 | 說明 |
|------|------|
| [Otter](https://github.com/maypok86/otter) | 高效能 Go 本地快取函式庫 |
| [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) | 令牌桶限流器 |

### 實用工具
| 組件 | 說明 |
|------|------|
| [Cron](https://github.com/robfig/cron) | 定時任務調度函式庫 |
| [go-qrcode](https://github.com/yeqown/go-qrcode) | 二維碼生成 |
| [go-captcha](https://github.com/wenlng/go-captcha) | 行為驗證碼生成 |
| [validator](https://github.com/go-playground/validator) | 結構體欄位驗證 |
| [UUID](https://github.com/google/uuid) | UUID 生成 |
| [Lumberjack](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2) | 日誌輪轉 |
| [Conc](https://github.com/sourcegraph/conc) | 更好的並發原語 |
| [Emperror](https://emperror.dev/errors) | 生產級錯誤處理 |

### 為什麼選擇這些組件？

- **生產驗證**：所有組件均在大型生產環境中得到廣泛驗證
- **高效能**：Gin、Zap、Otter 等都是各自領域效能最優的選擇
- **最佳實踐**：經過精心整合，開箱即用，無需繁瑣配置
- **生態成熟**：活躍的社群支持，持續迭代更新

## 🤖 為什麼選擇 Go Web Frame？

### 與其他框架對比

| 特性 | Go Web Frame | Gin | Beego | Echo |
|------|-------------|-----|-------|------|
| 開箱即用 | ✅ 完整方案 | ❌ 需自行整合 | ✅ 完整方案 | ⚠️ 部分整合 |
| 泛型 ORM | ✅ 零樣板 | ❌ 需自行選擇 | ❌ 無泛型 | ❌ 需自行選擇 |
| 依賴注入 | ✅ 內建 DI | ❌ 需自行實現 | ⚠️ 簡單支援 | ❌ 需自行實現 |
| 學習曲線 | 🟢 中等 | 🟢 簡單 | 🟡 較陡 | 🟢 簡單 |
| 功能完整度 | 🟢 高 | 🟡 中等 | 🟢 高 | 🟡 中等 |
| 效能 | 🟢 優秀 (42k QPS) | 🟢 最優 (45k QPS) | 🟡 一般 | 🟢 優秀 |

### 何時應該選擇 Go Web Frame？

**強烈推薦的場景：**

- 🚀 **快速原型開發**：需要短時間內完成專案原型，框架已整合所有必要組件
- 🏢 **企業級應用**：需要清晰架構、依賴注入、統一錯誤處理等企業級特性
- 📊 **管理後台系統**：內建 CRUD 操作、分頁、驗證等常用功能
- 🔌 **RESTful API 服務**：簡化的控制器實現，自動路由註冊
- ⚙️ **微服務開發**：輕量級但功能完整
- 🛠️ **全端 Go 專案**：從前端到後端到資料庫，一站式解決方案

**特別適合：**

- Go 初學者：想學習最佳實踐，避免自己摸索
- 獨立開發者：需要高效完成專案，減少技術選型時間
- 小型團隊：統一技術棧，降低協作成本
- AI 輔助開發：清晰的架構讓 AI 更容易理解和生成程式碼

### 選擇建議

如果你需要：
- 一個功能完整的 Go Web 框架，支援宣告式路由元資料
- 預整合的生產級組件
- 無需程式碼生成的類型安全泛型 ORM
- 清晰的架構，顯式的初始化流程

Go Web Frame 可能適合你。

## 🌟 特性

- 🏷️ **路由元資料 (WithMeta)**：為每個路由宣告元資料（認證、權限、限流），在 Filter 中統一處理 - 無需在每個 Handler裡重複檢查
- 🧱 **Builder 模式**：顯式註冊、可控的初始化順序，無隱式掃描或反射
- ⚡ **類型安全泛型 ORM**：基於 `Model[T]` 的零樣板 CRUD，無需程式碼生成
- 🧩 **依賴注入**：內建 DI 容器，透過 Context 取得 - `GetService[T]`、`GetModel[T]`
- 🎯 **類 MVC 架構**：清晰的關注點分離，包含服務、控制器和模型
- 📦 **資料庫整合**：SQLite、MySQL、PostgreSQL、Redis 支援，可配置連線池參數
- 🛠️ **組件系統**：可重用組件，包括快取、限流、驗證碼、二維碼、定時任務、輸入驗證
- 🌐 **RESTful 支援**：簡化的 REST 控制器實現
- ⚙️ **自動配置**：從 JSON、YAML 或 TOML 檔案自動載入配置
- 📝 **進階日誌**：基於 Zap 的結構化日誌，支援日誌輪轉
- 🔄 **後台任務**：內建的後台任務執行器系統
- 🛡️ **請求過濾**：HTTP 中間件/過濾器系統，處理橫切關注點
- 🎯 **統一錯誤處理**：服務錯誤自動轉換為標準化 HTTP 服務錯誤

## 🚀 快速開始

### 安裝

```bash
go get github.com/chuccp/go-web-frame
```

### 🏠 Hello World 範例

```go
package main

import (
    "context"
    "time"

    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/log"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    // 建立 Web 框架實例，自動載入配置
    builder := wf.NewBuilder(config.LoadAutoConfig())

    // 註冊簡單路由
    builder.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })

    // 構建應用
    app := builder.Build()

    // 使用上下文執行服務，支援優雅關閉
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 範例：10 秒後自動關閉
    go func() {
        time.Sleep(time.Second * 10)
        cancel()
    }()

    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

### 🔌 REST 控制器範例

```go
package main

import (
    "context"

    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/log"
    "github.com/chuccp/go-web-frame/web"
)

type UserController struct {
    //嵌入 core.IService 介面
    core.IService
}

// 初始化控制器並註冊路由
func (u *UserController) Init(ctx *core.Context) error {
    // 通過上下文註冊路由
    ctx.Get("/users", u.GetUsers)
    ctx.Get("/users/:id", u.GetUser)
    ctx.Post("/users", u.CreateUser)

    return nil
}

// 處理器：獲取所有使用者
func (u *UserController) GetUsers(c *web.Request) (any, error) {
    // 範例：存取查詢參數
    page := c.Query("page")
    limit := c.Query("limit")

    return map[string]any{
        "users": []string{"alice", "bob"},
        "page":  page,
        "limit": limit,
    }, nil
}

// 處理器：根據 ID 獲取單個使用者
func (u *UserController) GetUser(c *web.Request) (any, error) {
    id := c.Param("id")
    return map[string]any{
        "id":   id,
        "name": "alice",
    }, nil
}

// 處理器：建立新使用者
func (u *UserController) CreateUser(c *web.Request) (any, error) {
    var user struct {
        Name string `json:"name"`
    }
    if err := c.BindJSON(&user); err != nil {
        return nil, err
    }

    return map[string]any{
        "id":   1,
        "name": user.Name,
    }, nil
}

func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Rest(&UserController{})
    app := builder.Build()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

### Static Files 和 Reverse Proxy

```go
func (c *AssetsController) Init(ctx *core.Context) error {
    ctx.Static("/assets", "./public")
    ctx.ReverseProxy("/api", "http://127.0.0.1:8081")
    return nil
}
```

`Context.Static()` 用於掛載目前服務上的本地靜態目錄，`Context.ReverseProxy()` 用於把指定前綴轉發到上游服務。

### Context Path（路由前綴）

類似 Tomcat 的 context path，可以為所有路由設定全局前綴：

```yaml
# application.yml
web:
  server:
    port: 8080
    context_path: /api
```

配置後：
- 註冊路由 `/users` → 訪問地址 `/api/users`
- 註冊路由 `/orders` → 訪問地址 `/api/orders`
- WebSocket `/ws` → 訪問地址 `/api/ws`
- 靜態文件 `/assets` → 訪問地址 `/api/assets`

### WebSocket 支援

```go
// 簡單的 Echo 伺服器
ctx.WebSocket("/ws", func(conn *websocket.Conn) error {
    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        err = conn.WriteMessage(messageType, message)
        if err != nil {
            return err
        }
    }
})

// 自定義 Upgrader
upgrader := &websocket.Upgrader{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
    CheckOrigin: func(r *http.Request) bool {
        return r.Header.Get("Origin") == "https://example.com"
    },
}
ctx.WebSocket("/ws/chat", handler, upgrader)
```

### Server-Sent Events (SSE) 支援

```go
ctx.SSE("/events", func(stream *web.SSEStream) error {
    // 設定響應頭
    stream.SetHeaders()

    // 設定重連間隔（斷開後 3 秒重連）
    stream.SendRetry(3000)

    // 發送事件
    for i := 0; i < 10; i++ {
        // 發送帶事件名的消息
        stream.Send("update", fmt.Sprintf("Count: %d", i))

        // 或發送普通消息
        // stream.SendMessage("plain message")

        // 或發送帶 ID 的消息
        // stream.SendWithID("123", "event", "data")

        time.Sleep(time.Second)
    }
    return nil
})
```

**SSE Stream 方法：**
| 方法 | 說明 |
|------|------|
| `Send(event, data)` | 發送帶事件名的消息 |
| `SendMessage(data)` | 發送普通消息 |
| `SendWithID(id, event, data)` | 發送帶 ID 的消息 |
| `SendRetry(ms)` | 設定重連間隔 |
| `Heartbeat()` | 發送心跳註釋 |
| `StartHeartbeat(interval)` | 啟動心跳協程 |

### 🏷️ 路由元數據 `.WithMeta()` 用法

`.WithMeta()` 功能允許你為單個路由附加自定義元數據，過濾器可以讀取這些元數據實現靈活的橫切關注點，例如認證、權限檢查、功能開關、快取配置等等。

**基本用法：**
```go
// 建立元數據選項
func RequireAuth() web.MetaOption {
    return web.WithValue("require_auth", true)
}

func RequirePermission(perm string) web.MetaOption {
    return web.WithValue("require_permission", perm)
}

func SkipAuth() web.MetaOption {
    return web.WithValue("skip_auth", true)
}

// 在路由註冊時使用
func (c *ApiController) Init(ctx *core.Context) error {
    // 公開路由 - 不需要認證
    ctx.Get("/api/login", loginHandler).WithMeta(SkipAuth())

    // 受保護路由 - 需要認證
    ctx.Get("/api/profile", profileHandler).WithMeta(RequireAuth())

    // 受保護路由，多個元數據，同時需要認證和權限
    ctx.Post("/api/admin/users", createUserHandler).WithMeta(RequireAuth(), RequirePermission("admin:create_user"))

    return nil
}
```

**在過濾器中存取元數據：**
```go
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    meta := req.HandlerMeta()

    // 檢查此路由是否要求認證
    requireAuth, ok := meta.Get("require_auth").(bool)
    if ok && requireAuth {
        // 如果標記了跳過認證，則跳過檢查
        if meta.Has("skip_auth") {
            return fc.Next()
        }
        // 獲取令牌並驗證...
        token := req.Request().Header.Get("Authorization")
        if token == "" {
            return nil, errors.New("缺少授權令牌")
        }
    }

    return fc.Next()
}
```

完整範例請查看：[example/withmeta/withmeta.go](./example/withmeta/withmeta.go)

### ⚡ 泛型 ORM 操作範例

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

// 定義實體結構體
type User struct {
    Id   uint   `gorm:"primaryKey"`
    Name string
    Age  int
}

// UserModel 繼承泛型 Model
type UserModel struct {
    *model.Model[*User]
}

func (u *UserModel) Init(database *db.DB, ctx *core.Context) error {
    u.Model = model.NewModel[*User](database, "t_user")
    // 如果表不存在則自動建立
    return u.CreateTable()
}

func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Model(&UserModel{})

    // ORM 操作範例
    builder.Get("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // 鏈式 API 查詢
        users, err := userModel.Query().
            Where("age > ?", 18).
            Order("id desc").
            All()

        return users, err
    })

    builder.Post("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // 建立使用者
        user := &User{Name: "張三", Age: 25}
        err := userModel.Save(user)
        return user.Id, err
    })

    builder.Put("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // 更新使用者
        return nil, userModel.Update().
            Where("id = ?", id).
            UpdateColumn("name", "張三（已更新）")
    })

    builder.Delete("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // 刪除使用者
        return nil, userModel.Delete().
            Where("id = ?", id).
            Delete()
    })

    app := builder.Build()
    ctx := context.Background()
    app.Run(ctx)
}
```

### 為什麼不用 GORM 自帶的泛型？

GORM v1.30.0+ 引入了泛型 API（`gorm.G[T](db)`），但本框架的 ORM 層有明顯優勢：

| 特性 | Go Web Frame `Model[T]` | GORM `gorm.G[T]` |
|------|--------------------------|-------------------|
| 表名綁定 | 建構時自動綁定 | 每次查詢都要手動 `.Table()` |
| Context 傳播 | `WithContext(ctx)` 一次注入，自動傳播 | 每次操作都要傳 `ctx` |
| 分頁 | 內建 `Page` / `PageForWeb` | 無 |
| `EntryModel` 便捷方法 | `FindByPK`、`DeleteByPK`、`UpdateByPK`、`UpdateColumn` | 無 |
| Query/Update/Delete | 獨立的類型安全建構器（`Query[T]`、`Update[T]`、`Delete[T]`） | 統一的 `ChainInterface[T]` |
| 自動建表 | `CreateTable()` 自動遷移 | 需手動 `AutoMigrate` |

**對比範例：**

```go
// Go Web Frame — context 一次注入，表名自動綁定
m := userModel.WithContext(req.Ctx())
user, err := m.FindByPK(1)
users, err := m.Query().Where("age > ?", 18).All()
users, total, err := m.Page(&web.Page{PageNo: 1, PageSize: 10})

// GORM 自帶泛型 — 每次操作傳 ctx，手動指定表名
user, err := gorm.G[User](db).Table("t_user").Where("id = ?", 1).First(ctx)
users, err := gorm.G[User](db).Table("t_user").Where("age > ?", 18).Find(ctx)
// 沒有內建分頁
```

本框架的 ORM 是建立在 GORM 之上的更高層抽象——保留了 GORM 的全部能力，同時提供了 GORM 泛型 API 所沒有的表名管理、分頁、context 傳播和便捷方法。

## 📊 效能對比

| 框架 | QPS | 記憶體佔用 | 特點 |
|------|-----|---------|------|
| Go Web Frame | 42k | 12MB | 全功能、低開銷 |
| Gin | 45k | 8MB | 輕量、無內建功能 |
| Beego | 32k | 25MB | 重量級、內建功能全 |
| Iris | 38k | 18MB | 功能豐富、API 複雜 |

## 🎯 適用場景

- ✅ 企業級 Web 應用開發
- ✅ RESTful API 服務
- ✅ 後台管理系統
- ✅ 微服務開發
- ✅ 快速原型開發

## 🏗️ 架構概述

### 核心層級

1. **核心抽象層 (`./core`)**：定義基礎介面和 DI 容器
   - `IService`：所有需要初始化的服務基介面
   - `IModel`：資料存取層介面，包含 CRUD 和表管理
   - `IRest`：REST 控制器介面（繼承 IService）
   - `IService`：所有服務和組件的基介面
   - `IRunner`：後台任務執行器（繼承 IService）
   - `IFilter`：HTTP 請求過濾器/中間件（繼承 IService）
   - `Context`：管理所有組件的依賴注入容器

2. **Web 層 (`./web`)**：基於 Gin 的 HTTP 處理
   - 帶輔助方法的請求/響應抽象
   - 支援所有 HTTP 方法的路由
   - 過濾器/中間件系統
   - 服務響應自動轉換為標準化 HTTP 響應

3. **資料存取層 (`./db`, `./model`)**：資料庫抽象和 ORM
   - `./db`：基於 GORM 的多資料庫抽象（MySQL、SQLite、PostgreSQL）
   - `./model`：類型安全泛型基礎模型，提供零樣板 CRUD
   - `./sqlite`：SQLite 特定配置
   - `./redis`：Redis 快取和訊息整合
   - 可配置連線池參數，適合生產環境效能調優

4. **基礎設施組件**：
   - `./config`：使用 Viper 進行配置管理（支援 JSON/YAML/TOML）
   - `./log`：基於 Zap 的結構化日誌
   - `./component`：可復用組件（快取、限流、驗證碼、二維碼、定時任務、驗證）
   - `./util`：全面的工具函數（字串、時間、加密、網路等）

### 應用生命週期

1. **建立**：使用 `NewBuilder(config)` 初始化 `Builder`
2. **配置**：透過 Builder 方法鏈添加路由、控制器、模型、服務、組件和執行器
3. **構建**：使用 `builder.Build()` 建立 `WebFrame` 實例
4. **執行**：使用 `app.Run(ctx)` 啟動伺服器

## 配置範例

### 完整配置範例

```yaml
web:
  # 伺服器配置
  server:
    port: 8080                    # 服務埠，預設 19009
    locations:                    # 靜態檔案目錄（可選）
      - view/dist
      - www
    page404: 404.html             # 404 頁面（可選）
    # HTTPS/SSL 配置（可選）
    ssl:
      enabled: true               # 是否啟用 HTTPS
      hosts:                      # 域名列表（自動申請 Let's Encrypt 憑證）
        - example.com
        - api.example.com

  # 資料庫配置
  db:
    type: mysql                   # 資料庫類型: mysql, postgres, sqlite
    host: localhost
    port: 3306
    user: root                    # 使用者名稱（也支援 username）
    password: your_password
    database: your_database       # 資料庫名（也支援 dbname）
    charset: utf8mb4
    # 連線池設定（可選，有預設值）
    max_open_conns: 100           # 最大開啟連線數，預設 100
    max_idle_conns: 10            # 最大空閒連線數，預設 10
    conn_max_lifetime: 3600       # 連線最大生命週期（秒），預設 3600

  # 日誌配置
  log:
    level: info                   # 日誌級別: debug, info, warn, error
    path: ./logs/app.log          # 日誌檔案路徑
    write: false                  # 是否後台寫入模式
    # 日誌輪轉配置（可選，有預設值）
    max_size: 100                 # 單個日誌檔案最大大小 (MB)，預設 500
    max_backups: 5                # 保留的舊日誌檔案最大數量，預設 3
    max_age: 7                    # 保留舊日誌檔案的最大天數，預設 30
    compress: true                # 是否壓縮舊日誌檔案，預設 true
    local_time: false             # 是否使用本地時間，預設 false

  # Redis 配置（可選）
  redis:
    addr: localhost:6379          # Redis 位址
    password: ""                  # 密碼
    db: 0                         # 資料庫編號

# 本地快取配置（可選）
local_cache:
  path: ./cache                   # 快取檔案儲存路徑
  open: true                      # 是否啟用檔案快取

# 限流配置（可選）
rate_limit:
  limit: 600                      # 令牌填充間隔（秒）
  burst: 5                        # 令牌桶容量
  max_size: 1000000               # 快取最大條目數
  expiry: 3600                    # 快取過期時間（秒）
```

### 資料庫配置

#### MySQL 配置

```yaml
web:
  db:
    type: mysql
    host: localhost
    port: 3306
    user: root                    # 使用者名稱（也支援 username）
    password: your_password
    database: your_database       # 資料庫名（也支援 dbname）
    charset: utf8mb4              # 可選，預設 utf8
    max_open_conns: 100           # 可選，預設 100
    max_idle_conns: 10            # 可選，預設 10
    conn_max_lifetime: 3600       # 可選，預設 3600 秒
```

#### PostgreSQL 配置

```yaml
web:
  db:
    type: postgres                # 或 postgresql
    host: localhost
    port: 5432
    user: postgres
    password: your_password
    database: your_database
    sslmode: disable              # 可選: disable, require, verify-ca, verify-full
    timezone: Asia/Shanghai       # 可選
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600
```

#### SQLite 配置

```yaml
web:
  db:
    type: sqlite
    file_path: ./data/app.db      # 資料庫檔案路徑
    max_open_conns: 10            # 可選，預設 10
    max_idle_conns: 5             # 可選，預設 5
    conn_max_lifetime: 3600       # 可選，預設 3600 秒
```

### HTTPS 配置

框架支援自動申請和管理 Let's Encrypt SSL 憑證，無需手動配置憑證檔案。

```yaml
web:
  server:
    port: 443                     # HTTPS 預設埠
    ssl:
      enabled: true               # 啟用 HTTPS
      hosts:                      # 域名列表
        - example.com
        - api.example.com
```

**HTTPS 配置說明：**

- `enabled: true` - 啟用 HTTPS 模式
- `hosts` - 需要申請憑證的域名列表
- 憑證會自動申請並快取到 `./certs` 目錄
- 支援 HTTP/2 協議
- 埠 80 會自動設定 HTTP 到 HTTPS 的重定向

**注意事項：**

1. 域名必須正確解析到伺服器 IP
2. 伺服器需要能存取外網（Let's Encrypt 驗證）
3. 建議使用 443 埠，其他埠也可用

### 靜態檔案配置

```yaml
web:
  server:
    port: 8080
    locations:                    # 靜態檔案目錄列表
      - view/dist                 # 前端建構產物
      - www                       # 靜態資源目錄
    page404: 404.html             # SPA 應用 404 回退頁面
```

**靜態檔案說明：**

- `locations` - 靜態檔案查找目錄，按順序搜索
- `page404` - 當請求 HTML 頁面但檔案不存在時返回的 404 頁面
- 支援 SPA 應用的路由回退

### Redis 配置

```yaml
web:
  redis:
    addr: localhost:6379          # Redis 位址
    password: ""                  # 密碼（可選）
    db: 0                         # 資料庫編號
    pool_size: 10                 # 連線池大小（可選）
```

### 本地快取配置

```yaml
local_cache:
  path: ./cache                   # 快取檔案儲存路徑
  open: true                      # 是否啟用檔案快取
```

### 限流配置

```yaml
rate_limit:
  limit: 600                      # 令牌填充間隔（秒），每 limit 秒填充 1 個令牌
  burst: 5                        # 令牌桶容量
  max_size: 1000000               # 快取最大條目數
  expiry: 3600                    # 快取過期時間（秒）
```

## 專案結構

```
├── web_frame.go         # 主入口 - WebFrame 工廠方法
├── core/                # 核心抽象和 DI 容器
│   ├── interface.go     # 核心介面（IService, IModel, IRest 等）
│   ├── context.go       # 依賴注入上下文
│   ├── server.go        # 管理 REST 分組和執行器的服務端實現
│   └── db.go            # DB 封裝
├── web/                 # 基於 Gin 的 Web 層
│   ├── handles.go       # 路由註冊
│   ├── request.go       # 帶輔助方法的請求抽象
│   ├── response.go      # 響應轉換
│   └── filter.go        # 過濾器/中間件介面
├── db/                  # 資料庫抽象層
│   ├── db.go            # 資料庫建立和配置解析
│   ├── mysql.go         # MySQL 配置和連線
│   └── sqlite.go        # SQLite 配置和連線
├── model/               # 泛型 ORM 實現
│   └── model.go         # 带 CRUD 操作的基礎 Model
├── sqlite/              # SQLite 驅動
├── redis/               # Redis 整合
├── config/              # 配置管理
├── log/                 # Zap 日誌
├── component/           # 可復用組件
│   ├── cache.go         # 快取組件
│   ├── localcache.go    # 本地記憶體快取
│   ├── rate_limit.go    # 限流
│   ├── captcha.go       # 驗證碼生成
│   ├── qrcode.go        # 二維碼生成
│   ├── cron.go          # Cron 定時任務
│   └── validate.go      # 輸入驗證
├── util/                # 工具函數
└── example/             # 範例應用
    ├── helloworld/      # 基礎 hello world 範例
    ├── rest/            # REST 控制器範例
    ├── model/           # ORM 模型範例
    ├── filter/          # 自定義 HTTP 過濾器範例
    ├── background/      # 後台任務執行器範例
    └── withmeta/        # 路由元數據 .WithMeta() 範例
```

## 🛠️ 開發命令

### 建構和執行範例

```bash
# 執行 hello world 範例
go run example/helloworld/helloworld.go

# 執行 REST 範例
go run example/rest/rest.go

# 執行 ORM 模型範例
go run example/model/model.go

# 執行過濾器範例
go run example/filter/filter.go

# 執行後台任務範例
go run example/background/background.go

# 執行路由元數據 .WithMeta() 範例
go run example/withmeta/withmeta.go

# 建構框架（僅函式庫檔案）
go build
```

### 測試

```bash
# 執行所有測試
go test ./...

# 執行特定套件的測試
go test ./core
go test ./web

# 執行測試並輸出詳細資訊
go test -v ./core

# 執行特定測試案例
go test -v ./core -run TestSpecificFunction
```

### 格式化和程式碼檢查

```bash
# 使用 gofmt 格式化所有程式碼
gofmt -w ./...

# 使用 gofumpt 格式化（如果已安裝）
gofumpt -w ./...

# 安裝 linter
go install golang.org/x/lint/golint@latest

# 執行程式碼檢查
golint ./...
```

### 依賴管理

```bash
# 添加新依賴
go get github.com/example/package

# 更新依賴
go get -u ./...

# 整理 go.mod 和 go.sum
go mod tidy
```


## 開發說明

- 框架遵循 Go 約定，使用標準 Go 工具鏈
- 所有組件都實現 `IService` 介面的 `Init(ctx)` 方法
- 依賴注入通過上下文完成 - 使用 `wf.GetService[T](ctx)` 獲取服務
- 連線池有合理的預設值，適合大多數應用
- 開發和小型應用推薦使用 SQLite，生產環境推薦使用 MySQL

## 文件

- **[📖 使用手冊（繁體中文）](./docs-site/docs-zh/index.md)** - 完整使用手冊：安裝、路由、控制器、模型、過濾器、配置、日誌、後台任務、組件、API 參考
- **[📖 使用手冊（英文）](./docs-site/docs-en/index.md)** - 英文版使用手冊
- **[架構設計](./ARCHITECTURE.md)** - 內部架構和設計決策
- **[最佳實踐](./BEST_PRACTICES.md)** - 推薦的模式和實踐
- **[更新日誌](./CHANGELOG.md)** - 版本歷史和變更
- **[CLAUDE.md](./CLAUDE.md)** - AI 輔助開發指南
- [Go 參考文件](https://pkg.go.dev/github.com/chuccp/go-web-frame)
- [範例應用](./example/)

## 貢獻

歡迎提交 Issue 和 Pull Request！提交 PR 前請：

1. 執行測試確保一切通過
2. 保持程式碼風格與專案一致
3. 為新功能添加適當的測試
4. 更新相關文件

## 授權

MIT License - 詳見 [LICENSE](./LICENSE)
