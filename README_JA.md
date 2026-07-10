# Go Web Frame

**Go で CRUD バックエンドを構築する際の ORM ボイラープレートをゼロに。struct を定義し、ジェネリック Model を埋め込むだけ——型安全なクエリ、ページネーション、コンテキスト伝播が含まれています。**

---

**🌐 Language / 言語**
[English](README.md) • [中文](README_ZH.md) • [繁體中文](README_ZH_TW.md) • [日本語](README_JA.md)

---

## 得られるもの

Go Web Frame は統合されたバックエンドツールキットです。ルーター、ORM、キャッシュを個別に選んで配線する必要はありません——すべて事前統合済みです。

最大の特徴：**CRUD ボイラープレートを排除するジェネリック Model 層。** エンティティ struct を定義し、`Model[T]` を埋め込むだけで、コンパイラがデータベースからハンドラまで型をチェックします。`interface{}` もコード生成も不要です。

```go
// 一度定義すれば、どこでも使える
type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement"`
    Name string
}

type UserModel struct {
    *model.EntryModel[*User, uint]   // ← すべての CRUD メソッドがここから
}

func (m *UserModel) Init(db *db.DB, c *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](db, "t_user")
    return m.CreateTable()
}

// これだけ。あとはどこでも使うだけ：
user, _ := userModel.FindByPK(1)          // → *User
users, _ := userModel.FindAll()            // → []*User
users, _ := userModel.Query().Where("age > ?", 18).Order("id desc").All()
total, _ := userModel.Query().Where("status = ?", 1).Count()
users, total, _ := userModel.Page(&web.Page{PageNo: 1, PageSize: 10})
userModel.Update().Where("id = ?", 1).UpdateColumn("name", "新しい名前")
userModel.DeleteByPK(1)
```

さらに同梱：ルーティング（Gin）、認証フィルター、WebSocket、SSE、CORS、レート制限、cron、バリデーション、キャッシュ、Let's Encrypt HTTPS、マルチ DB（MySQL / PostgreSQL / SQLite / Redis）——すべて 1 つの YAML で設定できます。

---

## クイックスタート

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

### 5 分：REST + データベース

`application.yml` を作成：

```yaml
web:
  db:
    type: sqlite
    path: ./data.db
```

`main.go` を作成：

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

// ── エンティティ ──
type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement"`
    Name string `gorm:"size:255"`
}

// ── Model（CRUD ボイラープレートゼロ）──
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

テーブルは自動作成。CRUD はすべて動作。SQL も ORM 配線コードも不要です。

---

## データ操作

### Model の 2 階層

| 型 | 提供メソッド | 用途 |
|---|---|---|
| `Model[T]` | `Save`、`Query()`、`Update()`、`Delete()`、`CreateTable()`、`WithContext()` | フルコントロール、流暢なビルダー |
| `EntryModel[T, PK]` | Model の全機能 + `FindByPK`、`FindAll`、`DeleteByPK`、`UpdateByPK`、`UpdateColumn`、`Page` | エンティティに主キーがある場合（最も一般的） |

`PK` は `uint`、`int`、`string` など、`~uint | ~int | ~string` 制約を満たす任意の型です。

### 日常的なクエリ

```go
m := userModel.WithContext(req.Ctx())   // context がすべての DB 呼出に自動伝播

// 取得
user, err := m.FindByPK(1)                          // 主キーで
users, err := m.FindAll()                            // 全件
user, err := m.Query().Where("email = ?", email).One()
users, err := m.Query().Where("status = ?", 1).Order("id desc").List(100)

// ページネーション
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := m.Query().Where("age > ?", 18).Page(page)
pageAble, err := m.Query().Where("age > ?", 18).PageForWeb(page)  // PageAble[*User] を返す

// カウント
count, err := m.Query().Where("status = ?", 1).Count()

// アソシエーション（GORM Preload）
users, err := m.Query().Preload("Orders").Preload("Profile").All()
user, err := m.Query().Where("id = ?", 1).Preload("Orders").One()

// Join
users, err := m.Query().Joins("JOIN orders ON orders.user_id = t_user.id").All()
```

### 日常的な書き込み

