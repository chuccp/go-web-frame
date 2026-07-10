# 安装

## 环境要求

- **Go 1.18+**（需要泛型支持）
- **Git**

## 安装框架

### 方式一：Go Modules（推荐）

```bash
go get github.com/chuccp/go-web-frame
```

在代码中导入：

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

框架会自动引入以下核心依赖：

| 依赖 | 用途 |
|---|---|
| gin-gonic/gin | HTTP Web 框架 |
| gorm.io/gorm | ORM 库 |
| spf13/viper | 配置管理 |
| go.uber.org/zap | 结构化日志 |
| go-redis/redis | Redis 客户端 |

## 验证安装

创建 `main.go`：

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Get("/", func(req *web.Request) (any, error) {
        return "Hello, Go Web Frame!", nil
    })
    builder.Build().Run(context.Background())
}
```

创建 `application.yml`：

```yaml
web:
  server:
    port: 8081
    host: 0.0.0.0
  db:
    type: sqlite
    path: ./data.db
  log:
    level: debug
    path: ./logs/app.log
```

运行：

```bash
go run main.go
```

访问 `http://localhost:8081`，看到 `"Hello, Go Web Frame!"` 即表示安装成功。

## 下一步

- [5 分钟上手](quick-start.md) - 创建你的第一个应用
