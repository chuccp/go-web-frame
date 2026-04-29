# 工具库 API 参考

框架的 `util` 包提供丰富的工具函数，涵盖字符串、加密、文件、网络、时间等常见操作。

## 导入

```go
import "github.com/chuccp/go-web-frame/util"
```

## 字符串工具（str）

```go
// MD5 哈希
hash := util.MD5([]byte("hello"))           // "5d41402abc4b2a76b9719d911017c592"
hash := util.MD5Str("hello")                // 同上，接受字符串

// 随机字符串
s := util.GenerateRandomString(16, "abcdef")           // 指定字符集
n := util.GenerateRandomNum(6)                         // 随机数字字符串
a := util.GenerateRandomStringByAlphanumeric(8)        // 字母+数字

// 空值判断
util.IsBlank("")        // true
util.IsNotBlank("abc")  // true

// 包含判断
util.ContainsAny("hello", "ell", "xyz")               // true（包含任一）
util.ContainsAnyIgnoreCase("Hello", "HELLO")           // true
util.EqualsAnyIgnoreCase("Go", "go", "java")           // true
util.StartsWithAny("https://x.com", "http", "ftp")     // true

// 去重
ids := util.DeduplicateIds("1,2,3,2,1")                // "1,2,3"
unique := util.RemoveDuplicates([]string{"a","b","a"}) // ["a","b"]

// 类型转换
util.StringToInt("123")       // 123
util.StringToUInt("456")      // uint(456)
util.StringToBool("true")     // true
util.BoolToString(true)       // "true"
util.IsNumber("123")          // true

// 文本相似度（0-100）
score := util.TextSimilarity("hello", "hallo")  // ~80

// 截断
util.SubStringMaxLength("hello world", 5)       // "hello"
util.SubStringLastMaxLength("hello world", 5)   // "world"

// 其他
util.Trim("  hello  ")                          // "hello"
util.RemovePunctuation("hello, world!")          // "hello world"
```

## 加密工具（aes）

```go
// AES-256-CBC 加密/解密
encrypted, err := util.EncryptByCBC("plaintext", "0123456789abcdef", "abcdef0123456789")
decrypted, err := util.DecryptByCBC(encrypted, "0123456789abcdef", "abcdef0123456789")

// 字节/Base64 解密
bytes, err := util.DecryptCBCBytes(key, iv, ciphertext)
bytes, err := util.DecryptCBCBase64(key, iv, base64CipherText)

// PKCS7 去填充
data, err := util.PKCS7Unpad(paddedData)
```

## 哈希工具（hash）

```go
sha1 := util.SHA1("hello")          // SHA1 哈希（字符串）
sha1 := util.SHA1Bytes([]byte{...})  // SHA1 哈希（字节）
```

## 文件操作（file）

```go
// 读写文件
content, err := util.ReadFile("config.txt")           // 读取为字符串
bytes, err := util.ReadFileBytes("image.png")          // 读取为字节
err := util.WriteFile([]byte("hello"), "output.txt")   // 写入字节
err := util.WriteBase64File(base64Str, "output.png")   // 写入 Base64

// 目录操作
err := util.CreateDirIfNoExists("./data/cache")
err := util.CreateFileIfNoExists("./data/app.lock")
exists := util.ExistsFile("./data/app.lock")

// File 结构体（链式操作）
f, err := util.NewFile("./data/report.pdf")
abs := f.Abs()               // 绝对路径
name := f.Name()             // 文件名
parent := f.Parent()         // 父目录路径
isDir := f.IsDir()           // 是否目录
exists, _ := f.Exists()      // 是否存在
modTime, _ := f.ModTime()    // 修改时间
children, _ := f.List()      // 列出子文件
child, _ := f.Child("sub")   // 获取子文件
err := f.MkDirs()            // 创建目录
```

## 网络工具（net）

