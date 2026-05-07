# Util API

Utility functions.

## String Utilities

```go
import "github.com/chuccp/go-web-frame/util"

util.IsEmpty(str)
util.TrimSpace(str)
util.Contains(str, substr)
util.Split(str, sep)
util.Join(parts, sep)
```

## Crypto Utilities

```go
util.MD5(str)
util.SHA256(str)
util.Base64Encode(data)
util.Base64Decode(str)
```

## Time Utilities

```go
util.Now()
util.FormatTime(t, layout)
util.ParseTime(str, layout)
util.Timestamp()
```

## Network Utilities

```go
util.GetLocalIP()
util.IsValidIP(ip)
util.IsValidPort(port)
```

## File Utilities

```go
util.FileExists(path)
util.DirExists(path)
util.ReadFile(path)
util.WriteFile(path, data)
util.DeleteFile(path)
```

## UUID

```go
uuid := util.UUID()
```

## Random

```go
randStr := util.RandomString(length)
randInt := util.RandomInt(min, max)
```