```go
// 挿入
err := m.Save(&User{Name: "alice"})

// 主キーで更新
user.Name = "新しい名前"
err := m.UpdateByPK(user)

// 単一カラム更新
err := m.UpdateColumn(1, "status", 0)

// 条件付き更新
err := m.Update().Where("status = ?", 0).UpdateForMap(map[string]any{"status": 1})

// 削除
err := m.DeleteByPK(1)
err := m.Delete().Where("status = ?", -1).Delete()
```

### 生 GORM との比較

```go
// Go Web Frame — テーブル名は構築時に一度だけ、context も一度、ページネーション内蔵
m := userModel.WithContext(req.Ctx())
user, _ := m.FindByPK(1)
users, total, _ := m.Query().Where("age > ?", 18).Page(&web.Page{PageNo: 1, PageSize: 10})

// 生 GORM — GetGorm() で *gorm.DB を取得、GORM エコシステムを直接使用
var user User
db.GetGorm().WithContext(ctx).Table("t_user").Where("id = ?", 1).First(&user)
var users []User
db.GetGorm().WithContext(ctx).Table("t_user").Where("age > ?", 18).Offset(0).Limit(10).Find(&users)
var total int64
db.GetGorm().WithContext(ctx).Table("t_user").Where("age > ?", 18).Count(&total)
```

`EntryModel` は便利なラッパーであり、ボイラープレートを減らすだけで制限するものではありません。組み込みメソッドで不十分な場合（複雑な JOIN、サブクエリ、ウィンドウ関数など）、`db.GetGorm()` を呼んで [GORM](https://gorm.io/docs/) エコシステムを直接使用できます。両方のスタイルを自由に混在可能です。

生 GORM ではすべての呼出でテーブル名、コンテキスト、型を繰り返します。Go Web Frame は構築時に一度バインドするだけ。20 以上のモデルがあるときに違いが顕著になります。

---

## プロジェクト構成

### Builder：すべてを一箇所で登録

```go
builder := wf.NewBuilder(config.LoadAutoConfig())

// インフラ（最初に初期化）
builder.Filter(&cors.Filter{})        // CORS ヘッダー
builder.Filter(&AuthFilter{})         // 認証チェック

// データ層
builder.Model(&UserModel{})
builder.Model(&OrderModel{})
builder.Model(&ProductModel{})

// ビジネスロジック
builder.Service(&UserService{})       // 共有ビジネスロジック
builder.Service(&PaymentService{})

// HTTP 層
builder.Rest(&UserController{})
builder.Rest(&OrderController{})

// バックグラウンド
builder.Runner(&CleanupTask{})

// 起動
app := builder.Build()
app.Run(ctx)
```

登録順が初期化順を決定：Filter → Model → Service → Controller → Runner。依存関係は `Init()` を通じて自動的に注入されます。

### 実行時の依存取得

```go
// 任意の Init() またはハンドラ内で型指定で取得：
userModel := wf.GetModel[*UserModel](ctx)
userService := wf.GetService[*UserService](ctx)
authFilter := wf.GetFilter[*AuthFilter](ctx)
cleanupTask := wf.GetRunner[*CleanupTask](ctx)
```

文字列キーなし、型アサーションなし。ジェネリック関数がすべて処理します。

### Service 層：ビジネスロジックの共有

複数の Controller でロジックを共有する場合、Service に抽出します：

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

// 登録：
builder.Service(&UserService{})

// Controller で使用：
userService := wf.GetService[*UserService](ctx)
users, _ := userService.GetActiveUsers()
```

### Model Group：複数データベース

```go
// デフォルトデータベース
builder.Model(&UserModel{}, &LogModel{})

// 分析用に独立したデータベース
analyticsGroup := wf.NewModelGroupBuilder().
    Name("analytics").
    DB(analyticsDB).
    Model(&ReportModel{}).
    AutoCreateTable(true).
    Build()
builder.ModelGroup(analyticsGroup)
```

---

## 認証とミドルウェア

### ルート単位メタデータ（WithMeta）

ルートで宣言し、Filter で一括処理——各ハンドラに認証ロジックを書く必要なし：

```go
func (c *ApiController) Init(ctx *core.Context) error {
    // 公開
    ctx.Get("/api/login", login).WithMeta(SkipAuth())

    // 要ログイン
    ctx.Get("/api/profile", profile).WithMeta(RequireAuth())

    // 要ログイン + 権限
    ctx.Post("/api/admin/users", createUser).
        WithMeta(RequireAuth(), RequirePermission("admin:create_user"))
    return nil
}
```

```go
// メタデータファクトリ
func RequireAuth() web.MetaOption      { return web.WithValue("require_auth", true) }
func SkipAuth() web.MetaOption          { return web.WithValue("skip_auth", true) }
func RequirePermission(p string) web.MetaOption { return web.WithValue("require_permission", p) }
```

```go
// 一つの Filter ですべての認証ロジックを処理
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if !req.HasMeta(RequireAuth()) || req.HasMeta(SkipAuth()) {
        return fc.Next()
    }

    token := req.Request().Header.Get("Authorization")
    if token == "" {
        return nil, errors.New("認証が必要です")
    }

    // トークン検証、権限チェック...
    return fc.Next()
}
```

### グローバル Filter

`builder.Filter()` で登録された Filter はすべてのリクエストに適用されます：

```go
// ロギング
type LoggingFilter struct { core.IFilter }

