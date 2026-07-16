# 更新日志

本文档记录 Go Web Frame 的版本更新和历史变更。

## [未发布]

### 变更
- **CORS 拼写修正**：`NewCrosFilter` 重命名为 `NewCorsFilter`，`crosHandlerFunc` 重命名为 `corsHandlerFunc`。旧名称保留为 deprecated 别名，将在未来版本移除。

## [1.0.14] - 2026-07-16

### 新增
- **WebSocket Accept 选项函数**：为 `AcceptOptions` 的所有字段添加了选项函数，支持流式配置 WebSocket 接受行为（如 `SetInsecureSkipVerify`、`SetOriginPatterns`、`SetCompressionMode`）。
- **WebSocket `AcceptOptions` 字段对齐**：底层 `websocket.AcceptOptions` 的所有字段现在已完整映射到框架的 `AcceptOptions` 结构体，提供完整的配置支持。

### 重构
- **WebSocket 延迟初始化**：WebSocket 连接现在通过 `sync.Once` 在首次读写时延迟初始化，避免不必要的资源分配，并修复了 `OpenStream` 中的连接切换问题。
- **WebSocket 包装器**：引入 `WebSocketStream` 包装器，封装 WebSocket 连接，提供适当的连接生命周期管理和线程安全初始化。
- **组件独立模块**：`component/` 下所有包现均为独立的 Go 模块，各自拥有 `go.mod` 和版本标签 — `auth`、`cache`、`captcha`、`cors`、`localcache`、`qrcode`、`ratelimit`、`schedule`、`validator` 均有 `component/<name>/v1.0.14` 标签。

### 修复
- **`AddHandles` 空指针**：添加 nil 检查，防止路由处理器缺失时发生 panic。
- **WebSocket 数据竞争和上下文泄漏**：修复延迟连接初始化中的竞争条件，以及未取消的后台 goroutine 导致的上下文泄漏。
- **WebSocket 重试循环**：首次 `Accept` 失败后停止无限重试，立即返回错误。
- **`getConn` 非同步快速路径**：移除可能导致 WebSocket 连接并发读写竞争的非同步快速路径。
- **`AcceptOptions` 切片初始化**：`AcceptOptions` 中的切片字段现在正确初始化为空切片（而非 nil）。
- **GORM Model 与 Joins**：使用 `Joins` 时现在也会设置 GORM 的 `Model()` 子句，而不仅限于 `Preloads`，修复了 join 查询中的关联解析问题。
- **`Query.One()` 错误处理**：未找到记录时 `Query.One()` 现在返回零值和 nil 错误（而非包装过的 `gorm.ErrRecordNotFound`），调用方应检查返回值是否为零值来判断记录是否存在。

## [1.0.12] - 2026-07-15

### 新增
- **`MemFileSystem.ExistsFile()`**：判断路径是否作为文件（非目录）存在，单次 `Stat` 调用，用于 SPA 404 回退时排除目录误匹配。
- **`Table.Model()`**：`db.Table` 新增 `Model(value)` 方法，封装 GORM 的 Model，配合 Preload 正确解析关联。
- **`Server.ServerConfig()`**：暴露 Server 的配置对象，方便外部读取。
- **`Server.Listen()` / `Server.ListenTLS()`**：将 HTTP/TLS 监听逻辑从 Runner 移到 Server，便于独立使用。
- **`Server.AddFilters()`**：批量添加过滤器的便捷方法。
- **访问日志**：`optionsMiddleware` 增加 debug 级别访问日志，打印请求方法和路径。
- **`util.DefaultPage()`**：统一分页参数默认值（PageNo=1, PageSize=10），消除各处重复的分页校验代码。

### 重构
- **`core/server.go`**：去掉 `Init()` 阶段，改为 `Run()` 时按需创建 Server 并初始化 RestGroup；新增 `AddIRunner`/`AddRestGroup`。
- **`core/context.go`**：`NewContext` 不再依赖 `*web.Server`，移除 server 字段，根 Context 更纯粹。
- **`core/rest.go`**：`Handles` 通过 `RestGroup` 传递到 Server，`Build()` 直接构造结构体。
- **`web/runner.go`**：监听逻辑委托给 `Server.Listen`/`ListenTLS`，删除重复实现和死代码 `logTLSListen`。
- **`web_frame.go`**：不再预先创建 Server，handles 通过 RestGroup 传递，简化 init 流程。
- **`component/cors/cors.go`**：提取 `crosHandlerFunc`，`Handle` 加 nil 保护。
- **`model/clause.go`**：`ListPage`/`ExecPage`/`Page` 统一使用 `DefaultPage`，消除重复校验。

