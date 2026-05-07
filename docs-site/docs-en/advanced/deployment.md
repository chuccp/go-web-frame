# Deployment

Production deployment guide.

## HTTPS with Let's Encrypt

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

## Graceful Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())

// Handle shutdown signal
go func() {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    cancel()
}()

app.Run(ctx)
```

## Daemon Mode

```go
// Run as system service
app.Start()

// Stop service
app.Stop()
```

## Production Checklist

- [ ] Use MySQL/PostgreSQL instead of SQLite
- [ ] Configure connection pool
- [ ] Enable HTTPS
- [ ] Set log level to `info` or `warn`
- [ ] Configure rate limiting
- [ ] Set up health check endpoint