func (f *LoggingFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    start := time.Now()
    result, err := fc.Next()
    log.Info("リクエスト", zap.String("path", req.FullPath()), zap.Duration("elapsed", time.Since(start)))
    return result, err
}

// CORS（内蔵、登録するだけ）
builder.Filter(&cors.Filter{})
```

### RestGroup：ルートグループ単位の Filter

```go
apiGroup := core.NewRestGroupBuilder().
    ServerConfig(web.DefaultServerConfig()).
    ContextPath("/api/v1").
    Build()

apiGroup.AddFilter(&AuthFilter{})     // この Group のルートにのみ適用
apiGroup.AddRest(&UserController{})   // この Controller の全ルートに認証が必要

builder.RestGroup(apiGroup)
```

---

## よく使うレシピ

### ファイルアップロード

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
        stream.Write(stream.Context(), typ, msg)  // エコー
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

### 静的ファイル & SPA

```go
ctx.Static("/assets", "./public")
ctx.Static("/", "./frontend/dist")   // SPA ビルド出力
```

```yaml
# または設定で——複数ディレクトリ検索、自動 404 フォールバック：
web:
  server:
    locations:
      - view/dist
      - www
    page404: 404.html
```

### リバースプロキシ

```go
ctx.ReverseProxy("/api/legacy", "http://127.0.0.1:8081")
```

### バックグラウンドタスク

```go
type CleanupTask struct { core.IRunner }

func (t *CleanupTask) Init(ctx *core.Context) error { return nil }

func (t *CleanupTask) Run() error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        // 期限切れセッションのクリーンアップ...
    }
    return nil
}

builder.Runner(&CleanupTask{})
```

### バリデーション

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

    // input は検証済み、処理を続行...
}
```

### カスタム HTTP レスポンス

```go
// 通常の struct を返す → 自動ラップ：{"code":200, "data":{...}, "msg":"ok"}
return &User{Id: 1, Name: "alice"}, nil

// 文字列を返す → プレーンテキスト
return "ok", nil

// ステータスコード指定
return web.DataCode(http.StatusCreated, &user), nil

// エラーを返す
return nil, errors.New("問題が発生しました")

// ビジネスエラーコード
return nil, web.NewValidationError().WithDetail("名前は必須です")

// リダイレクト
return web.Redirect("/new-url"), nil

// ファイルダウンロード
return &web.FileResponse{Path: "/path/to/report.pdf", FileName: "report.pdf"}, nil
```

---

## 設定

1 つの YAML ですべてをカバー。`./config/`、`~/.<appname>/`、`/etc/<appname>/` から自動読み込み。

```yaml
web:
  server:
    port: 8080
    context_path: /api            # グローバルルートプレフィックス
    ssl:                          # HTTPS：自動証明書またはローカル証明書
      enabled: true
      hosts:                      # Let's Encrypt 自動証明書取得
        - example.com
      # certs:                    # またはローカル証明書ファイルを使用（オプション）
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
    max_size: 500                 # ローテーション閾値（MB）
    max_backups: 3
    max_age: 30                   # 保持日数
    compress: true

  redis:
    addr: localhost:6379
    password: ""
    db: 0
```

PostgreSQL と SQLite：

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

JSON、YAML、TOML 形式に対応。

### カスタムデータベースドライバー

