# 快速开始

本指南将帮助你快速上手 Go Web Frame 框架。

## 项目结构

一个典型的 Go Web Frame 应用结构如下：

```
myapp/
├── main.go                 # 应用入口
├── go.mod                  # Go 模块文件
├── go.sum
├── config.ini              # 配置文件（INI 格式）
├── controller/             # HTTP 处理器 / REST 控制器
│   ├── user_controller.go
│   └── order_controller.go
├── service/                # 业务逻辑层
│   ├── user_service.go
│   └── order_service.go
├── model/                  # 数据访问层
│   ├── user.go
│   ├── user_model.go
│   └── order_model.go
├── entity/                 # 领域实体
│   ├── user.go
│   └── order.go
├── filter/                 # HTTP 过滤器 / 中间件
│   ├── auth_filter.go
│   └── logging_filter.go
└── runner/                 # 后台任务
    └── cleanup_runner.go
```

## 创建第一个应用

### 1. 初始化项目

```bash
mkdir myapp
cd myapp
go mod init myapp
go get github.com/chuccp/go-web-frame
```

### 2. 创建配置文件

创建 `config.ini`：

```ini
[core]
init      = true
cachePath = .cache

[sqlite]
filename = data.db

[manage]
port     = 8081
username = admin
password = admin123

[api]
port = 8082
```

### 3. 编写入口文件

创建 `main.go`：

```go
package main

import (
	config2 "github.com/chuccp/go-web-frame/config"
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/web"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	fileConfig, err := config2.LoadSingleFileConfig("config.ini")
	if err != nil {
		zap.L().Fatal("加载配置失败", zap.Error(err))
	}

	// 创建应用
	builder := wf.NewBuilder(fileConfig)

	// 注册路由
	builder.Get("/", func(req *web.Request) (any, error) {
		return "Welcome!", nil
	})

	// 构建应用
	app := builder.Build()

	// 启动应用
	err = app.Start()
	if err != nil {
		zap.L().Fatal("启动应用失败", zap.Error(err))
	}
}
```

### 4. 运行应用

```bash
go run main.go
```

访问 `http://localhost:8081` 看到 `"Welcome!"` 即表示成功。

## 核心概念

### WebFrame 创建

```go
// 方式一：使用配置文件
fileConfig, err := config.LoadSingleFileConfig("config.ini")
builder := wf.NewBuilder(fileConfig)
app := builder.Build()

// 方式二：使用自动配置
app := wf.NewWithAutoConfig()
```

### 路由注册

```go
// 使用 Builder 注册路由
builder.Get("/users", handler)
builder.Post("/users", handler)
builder.Put("/users/:id", handler)
builder.Delete("/users/:id", handler)
```

### 请求处理

处理器函数签名统一为：

```go
func handler(req *web.Request) (any, error)
```

- 返回 `any`：自动转换为 JSON 响应
- 返回 `error`：自动转换为错误响应

## 完整示例

```go
package main

import (
	config2 "github.com/chuccp/go-web-frame/config"
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/web"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	fileConfig, err := config2.LoadSingleFileConfig("config.ini")
	if err != nil {
		zap.L().Fatal("加载配置失败", zap.Error(err))
	}

	// 创建 Builder
	builder := wf.NewBuilder(fileConfig)

	// 基本路由
	builder.Get("/", func(req *web.Request) (any, error) {
		return "Welcome!", nil
	})

	// 路径参数
	builder.Get("/users/:id", func(req *web.Request) (any, error) {
		id := req.Param("id")
		return map[string]any{"id": id}, nil
	})

	// 查询参数
	builder.Get("/search", func(req *web.Request) (any, error) {
		q := req.Query("q")
		return map[string]any{"keyword": q}, nil
	})

	// JSON 请求体
	builder.Post("/users", func(req *web.Request) (any, error) {
		var user struct {
			Name string `json:"name"`
		}
		if err := req.BindJSON(&user); err != nil {
			return nil, err
		}
		return map[string]any{"name": user.Name}, nil
	})

	// 构建应用
	app := builder.Build()

	// 启动应用
	err = app.Start()
	if err != nil {
		zap.L().Fatal("启动应用失败", zap.Error(err))
	}
}
```

## 下一步

- [Hello World](hello-world.md) - 更详细的示例
- [路由](guide/routing.md) - 深入了解路由系统
- [控制器](guide/controller.md) - 使用 REST 控制器组织代码
