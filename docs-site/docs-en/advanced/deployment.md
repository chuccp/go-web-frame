# Deployment

Production deployment guide for Go Web Frame applications.

## Graceful Shutdown

Use `Run(ctx)` with OS signals for graceful shutdown:

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
)

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    app := builder.Build()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()

    app.Run(ctx)
}
```

| Method | Behavior |
|--------|----------|
| `Start()` | Starts and blocks; no graceful shutdown |
| `Run(ctx)` | Starts; gracefully shuts down HTTP server and runners when context is cancelled |

## HTTPS / SSL

### Let's Encrypt Auto-Certificate

Built-in Let's Encrypt integration. `CertManager` automatically obtains and renews certificates:

```yaml
web:
  server:
    port: 443
    ssl:
      enabled: true
      hosts:
        - example.com
        - api.example.com
```

Certificates are auto-generated and cached in `./certs`.

> **Note**: Ports 80 and 443 must be publicly accessible, and the domain must resolve to the server IP.

### Local Certificate Files

If you already have certificate files (e.g., from a CA or self-signed), configure them directly via `certs`:

```yaml
web:
  server:
    port: 443
    ssl:
      enabled: true
      certs:
        - host: example.com
          cert-file: /etc/ssl/example.com/fullchain.pem
          key-file: /etc/ssl/example.com/privkey.pem
        - host: api.example.com
          cert-file: /etc/ssl/api.example.com/fullchain.pem
          key-file: /etc/ssl/api.example.com/privkey.pem
```

Each entry in `certs` maps a host to its certificate and key file. The framework selects the correct certificate based on the TLS SNI (Server Name Indication) of the incoming request.

### Mixed Mode: Local Certificates + Auto-Cert Fallback

You can combine both approaches — use local certificates for some domains and Let's Encrypt for others:

```yaml
web:
  server:
    port: 443
    ssl:
      enabled: true
      hosts:                      # auto-cert for these domains
        - auto.example.com
      certs:                      # local certs for these domains
        - host: example.com
          cert-file: /etc/ssl/example.com/fullchain.pem
          key-file: /etc/ssl/example.com/privkey.pem
```

Certificate selection priority: local cert match > autocert > first local cert as default.

### Running HTTP and HTTPS Simultaneously

Register two `RestGroup`s on different ports:

```go
httpsGroup := core.NewRestGroupBuilder().
    ServerConfig(sslServerConfig).
    Rest(&UserController{}).
    Build()

httpGroup := core.NewRestGroupBuilder().
    ServerConfig(web.DefaultServerConfig()). // port 80
    Rest(&UserController{}).
    Build()

builder.RestGroup(httpsGroup, httpGroup)
```

## Server Configuration

### Full Configuration Reference

```yaml
web:
  server:
    port: 8080
    context_path: /api
    locations:
      - ./view/dist
      - www
    page404: 404.html
    ssl:
      enabled: false
      hosts: []
      certs: []
```

| Key | Description | Default |
|-----|-------------|----------|
| `web.server.port` | Listen port | 19009 |
| `web.server.context_path` | Route prefix | none |
| `web.server.locations` | Static file directories | none |
| `web.server.page404` | 404 fallback page | none |
| `web.ssl.enabled` | Enable HTTPS | false |
| `web.ssl.hosts` | Domains for Let's Encrypt | none |
| `web.ssl.certs` | Local certificate entries | none |

## Production Checklist

- [ ] Use MySQL/PostgreSQL instead of SQLite
- [ ] Configure connection pool (`max_open_conns`, `max_idle_conns`)
- [ ] Enable HTTPS (Let's Encrypt or local certificates)
- [ ] Set log level to `info` or `warn`
- [ ] Configure rate limiting
- [ ] Set up health check endpoint
- [ ] Use `Run(ctx)` with signal handling for graceful shutdown