```go
// URL 解析
domain := util.GetDomainFromURL("https://www.example.com/path")  // "www.example.com"
host := util.GetHost("https://www.example.com:8080")             // "www.example.com:8080"
domain := util.GetDomainFromHost("www.example.com:8080")         // "www.example.com"

// 域名验证
util.IsDomain("example.com")   // true
util.IsDomain("not valid")     // false
```

## URL 工具（url）

```go
// URL 操作
url := util.AddQueryParam("/api/users", "page", "1")       // "/api/users?page=1"
url := util.AddQueryParamInt("/api/users", "limit", 10)    // "/api/users?limit=10"
url := util.AddQueryParamFlag("/api/users", "active")      // "/api/users?active"
url := util.JoinUrl("/api", "v1", "users")                  // "/api/v1/users"

// Base64 URL 解码
decoded, err := util.DecodeBase64URL("aHR0cHM6Ly9leGFtcGxlLmNvbQ==")
```

## 时间工具（time）

```go
// 当前时间
ms := util.Millisecond()               // 毫秒时间戳（uint32）
sec := util.Second()                    // 秒时间戳（int64）
now := util.NowDateTime()               // "2024-01-15 10:30:00"
t := util.GetNowTime()                  // time.Time

// 格式化
str := util.FormatTime(t)               // "2024-01-15 10:30:00"
str := util.FormatDate(t)               // "2024-01-15"

// 日期计算
days := util.DaysBetween(t1, t2)        // 两个日期之间的天数

// 时间比较
isAfter, err := util.IsAfterTime("2024-01-15 10:30:00", time.Now())
```

## 手机号工具（phone）

```go
// 手机号脱敏
masked := util.HidePhone("13812345678")  // "138****5678"
```

## 邮箱工具（mail）

```go
// 邮箱格式化
formatted := util.FormatMail("张三", "zhangsan@example.com")  // "张三 <zhangsan@example.com>"

// 邮箱解析
name, email, err := util.ParseMail("张三 <zhangsan@example.com>")

// 提取文本中的邮箱
emails := util.ExtractEmails("联系 admin@test.com 或 support@test.com")
```

## QR 码工具（qrcode）

```go
// 生成 QR 码到文件
file, _ := os.Create("qrcode.png")
err := util.GenerateQrcode("https://example.com", file)

// 生成 QR 码到缓冲区
buffer := util.CreateBufferWriteCloser()
err := util.GenerateQrcode("https://example.com", buffer, util.WithRoundedSquareShape())
qrBytes := buffer.Bytes()
```

## 环境变量（env）

```go
host := util.GetEnvOrDefault("APP_HOST", "localhost")
port := util.GetEnvIntOrDefault("APP_PORT", 8080)
debug := util.GetEnvBoolOrDefault("APP_DEBUG", false)
```

## 反射工具（reflect）

```go
// 类型信息
name := util.GetStructName(user)                  // "User"
fullName := util.GetStructFullName(user)          // "entity.User"
pkgPath := util.GetStructPkgPath(user)            // "myapp/entity"
qualified := util.GetStructFullQualifiedName(user) // "myapp/entity.User"

// 创建指针/切片
ptr := util.NewPtr(user)        // *User
slice := util.NewSlice(user)    // []*User
```

## Map 工具（map）

```go
m := util.GetMap("name", "Alice")                     // map[string]any{"name": "Alice"}
m := util.OfMap("name", "Alice")                      // 同上
m := util.OfMap2("name", "Alice", "age", 30)          // 两个键值对
```

## CRC 校验（crc）

```go
code := util.CRC(6, "order-12345")  // 生成 6 位 CRC 校验码
```

## 正则工具（regx）

```go
matched := util.IsMatch("hello123", `^\w+\d+$`)  // true
```

## JSON 工具（json）

```go
jsonStr, err := util.JsonEncode(data)  // 编码为 JSON 字符串
```

## 模板工具（template）

```go
result, err := util.ParseTemplate("Hello {{.name}}!", map[string]interface{}{"name": "World"})
// result: "Hello World!"
```

## 下一步

- [组件](../guide/components.md) - 框架内置组件
- [核心 API](core.md) - 核心 API 参考
