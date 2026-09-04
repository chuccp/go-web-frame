# Value 包 — 类型安全的动态值

`value` 包提供了一种类型安全的方式来处理动态 JSON 类似的数据结构。

## 概述

`value` 包包含以下核心类型：

| 类型 | 说明 | 用途 |
|------|------|------|
| `Object` | 键值对集合 | 类似 JSON 对象 |
| `Array` | 值的有序集合 | 类似 JSON 数组 |
| `Text` | 字符串值 | 文本数据 |
| `Number` | 数字值 | 整数或浮点数 |
| `Bool` | 布尔值 | true/false |
| `Any` | 任意值 | 非标准类型（结构体等） |
| `Null` | 空值 | null |

## 基本用法

### 创建 Object

```go
import "github.com/chuccp/go-web-frame/value"

// 创建空对象
obj := value.NewObject()

// 从 map 创建
m := map[string]any{"name": "张三", "age": 25}
obj := value.NewObjectFromMap(m)

// 从 JSON 创建
jsonBytes := []byte(`{"name":"张三","age":25}`)
obj, err := value.NewObjectFromJson(jsonBytes)
```

### 设置和获取值

```go
// 设置值
obj.PutAny("name", "张三")      // 自动转换类型
obj.PutAny("age", 25)
obj.PutAny("active", true)
obj.Put("key", value.NewText("value"))  // 使用具体类型

// 获取值
name := obj.GetString("name")           // "张三"
age := obj.GetInt("age")                // 25
active := obj.GetBool("active")         // true
score := obj.GetNumber("score")         // 0.0 (不存在时返回零值)

// 带默认值的获取
count := obj.GetIntForDefault("count", 0)        // 0 (默认值)
name := obj.GetStringForDefault("name", "匿名")   // "匿名" (默认值)

// 检查键是否存在
if obj.HasKey("name") {
    // 键存在
}

// 检查键值对
if obj.HasKeyValue("name", "张三") {
    // 键存在且值等于 "张三"
}
```

### 嵌套对象和数组

```go
// 创建嵌套对象
address := value.NewObject()
address.PutAny("city", "北京")
address.PutAny("street", "朝阳路")
obj.Put("address", address)

// 获取嵌套对象
addr := obj.GetObject("address")
city := addr.GetString("city")  // "北京"

// 创建数组
tags := value.NewArray(
    value.NewText("go"),
    value.NewText("web"),
    value.NewText("framework"),
)
obj.Put("tags", tags)

// 获取数组
arr := obj.GetArray("tags")
arr.ForEach(func(i int, v value.Value) bool {
    fmt.Println(v.String())  // go, web, framework
    return true
})
```

## 高级用法

### 解码到结构体

使用 `Decode` 方法将 Object 转换为 Go 结构体：

```go
type User struct {
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Email   string `json:"email"`
    Address struct {
        City    string `json:"city"`
        ZipCode string `json:"zip_code"`
    } `json:"address"`
}

obj := value.NewObject()
obj.PutAny("name", "张三")
obj.PutAny("age", 25)
obj.PutAny("email", "zhangsan@example.com")

address := value.NewObject()
address.PutAny("city", "北京")
address.PutAny("zip_code", "100000")
obj.Put("address", address)

var user User
err := obj.Decode(&user)
if err != nil {
    log.Fatal(err)
}

fmt.Println(user.Name)           // 张三
fmt.Println(user.Address.City)   // 北京
```

### 转换为 Map

```go
obj := value.NewObject()
obj.PutAny("name", "张三")
obj.PutAny("age", 25)

m := obj.ToMap()
// m = map[string]any{"name": "张三", "age": int64(25)}
```

### JSON 序列化

```go
obj := value.NewObject()
obj.PutAny("name", "张三")
obj.PutAny("age", 25)

// 转换为 JSON
jsonBytes := obj.ToJSON()
fmt.Println(string(jsonBytes))  // {"name":"张三","age":25}

// 实现了 json.Marshaler 接口
data, err := json.Marshal(obj)

// 实现了 json.Unmarshaler 接口
var obj2 value.Object
err := json.Unmarshal(data, &obj2)
```

### Any 类型

`Any` 类型用于处理非标准值（如自定义结构体）：

