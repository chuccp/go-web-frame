# 更新日志

本文档记录 Go Web Frame 的版本更新和历史变更。

## [未发布] - 开发中

### 新增
- **SSL 本地证书**：新增 `SSLCert` 结构体和 `SSLConfig.Certs` 字段，支持配置本地证书文件。支持多域名本地证书、SNI 动态匹配和 autocert 兜底。
- **Context 传播**：`Model[T]`、`EntryModel[T, PK]`、`Query[T]`、`Update[T]`、`Delete[T]` 新增 `WithContext(ctx)` 方法，支持请求级取消、超时和链路追踪传播到数据库操作。
- `web.Request` 新增 `Ctx()` 方法，暴露每个请求的 `context.Context`。
- `db` 包新增 `DB.WithContext(ctx)` 和 `Table.WithContext(ctx)` 作为基础层。
- 添加了完整的 MkDocs 文档站点
- 修正了文档中的 API 用法错误

### 修复
- `Builder.Any()` 现在正确注册所有 HTTP 方法（之前只注册了 GET）。
- `Config.GetBoolOrDefault()` 在 key 未配置时正确返回默认值。
- 重命名 `model.Model[T]` 内部方法 `getBb()` 为 `getDB()`。

### 变更
- ORM 对比文档更新——Context 传播现在是相比 GORM 内置泛型的关键差异化优势。

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
| 1.0.0 | 2026-04-07 | 初始版本 |

## 下一步

- [首页](index.md) - 返回首页
- [快速开始](getting-started/installation.md) - 重新查看安装指南
