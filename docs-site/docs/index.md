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

<div class="grid cards" markdown>

- :material-rocket-launch: **快速开始**
  
    5 分钟上手 Go Web Frame
    
    [:octicons-arrow-right-24: 开始使用](getting-started/installation.md)

- :material-book-open-page-variant: **用户指南**
  
    深入了解框架功能
    
    [:octicons-arrow-right-24: 查看指南](guide/routing.md)

- :material-api: **API 参考**
  
    完整的 API 文档
    
    [:octicons-arrow-right-24: 查看 API](api/core.md)

- :material-lightbulb: **最佳实践**
  
    推荐的使用模式和技巧
    
    [:octicons-arrow-right-24: 了解更多](best-practices.md)

</div>

## 主要特性

- :material-check-circle: **依赖注入** - 基于 Context 的类型安全 DI 容器
- :material-check-circle: **类型安全 ORM** - 零样板代码的泛型 Model
- :material-check-circle: **自动配置** - 多位置自动发现配置文件
- :material-check-circle: **组件系统** - 可复用的独立组件
- :material-check-circle: **守护进程模式** - 支持作为系统服务运行
- :material-check-circle: **HTTPS 自动证书** - 集成 Let's Encrypt

## 社区

- [GitHub](https://github.com/chuccp/go-web-frame)
- [问题反馈](https://github.com/chuccp/go-web-frame/issues)