```go
type Status struct {
    Code    int
    Message string
}

obj := value.NewObject()
obj.PutAny("status", Status{Code: 200, Message: "ok"})

// 获取 Any 值
v := obj.Get("status")
if v.IsAny() {
    status := v.AsAny().Value().(Status)
    fmt.Println(status.Code)  // 200
}

// Any 类型支持 JSON 序列化
jsonBytes := obj.ToJSON()
// {"status":{"Code":200,"Message":"ok"}}
```

### 遍历对象

```go
obj := value.NewObject()
obj.PutAny("name", "张三")
obj.PutAny("age", 25)
obj.PutAny("city", "北京")

// 使用 ForEach
obj.ForEach(func(key string, value value.Value) bool {
    fmt.Printf("%s: %v\n", key, value.String())
    return true  // 返回 false 停止遍历
})

// 使用 Iter (Go 1.23+)
for k, v := range obj.Iter {
    fmt.Printf("%s: %v\n", k, v.String())
}
```

## 在 Web 框架中使用

### JSONObject

`JSONObject` 是 `value.Object` 的类型别名，在 Web 请求中广泛使用：

```go
// 获取 JSON 请求体
jsonObj, err := req.Json()
if err != nil {
    return nil, err
}

// 类型安全的获取
name := jsonObj.GetString("name")
age := jsonObj.GetInt("age")

// 绑定到结构体
var input UserInput
err := req.BindJSON(&input)

// 获取分页参数
page, err := req.JsonPage()
```

### 与 KV 的区别

| 特性 | KV | value.Object |
|------|-----|--------------|
| 类型 | `map[string]any` | 自定义结构体 |
| 类型安全 | 运行时检查 | 编译时检查 |
| 嵌套支持 | 需要手动转换 | 原生支持 |
| JSON 序列化 | 标准库 | 自定义实现 |
| 结构体解码 | 需要 mapstructure | 内置 Decode 方法 |

## 最佳实践

1. **优先使用结构体**：对于已知结构的数据，使用结构体而非 Object
2. **使用类型安全的获取方法**：如 `GetString`、`GetInt`，避免类型断言
3. **处理默认值**：使用 `GetXxxForDefault` 方法处理可能不存在的字段
4. **嵌套对象**：使用 `GetObject` 获取嵌套对象，避免类型断言
5. **JSON 标签**：使用 `json` 标签确保与 JSON 字段名一致

## 完整示例

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/chuccp/go-web-frame/value"
)

func main() {
    // 创建用户对象
    user := value.NewObject()
    user.PutAny("id", 1)
    user.PutAny("name", "张三")
    user.PutAny("email", "zhangsan@example.com")
    user.PutAny("active", true)

    // 创建地址对象
    address := value.NewObject()
    address.PutAny("city", "北京")
    address.PutAny("street", "朝阳路 123 号")
    address.PutAny("zip_code", "100000")
    user.Put("address", address)

    // 创建标签数组
    tags := value.NewArray(
        value.NewText("developer"),
        value.NewText("go"),
        value.NewText("backend"),
    )
    user.Put("tags", tags)

    // 获取值
    fmt.Printf("ID: %d\n", user.GetInt("id"))
    fmt.Printf("Name: %s\n", user.GetString("name"))
    fmt.Printf("Email: %s\n", user.GetString("email"))
    fmt.Printf("Active: %v\n", user.GetBool("active"))
    fmt.Printf("City: %s\n", user.GetObject("address").GetString("city"))

    // 遍历标签
    fmt.Print("Tags: ")
    user.GetArray("tags").ForEach(func(i int, v value.Value) bool {
        if i > 0 {
            fmt.Print(", ")
        }
        fmt.Print(v.String())
        return true
    })
    fmt.Println()

    // 转换为 JSON
    jsonBytes := user.ToJSON()
    fmt.Printf("JSON: %s\n", string(jsonBytes))

    // 解码到结构体
    type User struct {
        ID      int    `json:"id"`
        Name    string `json:"name"`
        Email   string `json:"email"`
        Active  bool   `json:"active"`
        Address struct {
            City    string `json:"city"`
            Street  string `json:"street"`
            ZipCode string `json:"zip_code"`
        } `json:"address"`
        Tags []string `json:"tags"`
    }

    var u User
    if err := user.Decode(&u); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Decoded: %+v\n", u)
}
```
