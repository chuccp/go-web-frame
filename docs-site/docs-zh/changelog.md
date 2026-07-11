# 更新日志

本文档记录 Go Web Frame 的版本更新和历史变更。

## [未发布]

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
| 1.0.11 | 2026-07-11 | 自签证书、GetHandler 测试支持、TLS 容错 |
| 1.0.0 | 2026-04-07 | 初始版本 |

## 下一步

- [首页](index.md) - 返回首页
- [快速开始](getting-started/installation.md) - 重新查看安装指南
