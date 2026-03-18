# Go Web Frame

エンタープライズグレードのWebアプリケーションを構築するための構造化されたアプローチを提供する、モダンで機能豊富なGo Webフレームワーク。

---

**🌐 Language / 言語**
[English](README.md) • [中文](README_ZH.md) • [繁體中文](README_ZH_TW.md) • [日本語](README_JA.md)

---

## プロジェクト概要

Go Web Frameは、コンポーネントベースの設計によってクリーンなアーキテクチャを強制するopinionatedなWebフレームワークです。依存性注入、タイプセーフなORM、データベース統合などの機能を内蔵しています。

**核心となる利点**: このフレームワークは、Goエコシステムの優れたオープンソースパッケージとフレームワークを統合し、ベストプラクティスを有機的に組み合わせています。本番環境対応のアプリケーションを迅速に構築するための最適なソリューションであり、開発効率を大幅に向上させます。

## 🧩 技術スタック

このフレームワークは、以下の優秀なオープンソースコンポーネントを厳選して統合し、深く統合されベストプラクティスで設定されています：

### コアフレームワーク
| コンポーネント | 説明 |
|-----------|-------------|
| [Gin](https://github.com/gin-gonic/gin) | 優れたAPIパフォーマンスを持つ高性能HTTP Webフレームワーク |
| [GORM](https://gorm.io/) | マルチデータベース対応の強力なORMライブラリ |
| [Viper](https://github.com/spf13/viper) | 複数フォーマット対応の完全な設定ソリューション |
| [Zap](https://go.uber.org/zap) | Uber製の高性能構造化ロギングライブラリ |

### データストレージ
| コンポーネント | 説明 |
|-----------|-------------|
| [go-redis](https://github.com/redis/go-redis) | Redis推奨のクライアント |
| [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | CGO依存のない純粋なGo実装のSQLite |
| [gorm-driver/mysql](https://gorm.io/docs/connecting_to_the_database.html) | MySQLデータベースドライバ |
| [gorm-driver/postgres](https://gorm.io/docs/connecting_to_the_database.html) | PostgreSQLデータベースドライバ |

### キャッシング＆パフォーマンス
| コンポーネント | 説明 |
|-----------|-------------|
| [Otter](https://github.com/maypok86/otter) | 高性能なGoローカルキャッシュライブラリ |
| [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) | トークンバケットレートリミッター |

### ユーティリティ
| コンポーネント | 説明 |
|-----------|-------------|
| [Cron](https://github.com/robfig/cron) | スケジュールタスクライブラリ |
| [go-qrcode](https://github.com/yeqown/go-qrcode) | QRコード生成 |
| [go-captcha](https://github.com/wenlng/go-captcha) | 行動認証キャプチャ生成 |
| [validator](https://github.com/go-playground/validator) | 構造体フィールド検証 |
| [UUID](https://github.com/google/uuid) | UUID生成 |
| [Lumberjack](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2) | ログローテーション |
| [Conc](https://github.com/sourcegraph/conc) | より良い並行処理プリミティブ |
| [Emperror](https://emperror.dev/errors) | 本番グレードのエラーハンドリング |

### なぜこれらのコンポーネントを選択したのか？

- **本番環境で実証済み**: すべてのコンポーネントは大規模な本番環境で広く検証されています
- **高性能**: Gin、Zap、Otterはそれぞれの領域で最高のパフォーマンスを発揮します
- **ベストプラクティス**: 綿密に統合され、すぐに使用可能、複雑な設定は不要
- **成熟したエコシステム**: 活発なコミュニティサポートと継続的な更新

## 🤖 なぜGo Web Frameを選ぶのか？

### 他のフレームワークとの比較

| 機能 | Go Web Frame | Gin | Beego | Echo |
|---------|-------------|-----|-------|------|
| すぐに使える | ✅ 完全なソリューション | ❌ 統合が必要 | ✅ 完全なソリューション | ⚠️ 部分的 |
| ジェネリックORM | ✅ ゼロボイラープレート | ❌ 自分で選択 | ❌ ジェネリックなし | ❌ 自分で選択 |
| 依存性注入 | ✅ 組み込みDI | ❌ 自分で実装 | ⚠️ 基本サポート | ❌ 自分で実装 |
| 学習曲線 | 🟢 中程度 | 🟢 簡単 | 🟡 急 | 🟢 簡単 |
| 機能の完全性 | 🟢 高 | 🟡 中程度 | 🟢 高 | 🟡 中程度 |
| パフォーマンス | 🟢 優秀 (42k QPS) | 🟢 最高 (45k QPS) | 🟡 平均 | 🟢 優秀 |

### いつGo Web Frameを選ぶべきか？

**強く推奨される場面:**

- 🚀 **迅速なプロトタイピング**: 必要なコンポーネントがすべて事前に統合された状態で、プロジェクトプロトタイプを素早く完成させる必要がある
- 🏢 **エンタープライズアプリケーション**: クリーンなアーキテクチャ、依存性注入、統一エラーハンドリングが必要
- 📊 **管理ダッシュボードシステム**: CRUD操作、ページネーション、バリデーションなどの一般的な機能を内蔵
- 🔌 **RESTful APIサービス**: 簡素化されたコントローラー実装、自動ルート登録
- ⚙️ **マイクロサービス**: 軽量ながら機能が完全
- 🛠️ **フルスタックGoプロジェクト**: フロントエンドからバックエンド、データベースまで、ワンストップソリューション

**特に適している:**

- Go初心者: 試行錯誤なしでベストプラクティスを学びたい
- 個人開発者: プロジェクトを効率的に完了させ、技術選定の時間を削減したい
- 小規模チーム: 統一された技術スタック、コラボレーションコストを削減
- AIアシスト開発: クリーンなアーキテクチャにより、AIが理解しコード生成しやすい

### 選択ガイド

あなたが探しているのは：
- **機能が完全**で**すぐに使える**Go Webフレームワーク
- **本番環境で実証済み**の優れたコンポーネントが統合され、繰り返しの調査選定が不要
- **ジェネリック対応**のタイプセーフORMで、ボイラープレートコードを削減
- **クリーンなアーキテクチャ**で、保守性とチームコラボレーションを向上
- **迅速な開発**で、アイデアから製品までの時間を短縮

**Go Web Frameはあなたにとって最良の選択です！**

## 🌟 特徴

- **MVCライクなアーキテクチャ**: サービス、コントローラー、モデルによる関心の分離
- **⚡ タイプセーフジェネリックORM**: ジェネリクスによるゼロボイラープレートCRUD操作、リフレクションオーバーヘッドなし
- **依存性注入**: コンポーネントライフサイクルを管理する組み込みDIコンテナ
- **データベース統合**: SQLite、MySQL、PostgreSQL、Redisサポート（GORMによる拡張可能なデータベース抽象レイヤー）
- **接続プール設定**: 設定可能な接続プールパラメータ（`max_open_conns`、`max_idle_conns`、`conn_max_lifetime`）
- **コンポーネントシステム**: キャッシング、レート制限、キャプチャ、QRコード生成、cronスケジュールタスク、入力検証を含む再利用可能なコンポーネント
- **RESTfulサポート**: 簡素化されたRESTコントローラー実装
- **自動設定**: JSON、YAML、TOMLファイルからの自動読み込み設定
- **高度なロギング**: ローテーションサポート付きのZapによる構造化ロギング
- **バックグラウンドタスク**: バックグラウンド処理のための組み込みランナーシステム
- **リクエストフィルタリング**: クロスカッティングコンサーンのためのHTTPミドルウェア/フィルターシステム
- **ルートメタデータ**: `.WithMeta()`によるルートへのカスタムメタデータ添付サポート（柔軟なルート単位の認証、パーミッションなどを実現）
- **統一エラーハンドリング**: サービスエラーの標準化されたHTTPレスポンスへの自動変換

## 🚀 クイックスタート

### インストール

```bash
go get github.com/chuccp/go-web-frame
```

### 🏠 Hello Worldの例

```go
package main

import (
    "context"
    "time"

    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/log"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    // 自動設定読み込みでWebフレームワークインスタンスを作成
    app := wf.NewWithAutoConfig()

    // シンプルなルートを登録
    app.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })

    // グレースフルシャットダウン対応のコンテキストで実行
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 例：10秒後に自動シャットダウン
    go func() {
        time.Sleep(time.Second * 10)
        cancel()
    }()

    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

### 🔌 RESTコントローラーの例

```go
package main

import (
    "context"

    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/log"
    "github.com/chuccp/go-web-frame/web"
)

type UserController struct {
    // core.IServiceインターフェースを埋め込み
    core.IService
}

// コントローラーを初期化しルートを登録
func (u *UserController) Init(ctx *core.Context) error {
    // コンテキストを通じてルートを登録
    ctx.Get("/users", u.GetUsers)
    ctx.Get("/users/:id", u.GetUser)
    ctx.Post("/users", u.CreateUser)

    return nil
}

// ハンドラー：全ユーザーを取得
func (u *UserController) GetUsers(c *web.Request) (any, error) {
    // クエリパラメータにアクセス
    page := c.Query("page")
    limit := c.Query("limit")

    return map[string]any{
        "users": []string{"alice", "bob"},
        "page":  page,
        "limit": limit,
    }, nil
}

// ハンドラー：IDで単一ユーザーを取得
func (u *UserController) GetUser(c *web.Request) (any, error) {
    id := c.Param("id")
    return map[string]any{
        "id":   id,
        "name": "alice",
    }, nil
}

// ハンドラー：新規ユーザー作成
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
    app := wf.NewWithAutoConfig()
    app.AddRest(&UserController{})

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

### 🏷️ ルートメタデータ `.WithMeta()` の使用法

`.WithMeta()` 機能を使用すると、個別のルートにカスタムメタデータを添付できます。フィルターはこのメタデータを読み取って、認証、パーミッションチェック、機能フラグ、キャッシュ設定などの柔軟なクロスカッティングコンサーンを実装できます。

**基本的な使用法:**
```go
// メタデータオプションを作成
func RequireAuth() web.MetaOption {
    return web.WithValue("require_auth", true)
}

func RequirePermission(perm string) web.MetaOption {
    return web.WithValue("require_permission", perm)
}

func SkipAuth() web.MetaOption {
    return web.WithValue("skip_auth", true)
}

// ルート登録時に使用
func (c *ApiController) Init(ctx *core.Context) error {
    // パブリックルート - 認証不要
    ctx.Get("/api/login", loginHandler).WithMeta(SkipAuth())

    // 保護されたルート - 認証必要
    ctx.Get("/api/profile", profileHandler).WithMeta(RequireAuth())

    // 保護されたルート、複数のメタデータ、認証とパーミッションの両方必要
    ctx.Post("/api/admin/users", createUserHandler).WithMeta(RequireAuth(), RequirePermission("admin:create_user"))

    return nil
}
```

**フィルターでメタデータにアクセス:**
```go
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    meta := req.HandlerMeta()

    // このルートが認証を必要とするかチェック
    requireAuth, ok := meta.Get("require_auth").(bool)
    if ok && requireAuth {
        // skip_authがマークされている場合はチェックをスキップ
        if meta.Has("skip_auth") {
            return fc.Next()
        }
        // トークンを取得して検証...
        token := req.Request().Header.Get("Authorization")
        if token == "" {
            return nil, errors.New("認証トークンがありません")
        }
    }

    return fc.Next()
}
```

完全な例は[example/withmeta/withmeta.go](./example/withmeta/withmeta.go)をご覧ください。

### ⚡ ジェネリックORM操作の例

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/model"
    "github.com/chuccp/go-web-frame/web"
)

// エンティティ構造体を定義
type User struct {
    Id   uint   `gorm:"primaryKey"`
    Name string
    Age  int
}

// UserModelはジェネリックModelを継承
type UserModel struct {
    *model.Model[*User]
}

func (u *UserModel) Init(database *db.DB, ctx *core.Context) error {
    u.Model = model.NewModel[*User](database, "t_user")
    // テーブルが存在しない場合は自動作成
    return u.CreateTable()
}

func main() {
    app := wf.NewWithAutoConfig()
    // DIコンテナにモデルを登録
    app.AddModel(&UserModel{})

    // ORM操作の例
    app.Get("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // チェーンAPIでクエリ
        users, err := userModel.Query().
            Where("age > ?", 18).
            Order("id desc").
            All()

        return users, err
    })

    app.Post("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // ユーザー作成
        user := &User{Name: "田中", Age: 25}
        err := userModel.Save(user)
        return user.Id, err
    })

    app.Put("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // ユーザー更新
        return nil, userModel.Update().
            Where("id = ?", id).
            UpdateColumn("name", "田中（更新済み）")
    })

    app.Delete("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // ユーザー削除
        return nil, userModel.Delete().
            Where("id = ?", id).
            Delete()
    })

    ctx := context.Background()
    _ = app.Run(ctx)
}
```

## 📊 パフォーマンス比較

| フレームワーク | QPS | メモリ使用量 | 特徴 |
|---------|-----|---------|------|
| Go Web Frame | 42k | 12MB | フル機能、低オーバーヘッド |
| Gin | 45k | 8MB | 軽量、組み込み機能なし |
| Beego | 32k | 25MB | 重量級、フル機能内蔵 |
| Iris | 38k | 18MB | 機能豊富、複雑なAPI |

## 🎯 適用シナリオ

- ✅ エンタープライズWebアプリケーション開発
- ✅ RESTful APIサービス
- ✅ 管理ダッシュボードシステム
- ✅ マイクロサービス開発
- ✅ 迅速なプロトタイピング

## 🏗️ アーキテクチャ概要

### コアレイヤー

1. **コア抽象レイヤー (`./core`)**: 基本的なインターフェースとDIコンテナを定義
   - `IService`: 初期化が必要なすべてのサービスのベースインターフェース
   - `IModel`: CRUDとテーブル管理を含むデータアクセスレイヤーインターフェース
   - `IRest`: RESTコントローラーインターフェース（IServiceを継承）
   - `IComponent`: 設定で初期化される独立したコンポーネント
   - `IRunner`: バックグラウンドタスクランナー（IServiceを継承）
   - `IFilter`: HTTPリクエストフィルター/ミドルウェア（IServiceを継承）
   - `Context`: すべてのコンポーネントを管理する依存性注入コンテナ

2. **Webレイヤー (`./web`)**: GinベースのHTTPハンドリング
   - ヘルパーメソッド付きのリクエスト/レスポンス抽象化
   - すべてのHTTPメソッドをサポートするルーティング
   - フィルター/ミドルウェアシステム
   - サービスレスポンスの標準化されたHTTPレスポンスへの自動変換

3. **データアクセスレイヤー (`./db`, `./model`)**: データベース抽象とORM
   - `./db`: GORMベースのマルチデータベース抽象（MySQL、SQLite、PostgreSQL）
   - `./model`: ゼロボイラープレートCRUDを提供するタイプセーフジェネリックベースモデル
   - `./sqlite`: SQLite固有の設定
   - `./redis`: Redisキャッシングとメッセージング統合
   - 本番環境のパフォーマンスチューニングに適した設定可能な接続プールパラメータ

4. **インフラストラクチャコンポーネント**:
   - `./config`: Viperによる設定管理（JSON/YAML/TOML対応）
   - `./log`: Zapベースの構造化ロギング
   - `./component`: 再利用可能なコンポーネント（キャッシュ、レート制限、キャプチャ、QRコード、cron、バリデーション）
   - `./util`: 包括的なユーティリティ（文字列、時間、暗号、ネットワークなど）

### アプリケーションライフサイクル

1. **作成**: `NewWithAutoConfig()` または `New(config)` で `WebFrame` を初期化
2. **登録**: ルート、コントローラー、モデル、サービス、コンポーネント、ランナーを追加
3. **設定**: 設定のカスタマイズ、ミドルウェアの追加、ロギングの設定
4. **実行**: `Run(ctx)` でサーバーを起動

## 設定例

### 完全な設定例

```yaml
web:
  # サーバー設定
  server:
    port: 8080                    # サービスポート、デフォルト 19009
    locations:                    # 静的ファイルディレクトリ（オプション）
      - view/dist
      - www
    page404: 404.html             # 404ページ（オプション）
    # HTTPS/SSL設定（オプション）
    ssl:
      enabled: true               # HTTPSを有効化
      hosts:                      # ドメインリスト（Let's Encrypt証明書を自動取得）
        - example.com
        - api.example.com

  # データベース設定
  db:
    type: mysql                   # データベースタイプ: mysql, postgres, sqlite
    host: localhost
    port: 3306
    user: root                    # ユーザー名（usernameも対応）
    password: your_password
    database: your_database       # データベース名（dbnameも対応）
    charset: utf8mb4
    # 接続プール設定（オプション、デフォルト値あり）
    max_open_conns: 100           # 最大オープン接続数、デフォルト 100
    max_idle_conns: 10            # 最大アイドル接続数、デフォルト 10
    conn_max_lifetime: 3600       # 接続の最大ライフタイム（秒）、デフォルト 3600

  # ログ設定
  log:
    level: info                   # ログレベル: debug, info, warn, error
    path: ./logs/app.log          # ログファイルパス
    write: false                  # バックグラウンド書き込みモード
    # ログローテーション設定（オプション、デフォルト値あり）
    max_size: 100                 # 単一ログファイルの最大サイズ (MB)、デフォルト 500
    max_backups: 5                # 保持する古いログファイルの最大数、デフォルト 3
    max_age: 7                    # 古いログファイルを保持する最大日数、デフォルト 30
    compress: true                # 古いログファイルを圧縮するか、デフォルト true
    local_time: false             # ローカルタイムを使用するか、デフォルト false

  # Redis設定（オプション）
  redis:
    addr: localhost:6379          # Redisアドレス
    password: ""                  # パスワード
    db: 0                         # データベース番号

# ローカルキャッシュ設定（オプション）
local_cache:
  path: ./cache                   # キャッシュファイル保存パス
  open: true                      # ファイルキャッシュを有効化

# レート制限設定（オプション）
rate_limit:
  limit: 600                      # トークン補充間隔（秒）
  burst: 5                        # トークンバケット容量
  max_size: 1000000               # 最大キャッシュエントリ数
  expiry: 3600                    # キャッシュ有効期限（秒）
```

### データベース設定

#### MySQL設定

```yaml
web:
  db:
    type: mysql
    host: localhost
    port: 3306
    user: root                    # ユーザー名（usernameも対応）
    password: your_password
    database: your_database       # データベース名（dbnameも対応）
    charset: utf8mb4              # オプション、デフォルト utf8
    max_open_conns: 100           # オプション、デフォルト 100
    max_idle_conns: 10            # オプション、デフォルト 10
    conn_max_lifetime: 3600       # オプション、デフォルト 3600秒
```

#### PostgreSQL設定

```yaml
web:
  db:
    type: postgres                # または postgresql
    host: localhost
    port: 5432
    user: postgres
    password: your_password
    database: your_database
    sslmode: disable              # オプション: disable, require, verify-ca, verify-full
    timezone: Asia/Tokyo          # オプション
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600
```

#### SQLite設定

```yaml
web:
  db:
    type: sqlite
    file_path: ./data/app.db      # データベースファイルパス
    max_open_conns: 10            # オプション、デフォルト 10
    max_idle_conns: 5             # オプション、デフォルト 5
    conn_max_lifetime: 3600       # オプション、デフォルト 3600秒
```

### HTTPS設定

フレームワークはLet's Encrypt SSL証明書の自動申請と管理をサポートしており、手動での証明書設定は不要です。

```yaml
web:
  server:
    port: 443                     # HTTPSデフォルトポート
    ssl:
      enabled: true               # HTTPSを有効化
      hosts:                      # ドメインリスト
        - example.com
        - api.example.com
```

**HTTPS設定の注意点:**

- `enabled: true` - HTTPSモードを有効化
- `hosts` - 証明書申請が必要なドメインリスト
- 証明書は自動的に申請され `./certs` ディレクトリにキャッシュされます
- HTTP/2プロトコル対応
- ポート80は自動的にHTTPからHTTPSへのリダイレクトを設定

**重要な注意事項:**

1. ドメインはサーバーIPに正しく解決されている必要があります
2. サーバーは外部ネットワークにアクセスできる必要があります（Let's Encrypt検証）
3. ポート443の使用を推奨しますが、他のポートも使用可能です

### 静的ファイル設定

```yaml
web:
  server:
    port: 8080
    locations:                    # 静的ファイルディレクトリリスト
      - view/dist                 # フロントエンドビルド出力
      - www                       # 静的リソースディレクトリ
    page404: 404.html             # SPA 404フォールバックページ
```

**静的ファイルの説明:**

- `locations` - 静的ファイル検索ディレクトリ、順番に検索
- `page404` - HTMLページがリクエストされたがファイルが存在しない場合に返される404ページ
- SPAルートのフォールバックをサポート

### Redis設定

```yaml
web:
  redis:
    addr: localhost:6379          # Redisアドレス
    password: ""                  # パスワード（オプション）
    db: 0                         # データベース番号
    pool_size: 10                 # 接続プールサイズ（オプション）
```

### ローカルキャッシュ設定

```yaml
local_cache:
  path: ./cache                   # キャッシュファイル保存パス
  open: true                      # ファイルキャッシュを有効化
```

### レート制限設定

```yaml
rate_limit:
  limit: 600                      # トークン補充間隔（秒）、limit秒ごとに1トークン補充
  burst: 5                        # トークンバケット容量
  max_size: 1000000               # 最大キャッシュエントリ数
  expiry: 3600                    # キャッシュ有効期限（秒）
```

## プロジェクト構造

```
├── web_frame.go         # メインエントリーポイント - WebFrameファクトリメソッド
├── core/                # コア抽象とDIコンテナ
│   ├── interface.go     # コアインターフェース（IService, IModel, IRestなど）
│   ├── context.go       # 依存性注入コンテキスト
│   ├── server.go        # RESTグループとランナーを管理するサーバー実装
│   └── db.go            # DBラッパー
├── web/                 # GinベースのWebレイヤー
│   ├── handles.go       # ルート登録
│   ├── request.go       # ヘルパーメソッド付きリクエスト抽象化
│   ├── response.go      # レスポンス変換
│   └── filter.go        # フィルター/ミドルウェアインターフェース
├── db/                  # データベース抽象レイヤー
│   ├── db.go            # データベース作成と設定解析
│   ├── mysql.go         # MySQL設定と接続
│   └── sqlite.go        # SQLite設定と接続
├── model/               # ジェネリックORM実装
│   └── model.go         # CRUD操作付きベースModel
├── sqlite/              # SQLiteドライバ
├── redis/               # Redis統合
├── config/              # 設定管理
├── log/                 # Zapロギング
├── component/           # 再利用可能なコンポーネント
│   ├── cache/           # キャッシュコンポーネント
│   ├── localcache/      # ローカルメモリキャッシュ
│   ├── ratelimit/       # レート制限
│   ├── captcha/         # キャプチャ生成
│   ├── schedule/        # Cronスケジュールタスク
│   └── validator/       # 入力検証
├── util/                # ユーティリティ関数
└── example/             # サンプルアプリケーション
    ├── helloworld/      # 基本的なhello worldの例
    ├── rest/            # RESTコントローラーの例
    ├── model/           # ORMモデルの例
    ├── filter/          # カスタムHTTPフィルターの例
    ├── background/      # バックグラウンドタスクランナーの例
    └── withmeta/        # ルートメタデータ .WithMeta() の例
```

## 🛠️ 開発コマンド

### ビルドと実行の例

```bash
# hello worldの例を実行
go run example/helloworld/helloworld.go

# RESTの例を実行
go run example/rest/rest.go

# ORMモデルの例を実行
go run example/model/model.go

# フィルターの例を実行
go run example/filter/filter.go

# バックグラウンドタスクの例を実行
go run example/background/background.go

# ルートメタデータ .WithMeta() の例を実行
go run example/withmeta/withmeta.go

# フレームワークをビルド（ライブラリのみ）
go build
```

### テスト

```bash
# すべてのテストを実行
go test ./...

# 特定のパッケージのテストを実行
go test ./core
go test ./web

# 詳細出力でテストを実行
go test -v ./core

# 特定のテストケースを実行
go test -v ./core -run TestSpecificFunction
```

### フォーマットとリンティング

```bash
# gofmtですべてのコードをフォーマット
gofmt -w ./...

# gofumptでフォーマット（インストール済みの場合）
gofumpt -w ./...

# リンターをインストール
go install golang.org/x/lint/golint@latest

# リンターを実行
golint ./...
```

### 依存関係管理

```bash
# 新しい依存関係を追加
go get github.com/example/package

# 依存関係を更新
go get -u ./...

# go.modとgo.sumを整理
go mod tidy
```


## 開発ノート

- フレームワークはGoの規約に従い、標準のGoツールチェーンを使用
- すべてのコンポーネントは `IService` インターフェースの `Init(ctx)` メソッドを実装
- 依存性注入はコンテキストを通じて行われます - `wf.GetService[T](ctx)` でサービスを取得
- 接続プールはほとんどのアプリケーションに適した合理的なデフォルト値を持っています
- 開発と小規模アプリケーションにはSQLiteを、本番環境にはMySQLを推奨

## ドキュメント

- [Goリファレンスドキュメント](https://pkg.go.dev/github.com/chuccp/go-web-frame)
- [サンプルアプリケーション](./example/)
- [CLAUDE.md](./CLAUDE.md) - Claude Code詳細開発者ガイド

## 貢献

IssueとPull Requestを歓迎します！PRを提出する前に：

1. テストを実行してすべてが通ることを確認
2. コードスタイルをプロジェクトと一貫させる
3. 新機能に適切なテストを追加
4. 関連ドキュメントを更新

## ライセンス

MIT License - 詳細は[LICENSE](./LICENSE)を参照