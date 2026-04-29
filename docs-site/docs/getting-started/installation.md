# 安装

## 环境要求

- **Go 1.18+** （需要泛型支持）
- **Git**

## 安装框架

### 方式一：Go Modules（推荐）

在你的 Go 项目中，使用 `go get` 安装：

```bash
go get github.com/chuccp/go-web-frame
```

然后在代码中导入：

```go
import (
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/web"
)
```

### 方式二：克隆仓库

```bash
git clone https://github.com/chuccp/go-web-frame.git
cd go-web-frame
go mod download
```

## 依赖说明

框架会自动引入以下依赖：

| 依赖 | 用途 |
|------|------|
| gin-gonic/gin | HTTP Web 框架 |
| gorm.io/gorm | ORM 库 |
| spf13/viper | 配置管理 |
| go.uber.org/zap | 结构化日志 |
| go-redis/redis | Redis 客户端 |

## 验证安装

创建 `main.go` 文件：

```go
package main

import (
	"github.com/chuccp/go-web-frame/config"
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/web"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.LoadSingleFileConfig("application.yml")
	if err != nil {
		zap.L().Fatal("加载配置失败", zap.Error(err))
	}

	// 创建应用
	builder := wf.NewBuilder(cfg)
	builder.Get("/", func(req *web.Request) (any, error) {
		return "Hello, Go Web Frame!", nil
	})

	app := builder.Build()

	// 启动应用
	err = app.Start()
	if err != nil {
		zap.L().Fatal("启动应用失败", zap.Error(err))
	}
}
```

创建 `application.yml` 配置文件：

```yaml
# 服务器配置
web:
  server:
    port: 8081
    host: 0.0.0.0
  # 数据库配置
  db:
    type: sqlite
    path: ./data.db

# 日志配置
log:
  level: debug
  path: ./logs/app.log
```

运行：

```bash
go run main.go
```

访问 `http://localhost:8081` 看到 `"Hello, Go Web Frame!"` 即表示安装成功。

## 下一步

- [快速开始](quick-start.md) - 创建你的第一个应用
- [Hello World](hello-world.md) - 最简单的示例
