# Components

Built-in framework components.

## Available Components

| Component | Description |
|-----------|-------------|
| Cache | Redis caching |
| LocalCache | In-memory cache (Otter) |
| RateLimit | Token bucket rate limiting |
| Captcha | Behavioral captcha |
| QRCode | QR code generation |
| Cron | Scheduled tasks |
| Validate | Input validation |
| CORS | Cross-origin support |

## Rate Limiting

```yaml
rate_limit:
  limit: 600      # token fill interval (seconds)
  burst: 5        # bucket capacity
```

## CORS

```go
builder.Filter(&cors.Filter{})
```

## Cache

```go
cache := component.NewCache(redisClient)
cache.Set("key", "value", time.Hour)
val := cache.Get("key")
```

## Local Cache

```yaml
local_cache:
  path: ./cache
  open: true
```

## Validation

```go
import "github.com/chuccp/go-web-frame/component"

validate := component.NewValidate()
err := validate.Struct(user)
```