フレームワークに内蔵されていないデータベース（SQL Server、ClickHouse など）を使用する場合、カスタムドライバーを登録できます。フレームワークは基盤 ORM として GORM を使用しているため、カスタムデータベースには対応する [GORM ドライバー](https://gorm.io/docs/connecting_to_the_database.html) が必要です。

```go
// 1. db.IConfig インターフェースを実装
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

// 2. アプリケーション起動前に登録
func main() {
    db.RegisterDB("clickhouse", &ClickHouseConfig{})

    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Build().Start()
}
```

設定ファイルで `type` に登録したタイプ名を指定：

```yaml
web:
  db:
    type: clickhouse
    host: localhost
    port: 9000
    database: mydb
```

---

## 技術スタック

事前統合された、本番環境で検証済みのコンポーネント：

| 層 | ライブラリ | 役割 |
|---|---|---|
| HTTP | Gin | ルーティング、ミドルウェアチェーン、パラメータバインディング |
| ORM | GORM | SQL ドライバ、マイグレーション、join/preload |
| 設定 | Viper | 複数フォーマット、複数パス自動読み込み |
| ログ | Zap | 構造化、レベル別、ローテーション |
| JSON | Sonic | 高速シリアライゼーション |
| キャッシュ | Otter | ローカルインメモリキャッシュ |
| 並行性 | Conc | 構造化並行性プール、サーバーライフサイクル管理 |
| Redis | go-redis | Pub/Sub、キャッシング |
| SQLite | modernc/sqlite | Pure Go、CGO 不要 |
| バリデーション | go-playground/validator | struct タグバリデーション |
| WebSocket | coder/websocket | アップグレード + 読み書き |
| Cron | robfig/cron | 式ベースのスケジューラ |

---

## プロジェクト構造

```
├── web_frame.go        # Builder、WebFrame ファクトリ
├── core/               # インターフェース（IService, IModel, IRest, IRunner, IFilter）、DI context
├── web/                # Request、Response、ルーティング、Filter、SSE、WebSocket、静的ファイル
├── model/              # Model[T]、EntryModel[T, PK]、Query[T]、Update[T]、Delete[T]
├── db/                 # MySQL、PostgreSQL、SQLite 接続管理
├── redis/              # Redis クライアントラッパー
├── config/             # Viper 自動読み込み
├── log/                # Zap + lumberjack ローテーション
├── component/          # cors、cache、rate-limit、captcha、qrcode、cron、validator
├── util/               # 暗号化、ファイル、文字列ヘルパー
└── example/            # 実行可能なサンプル
    ├── helloworld/     # 最小アプリ
    ├── rest/           # REST Controller
    ├── model/          # ジェネリック ORM の使い方
    ├── filter/         # 認証 + ロギング Filter
    ├── withmeta/       # ルートメタデータ
    └── background/     # バックグラウンドタスク
```

---

## オプションモジュール

重い依存関係は独立したサブモジュールに分割されています。必要なものだけインストール：

<!-- component-badges -->
[![cache](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/cache/*&label=cache&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/cache)
[![captcha](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/captcha/*&label=captcha&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/captcha)
[![schedule](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/schedule/*&label=schedule&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/schedule)
[![qrcode](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/qrcode/*&label=qrcode&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/qrcode)
[![ratelimit](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/ratelimit/*&label=ratelimit&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/ratelimit)
<!-- /component-badges -->

```bash
# コアフレームワーク（captcha/qrcode/cron/otter なし）
go get github.com/chuccp/go-web-frame

# 必要に応じてインストール
go get github.com/chuccp/go-web-frame/component/captcha@v1.0.7
go get github.com/chuccp/go-web-frame/component/schedule@v1.0.7
go get github.com/chuccp/go-web-frame/component/qrcode@v1.0.7
go get github.com/chuccp/go-web-frame/component/cache@v1.0.7
go get github.com/chuccp/go-web-frame/component/ratelimit@v1.0.7
```

### 使用方法

#### Cache — 高速インメモリキャッシュ

```go
import "github.com/chuccp/go-web-frame/component/cache"

// 1. 登録
builder.Service(&cache.Cache{})

// 2. コントローラやサービスで使用
c := core.GetService[*cache.Cache](ctx)
c.Set("user:1", user)
val, ok := c.Get("user:1")
c.SetNX("lock:task", true, 30*time.Second) // 存在しない場合のみ設定（TTL付き）
```

#### Captcha — スライドパズル CAPTCHA

```yaml
# application.yml
captcha:
  code_key: "your-32-character-key-here!!"  # 32文字（必須）
  code_iv: "your-16-char-iv"                # 16文字（必須）
```

```go
import "github.com/chuccp/go-web-frame/component/captcha"

// 1. 登録
builder.Service(&captcha.Captcha{})

// 2. CAPTCHA を生成（フロントエンドに返す）
c := core.GetService[*captcha.Captcha](ctx)
data, _ := c.GetCaptchaData()
// → SlideCaptchaData（TileImage, MasterImage, ThumbCode を含む）

// 3. ユーザーのスライド操作を検証
result, ok := c.ValidateThumb(data.ThumbCode, userXOffset)
if ok {
    valid, _ := c.ValidateCode(result.CaptchaCode)
}
```

#### Schedule — Cron ジョブスケジューラ

```go
import (
    "github.com/chuccp/go-web-frame/component/schedule"
    "github.com/robfig/cron/v3"
)

// 1. Runner として登録（Service ではありません！）
sched := schedule.NewSchedule(cron.WithSeconds())
builder.Runner(sched)

// 2. ジョブを追加 — Build() の前後どちらでも可能
sched.AddKeyFunc("cleanup", "0 0 * * *", func(ctx *core.Context) {
    // 毎日深夜に実行
})
sched.AddKeyFunc("report", "*/30 * * * * *", func(ctx *core.Context) {
    // 30秒ごとに実行
})
```

#### QRCode — QR コードジェネレータ

```go
import "github.com/chuccp/go-web-frame/component/qrcode"

// ファイルに生成（ストライプスタイル）
qrcode.GenerateStripeQRCode("https://example.com", "qr.png")

// メモリ上に生成、スタイルをカスタマイズ
buf := qrcode.CreateBufferWriteCloser()
qrcode.GenerateQrcode("hello", buf, qrcode.WithCircleShape())
pngBytes := buf.Bytes()

// 色をカスタマイズ
s := qrcode.NewStripeQRCode().WithModuleSize(10)
s.GenerateFile("https://example.com", "custom.png")
```

> QRCode はスタンドアロンユーティリティです。登録不要で、そのまま import して使用できます。

#### RateLimit — トークンバケットレートリミッタ

```yaml
# application.yml（任意 — デフォルト値を表示）
rate_limit:
  limit: 10     # 1秒あたりのトークン数
  burst: 5      # 最大バーストサイズ
```

```go
import "github.com/chuccp/go-web-frame/component/ratelimit"

// 1. 登録
builder.Service(&ratelimit.RateLimit{})

// 2. 使用 — 通常は Filter 内で
r := core.GetService[*ratelimit.RateLimit](ctx)
if r.Allow(req.ClientIP()) {
    return fc.Next()
}
return nil, errors.New("レート制限中")
```

### サブモジュールの公開

各サブモジュールはタグプレフィックスで独立してバージョン管理：

```bash
# 1. メインモジュールを公開
git tag v1.0.1
git push origin v1.0.1

# 2. 各サブモジュールの依存関係を更新
for mod in captcha schedule qrcode cache ratelimit; do
  cd component/$mod
  go get github.com/chuccp/go-web-frame@v1.0.1
  go mod edit -dropreplace github.com/chuccp/go-web-frame
  go mod tidy
  cd ../..
done

# 3. 各サブモジュールにタグを付ける（形式：component/<name>/vX.Y.Z）
git tag component/captcha/v1.0.1
git tag component/schedule/v1.0.1
git tag component/qrcode/v1.0.1
git tag component/cache/v1.0.1
git tag component/ratelimit/v1.0.1

# 4. すべてのタグをプッシュ
git push origin --tags
```

Go proxy は `component/captcha/v1.0.1` タグから自動的に対応ディレクトリの `go.mod` を見つけます。

---

## ヘルプ

- **[ユーザーガイド (英語)](./docs-site/docs-en/index.md)** — 完全なドキュメント
- **[ユーザーガイド (中国語)](./docs-site/docs-zh/index.md)** — 中国語ドキュメント
- **[アーキテクチャ](./ARCHITECTURE.md)** — 設計上の決定と内部実装
- **[CLAUDE.md](./CLAUDE.md)** — AI コーディングアシスタント用ガイド
- **[Go Reference](https://pkg.go.dev/github.com/chuccp/go-web-frame)** — パッケージ API ドキュメント

---

## コントリビューション

PR 歓迎。提出前に `go test ./...` を実行してください。

## ライセンス

MIT — [LICENSE](./LICENSE) を参照
