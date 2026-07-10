# Static Files & Proxy

This document covers static file serving and reverse proxy usage in Go Web Frame.

## Static Files

### Static File Directory

Register in a controller's `Init` method:

```go
func (c *MyController) Init(ctx *core.Context) error {
    // Register static file directory: /static/* serves ./www/*
    ctx.Static("/static", "./www")
    return nil
}
```

Visiting `/static/style.css` returns `./www/style.css`.

### Static File System

Use `http.FileSystem` interface to register static files, supporting embedded file systems:

```go
import "net/http"

func (c *MyController) Init(ctx *core.Context) error {
    ctx.StaticFs("/assets", http.Dir("./dist"))
    return nil
}
```

### Register via Configuration

Specify static file directories in the configuration file:

```yaml
web:
  server:
    locations:              # Static file directories
      - ./view/dist         # Vue/React build output
      - ./www               # Static resources
    page404: index.html     # SPA mode: unmatched routes return index.html
```

When `page404` is set to `index.html`, all unmatched routes return that file, enabling SPA frontend routing.

## Reverse Proxy

Register a reverse proxy in a controller to forward requests to a backend service:

```go
func (c *MyController) Init(ctx *core.Context) error {
    // All /api/* requests are proxied to http://backend:8080/api/*
    ctx.ReverseProxy("/api", "http://backend:8080")
    return nil
}
```

All `/api/*` requests are proxied to `http://backend:8080/api/*`.

## Next Steps

- [Routing](../guide/routing.md) - Routing system
- [WebSocket & SSE](websocket-sse.md) - Real-time communication
- [Deployment](deployment.md) - Production deployment (including HTTPS configuration)
