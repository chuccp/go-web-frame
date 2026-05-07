# Quick Start

This guide helps you create your first Go Web Frame application in minutes.

## Create Project

```bash
mkdir myapp && cd myapp
go mod init myapp
go get github.com/chuccp/go-web-frame
```

## Create main.go

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    
    builder.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })
    
    builder.Get("/users/:id", func(c *web.Request) (any, error) {
        id := c.Param("id")
        return map[string]any{"id": id, "name": "alice"}, nil
    })
    
    app := builder.Build()
    app.Run(context.Background())
}
```

## Run

```bash
go run main.go
```

Server starts at `http://localhost:19009`.

## Test

```bash
curl http://localhost:19009/
curl http://localhost:19009/users/123
```

## Configuration

Create `config/config.json`:

```json
{
  "web": {
    "server": {
      "port": 8080
    }
  }
}
```

## Next Steps

- [Hello World](hello-world.md) - Detailed hello world example
- [Routing](../guide/routing.md) - Learn about routing
- [Controller](../guide/controller.md) - REST controllers