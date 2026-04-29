# Go Web Frame 使用手册

欢迎使用 Go Web Frame —— 一个现代化、功能丰富的 Go Web 框架。

## 什么是 Go Web Frame？

Go Web Frame 是一个基于 Gin 的 Go 语言 Web 框架，提供结构化的 Web 应用开发方式，内置依赖注入、类型安全的 ORM 和数据库集成。

**核心优势：** 框架集成了 Go 生态中最佳的开源包和框架，以最佳实践有机组合，是快速构建生产级应用的最优方案。

## 技术栈

### 核心框架
| 组件 | 说明 |
|------|------|
| [Gin](https://github.com/gin-gonic/gin) | 高性能 HTTP Web 框架 |
| [GORM](https://gorm.io/) | 强大的 ORM 库，支持多数据库 |
| [Viper](https://github.com/spf13/viper) | 完整的配置解决方案 |
| [Zap](https://go.uber.org/zap) | Uber 的高性能结构化日志库 |

### 数据存储
| 组件 | 说明 |
|------|------|
| [go-redis](https://github.com/redis/go-redis) | Redis 官方推荐客户端 |
| [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | 纯 Go SQLite 实现，无 CGO 依赖 |

## 快速链接

### Getting Started

- [安装](getting-started/installation.md) - 环境要求和安装方式
- [快速开始](getting-started/quick-start.md) - 创建第一个应用
- [Hello World](getting-started/hello-world.md) - 最简单的示例

### 用户指南

- [路由](guide/routing.md) - HTTP 路由系统（路径参数、查询参数、REST 控制器、静态文件、反向代理、WebSocket、SSE）
- [控制器](guide/controller.md) - REST 控制器和请求处理
- [服务](guide/service.md) - 业务逻辑层和依赖注入
- [模型](guide/model.md) - 类型安全的 ORM（Model、EntryModel、查询构建器、事务）
- [过滤器/中间件](guide/filter.md) - HTTP 请求过滤（认证、日志、CORS、限流、路由元数据）
- [配置](guide/configuration.md) - 配置管理（YAML/JSON/TOML、环境变量、多环境）
- [后台任务](guide/runner.md) - Runner 和定时任务调度
- [组件](guide/components.md) - 框架内置组件（限流、认证、定时任务、缓存、验证码等）

### 高级主题

- [数据库高级用法](advanced/database.md) - 事务、模型组、迁移、原生 SQL

### API 参考

- [核心 API](api/core.md) - WebFrame、Builder、Context、Request、Response
- [Web API](api/web.md) - HandlerFunc、路由注册、响应类型、web.Message
- [模型 API](api/model.md) - Model、EntryModel、Query、Transaction

### 最佳实践

- [最佳实践](best-practices.md) - 推荐的项目结构、分层分离、错误处理、认证、测试

## 主要特性

- **依赖注入** - 基于 Context 的类型安全 DI 容器
- **类型安全 ORM** - 零样板代码的泛型 Model
- **灵活配置** - 多位置、多格式配置文件（YAML/JSON/TOML）
- **组件系统** - 可复用的独立组件（限流、认证、定时任务等）
- **HTTPS 自动证书** - 集成 Let's Encrypt

## 社区

- [GitHub](https://github.com/chuccp/go-web-frame)
- [问题反馈](https://github.com/chuccp/go-web-frame/issues)
