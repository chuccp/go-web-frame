# Util API

The framework's `util` package provides a rich set of utility functions covering strings, crypto, files, networking, time, and more.

## Import

```go
import "github.com/chuccp/go-web-frame/util"
```

## String Utilities (str)

```go
// MD5 hash
hash := util.MD5([]byte("hello"))           // "5d41402abc4b2a76b9719d911017c592"
hash := util.MD5Str("hello")                // same, accepts string

// Random strings
s := util.GenerateRandomString(16, "abcdef")           // custom charset
n := util.GenerateRandomNum(6)                         // random numeric string
a := util.GenerateRandomStringByAlphanumeric(8)        // alphanumeric

// Blank checks
util.IsBlank("")        // true
util.IsNotBlank("abc")  // true

// Contains checks
util.ContainsAny("hello", "ell", "xyz")               // true (contains any)
util.ContainsAnyIgnoreCase("Hello", "HELLO")           // true
util.EqualsAnyIgnoreCase("Go", "go", "java")           // true
util.StartsWithAny("https://x.com", "http", "ftp")     // true

// Deduplication
ids := util.DeduplicateIds("1,2,3,2,1")                // "1,2,3"
unique := util.RemoveDuplicates([]string{"a","b","a"}) // ["a","b"]

// Type conversion
util.StringToInt("123")       // 123
util.StringToUInt("456")      // uint(456)
util.StringToBool("true")     // true
util.BoolToString(true)       // "true"
util.IsNumber("123")          // true

// Text similarity (0-100)
score := util.TextSimilarity("hello", "hallo")  // ~80

// Truncation
util.SubStringMaxLength("hello world", 5)       // "hello"
util.SubStringLastMaxLength("hello world", 5)   // "world"

// Other
util.Trim("  hello  ")                          // "hello"
util.RemovePunctuation("hello, world!")          // "hello world"
```

## Crypto (aes)

```go
// AES-256-CBC encrypt/decrypt
encrypted, err := util.EncryptByCBC("plaintext", "0123456789abcdef", "abcdef0123456789")
decrypted, err := util.DecryptByCBC(encrypted, "0123456789abcdef", "abcdef0123456789")

// Byte/Base64 decrypt
bytes, err := util.DecryptCBCBytes(key, iv, ciphertext)
bytes, err := util.DecryptCBCBase64(key, iv, base64CipherText)

// PKCS7 unpad
data, err := util.PKCS7Unpad(paddedData)
```

## Hash (hash)

```go
sha1 := util.SHA1("hello")          // SHA1 hash (string)
sha1 := util.SHA1Bytes([]byte{...})  // SHA1 hash (bytes)
```

## File Operations (file)

```go
// Read/write files
content, err := util.ReadFile("config.txt")           // read as string
bytes, err := util.ReadFileBytes("image.png")          // read as bytes
err := util.WriteFile([]byte("hello"), "output.txt")   // write bytes
err := util.WriteBase64File(base64Str, "output.png")   // write Base64

// Directory operations
err := util.CreateDirIfNoExists("./data/cache")
err := util.CreateFileIfNoExists("./data/app.lock")
exists := util.ExistsFile("./data/app.lock")

// File struct (chainable)
f, err := util.NewFile("./data/report.pdf")
abs := f.Abs()               // absolute path
name := f.Name()             // filename
parent := f.Parent()         // parent directory
isDir := f.IsDir()           // is directory
exists, _ := f.Exists()      // exists
modTime, _ := f.ModTime()    // modification time
children, _ := f.List()      // list children
child, _ := f.Child("sub")   // get child
err := f.MkDirs()            // create directories
```

## Network (net)

```go
// URL parsing
domain := util.GetDomainFromURL("https://www.example.com/path")  // "www.example.com"
host := util.GetHost("https://www.example.com:8080")             // "www.example.com:8080"
domain := util.GetDomainFromHost("www.example.com:8080")         // "www.example.com"

// Domain validation
util.IsDomain("example.com")   // true
util.IsDomain("not valid")     // false
```

## URL (url)

```go
// URL operations
url := util.AddQueryParam("/api/users", "page", "1")       // "/api/users?page=1"
url := util.AddQueryParamInt("/api/users", "limit", 10)    // "/api/users?limit=10"
url := util.AddQueryParamFlag("/api/users", "active")      // "/api/users?active"
url := util.JoinUrl("/api", "v1", "users")                  // "/api/v1/users"

// Base64 URL decode
decoded, err := util.DecodeBase64URL("aHR0cHM6Ly9leGFtcGxlLmNvbQ==")
```

## Time (time)

```go
// Current time
ms := util.Millisecond()               // millisecond timestamp (uint32)
sec := util.Second()                    // second timestamp (int64)
now := util.NowDateTime()               // "2024-01-15 10:30:00"
t := util.GetNowTime()                  // time.Time

// Formatting
str := util.FormatTime(t)               // "2024-01-15 10:30:00"
str := util.FormatDate(t)               // "2024-01-15"

// Date calculation
days := util.DaysBetween(t1, t2)        // days between two dates

// Time comparison
isAfter, err := util.IsAfterTime("2024-01-15 10:30:00", time.Now())
```

## Phone (phone)

```go
// Mask phone number
masked := util.HidePhone("13812345678")  // "138****5678"
```

## Email (mail)

```go
// Format email
formatted := util.FormatMail("John", "john@example.com")  // "John <john@example.com>"

// Parse email
name, email, err := util.ParseMail("John <john@example.com>")

// Extract emails from text
emails := util.ExtractEmails("Contact admin@test.com or support@test.com")
```

## QR Code (qrcode)

```go
// Generate QR code to file
file, _ := os.Create("qrcode.png")
err := util.GenerateQrcode("https://example.com", file)

// Generate QR code to buffer
buffer := util.CreateBufferWriteCloser()
err := util.GenerateQrcode("https://example.com", buffer, util.WithRoundedSquareShape())
qrBytes := buffer.Bytes()
```

## Environment Variables (env)

```go
host := util.GetEnvOrDefault("APP_HOST", "localhost")
port := util.GetEnvIntOrDefault("APP_PORT", 8080)
debug := util.GetEnvBoolOrDefault("APP_DEBUG", false)
```

## Reflection (reflect)

```go
// Type info
name := util.GetStructName(user)                  // "User"
fullName := util.GetStructFullName(user)          // "entity.User"
pkgPath := util.GetStructPkgPath(user)            // "myapp/entity"
qualified := util.GetStructFullQualifiedName(user) // "myapp/entity.User"

// Create pointer/slice
ptr := util.NewPtr(user)        // *User
slice := util.NewSlice(user)    // []*User
```

## Map (map)

```go
m := util.GetMap("name", "Alice")                     // map[string]any{"name": "Alice"}
m := util.OfMap("name", "Alice")                      // same
m := util.OfMap2("name", "Alice", "age", 30)          // two key-value pairs
```

## CRC (crc)

```go
code := util.CRC(6, "order-12345")  // generate 6-digit CRC code
```

## Regex (regx)

```go
matched := util.IsMatch("hello123", `^\w+\d+$`)  // true
```

## JSON (json)

```go
jsonStr, err := util.JsonEncode(data)  // encode to JSON string
```

## Template (template)

```go
result, err := util.ParseTemplate("Hello {{.name}}!", map[string]interface{}{"name": "World"})
// result: "Hello World!"
```

## Next Steps

- [Components](../guide/components.md) - Built-in framework components
- [Core API](core.md) - Core API reference