### 修复
- SPA 404 回退路径检查从 `Exists` 改为 `ExistsFile`，避免目录被误判为有效回退文件。
- `initServer` 中 server 创建失败不再吞 nil，错误统一用 `errors.WithStackIf` 包装。
- Runner `Init` 从 `Run()` 提前到 `web_frame.init()`，确保 runner 不依赖未初始化的 REST 路由。
- `optionsMiddleware` 先于 `justInitRoute` 注册，确保中间件链顺序正确。

## [1.0.11] - 2026-07-11

### 新增
- **自签证书生成**：当没有通配符证书匹配时，框架自动生成内存中的 ECDSA P-256 自签证书（有效期 1 年）并缓存。开发环境无需外部证书。
- **IPv6 TLS 支持**：自签证书正确处理 IPv6 地址，包括 HTTP Host header 中的方括号格式（`[::1]`、`[::1]:8443`）。
- **`Servers.GetHandler()`**：返回 `http.Handler`，根据 Host header 中的端口号分发请求到对应的 Server engine。每个 server 的路由、过滤器和 ContextPath 完全独立。
- **`WebFrame.GetHandler()`**：初始化完整应用（数据库、服务、路由）并返回 `http.Handler`，可直接用于 `httptest.NewServer()`。
- **`core.Server.GetHandler()`**：暴露底层 `Servers.GetHandler()`，用于测试和嵌入场景。
- 多域名 server listening — 所有匹配的域名各自打印监听 URL。

### 修复
- `certStore.init()` 现在收集所有证书错误后统一返回，确保尽可能多的证书被加载。
- 证书加载失败不再阻塞 HTTP 服务器启动 — 错误记录到日志，服务器继续运行（无 TLS）。
- MkDocs 2.0 兼容性修复（移除废弃插件、空 overrides、无效图标）。

### 变更
- `core.NewServer()` 现在接收外部 `*web.Servers` 参数，确保 builder 注册的路由在 `GetHandler()` 中可用。
- 文档改用标准 markdown 以兼容 MkDocs 2.0。

## [1.0.0] - 2026-04-07

### 新增
- 初始版本发布
- **核心功能**
  - 基于 Context 的依赖注入容器
  - MVC 风格架构，职责分离清晰
  - 类型安全的泛型 ORM，零样板代码 CRUD 操作
  - 内置支持 SQLite、MySQL、PostgreSQL、Redis

- **Web 层**
  - REST 控制器，自动路由注册
  - 过滤器/中间件链
  - 路由元数据支持（`.WithMeta()`）
  - 静态文件服务，支持 SPA 回退
  - 反向代理
  - WebSocket 和 Server-Sent Events (SSE)
  - Let's Encrypt 自动 HTTPS 证书

- **数据访问**
  - 泛型 `Model[T]` 提供类型安全的数据库操作
  - `EntryModel[T, PK]` 支持主键实体
  - 链式查询构建器
  - 事务支持
  - 连接池配置

- **组件**
  - 缓存组件（Redis）
  - 本地缓存组件（Otter）
  - 限流组件
  - 验证码生成
  - 二维码生成
  - 定时任务（Cron）
  - 输入校验
  - 认证过滤器
  - CORS 过滤器

- **基础设施**
  - 自动加载配置（JSON/YAML/TOML/INI）
  - 结构化日志（Zap + 日志轮转）
  - Builder 模式构建应用

- **文档**
  - 多语言 README（英文、中文、繁体中文、日文）
  - MkDocs 文档站点（英文 + 中文）
  - 常见用例示例应用

## 发布历史

| 版本 | 发布日期 | 主要变更 |
|------|----------|----------|
| 1.0.14 | 2026-07-16 | WebSocket AcceptOptions、延迟初始化、数据竞争修复、组件独立模块 |
| 1.0.13 | 2026-07-16 | WebSocket AcceptOptions、延迟初始化、数据竞争修复、Query.One() 修复 |
| 1.0.12 | 2026-07-15 | ExistsFile、Model/Preload 关联、核心重构、DefaultPage |
| 1.0.11 | 2026-07-11 | 自签证书、GetHandler 测试支持、TLS 容错 |
| 1.0.0 | 2026-04-07 | 初始版本 |

## 下一步

- [首页](index.md) - 返回首页
- [快速开始](getting-started/installation.md) - 重新查看安装指南
