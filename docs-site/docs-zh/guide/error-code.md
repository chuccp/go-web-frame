# 错误码

`web/error_code.go` 提供了一套结构化的错误类型 `ErrorCode`，用于在 Handler 中返回带业务码的 HTTP 错误响应。

## ErrorCode 结构

```go
type ErrorCode struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
    Err     error  `json:"-"`
}
```

| 字段 | 说明 |
|---|---|
| `Code` | 业务/HTTP 错误码，会序列化到响应 JSON |
| `Message` | 错误摘要 |
| `Detail` | 可选的详细说明 |
| `Err` | 包装的原始错误，不参与序列化，但支持 `errors.Is` / `errors.As` |

## 错误码常量

| 常量 | 值 | 说明 |
|---|---|---|
| `CodeOK` | 200 | 成功 |
| `CodeBadRequest` | 400 | 请求参数错误 |
| `CodeUnauthorized` | 401 | 未授权 |
| `CodeForbidden` | 403 | 禁止访问 |
| `CodeNotFound` | 404 | 资源不存在 |
| `CodeMethodNotAllowed` | 405 | 方法不允许 |
| `CodeTooManyRequests` | 429 | 请求过于频繁 |
| `CodeInternalError` | 500 | 服务器内部错误 |
| `CodeServiceUnavailable` | 503 | 服务不可用 |
| `CodeValidationFailed` | 1001 | 校验失败 |
| `CodeDuplicateEntry` | 1002 | 重复记录 |
| `CodeTokenExpired` | 1003 | Token 已过期 |
| `CodeTokenInvalid` | 1004 | Token 无效 |

## 构造错误

### 预定义构造函数

每个预定义错误都有对应的构造函数，返回一个可修改的副本：

```go
return nil, web.NewBadRequest().WithDetail("参数错误")          // 400
return nil, web.NewUnauthorized().WithError(err)              // 401
return nil, web.NewForbidden().WithDetail("无权限")            // 403
return nil, web.NewNotFound().WithDetail("user not found")    // 404
return nil, web.NewMethodNotAllowed().WithDetail("不支持的方法") // 405
return nil, web.NewTooManyRequests().WithDetail("请稍后再试")   // 429
return nil, web.NewInternalError().WithError(err)             // 500
return nil, web.NewServiceUnavailable().WithDetail("服务暂不可用") // 503

// 业务错误
return nil, web.NewValidationError().WithError(err)           // 1001
return nil, web.NewDuplicateEntry().WithDetail("邮箱已存在")    // 1002
return nil, web.NewTokenExpired().WithDetail("token 过期")     // 1003
return nil, web.NewTokenInvalid().WithDetail("token 无效")     // 1004
```

### 自定义错误码

```go
return nil, web.NewErrorCode(2001, "custom error").WithDetail("自定义错误详情")
```

## 链式方法

```go
err := web.NewNotFound().
    WithDetail("user #123 not found").
    WithError(sqlErr)
```

- `WithDetail(detail string) *ErrorCode`：附加详细说明
- `WithError(err error) *ErrorCode`：包装原始错误

## 在 Handler 中使用

```go
func (c *UserController) Get(req *web.Request) (any, error) {
    user, err := c.userService.GetById(req.ParamUint("id"))
    if err != nil {
        return nil, web.NewInternalError().WithError(err)
    }
    if user == nil {
        return nil, web.NewNotFound().WithDetail("user not found")
    }
    return user, nil
}
```

## 错误映射

`web.ClassifyError(value, err)` 用于把 error 映射为 HTTP 状态码和消息。DefaultConverter 内部就是用这个方法处理错误的：

1. 如果 error 链中包含 `*ErrorCode`，使用其 `Code`。
2. 如果返回值是 `*web.Message` 且 `Code != 200`，使用 `Message.Code`。
3. `os.ErrNotExist` → 404；`os.ErrPermission` → 403。
4. 其他 error → 500。

你也可以在自己的 Converter 或 Filter 中直接调用：

```go
code, msg := web.ClassifyError(nil, err)
req.Response().AbortWithStatusJSON(code, &web.Message{Code: code, Msg: msg})
```

## 在 Filter 中统一处理错误

```go
func (f *ErrorHandlerFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    result, err := fc.Next()
    if err == nil {
        return result, nil
    }

    var ec *web.ErrorCode
    if errors.As(err, &ec) {
        req.Response().AbortWithStatusJSON(ec.Code, ec)
        return nil, nil
    }

    req.Response().AbortWithStatusJSON(500, web.NewInternalError().WithError(err))
    return nil, nil
}
```

## 完整示例

```go
package main

import (
    "context"
    "errors"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type UserController struct{ core.IService }

func (c *UserController) Init(ctx *core.Context) error {
    ctx.Get("/users/:id", c.Get)
    return nil
}

func (c *UserController) Get(req *web.Request) (any, error) {
    if req.ParamUint("id") == 0 {
        return nil, web.NewBadRequest().WithDetail("id is required")
    }
    return map[string]any{"id": req.ParamUint("id"), "name": "alice"}, nil
}

type AuthFilter struct{ core.IFilter }

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if req.GetHeader("Authorization") == "" {
        return nil, web.NewUnauthorized().WithDetail("missing token")
    }
    return fc.Next()
}

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Filter(&AuthFilter{})
    builder.Rest(&UserController{})
    builder.Build().Run(context.Background())
}
```

## 下一步

- [响应转换器](converter.md) - Converter 如何处理 ErrorCode
- [过滤器/中间件](filter.md) - 在 Filter 中统一处理错误
- [Web API 参考](../api/web.md) - web.Message 与错误响应
