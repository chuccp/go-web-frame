# 更新日志

本文档记录 Go Web Frame 的版本更新和历史变更。

## [未发布] - 开发中

### 新增
- 添加了完整的 MkDocs 文档站点
- 为两个示例项目（http2smtp、antilost_qrcode_go）创建了 CODEBUDDY.md 文件
- 修正了文档中的 API 用法错误

### 修正
- 修正了控制器接口名称：`core.IRest` → `core.IService`
- 修正了路由注册方式：`ctx.Group()` → 直接使用 `context.Get()`、`context.Post()` 等
- 修正了配置格式：从 JSON 改为 YAML（实际使用的格式）
- 修正了 Init 方法签名：`Init(ctx)` → `Init(context)`

### 移除
- 移除了对不存在的 RSS 插件的引用

## [1.0.0] - 2026-04-01

### 新增
- 初始版本发布
- 基于 Gin 的 Web 框架
- 支持依赖注入
- 类型安全的泛型 ORM
- 组件系统
- 配置管理（支持 INI、JSON、YAML、TOML）
- 过滤器/中间件支持
- 后台任务（Runner）
- 静态文件服务
- 反向代理
- HTTPS 自动证书（Let's Encrypt）

## 发布历史

| 版本 | 发布日期 | 主要变更 |
|------|----------|----------|
| 1.0.0 | 2026-04-01 | 初始版本 |

## 升级指南

### 从 0.x 升级到 1.0.0

1. **接口变更**：将 `core.IRest` 改为 `core.IService`
2. **路由注册**：使用 `context.Get()`、`context.Post()` 等方法直接注册路由
3. **配置格式**：推荐使用 YAML 格式
4. **Init 方法**：确保方法签名为 `Init(context *core.Context) error`

## 下一步

- [首页](../index.md) - 返回首页
- [快速开始](../getting-started/installation.md) - 重新查看安装指南
