# Value Package — Type-Safe Dynamic Values

The `value` package provides a type-safe way to work with dynamic JSON-like data structures.

## Overview

The `value` package contains the following core types:

| Type | Description | Use Case |
|------|-------------|----------|
| `Object` | Key-value pair collection | Like JSON objects |
| `Array` | Ordered collection of values | Like JSON arrays |
| `Text` | String value | Text data |
| `Number` | Numeric value | Integer or float |
| `Bool` | Boolean value | true/false |
| `Any` | Any value | Non-standard types (structs, etc.) |
| `Null` | Null value | null |

## Basic Usage

### Creating Objects

```go
import "github.com/chuccp/go-web-frame/value"

// Create empty object
obj := value.NewObject()

// Create from map
m := map[string]any{"name": "Alice", "age": 25}
obj := value.NewObjectFromMap(m)

// Create from JSON
jsonBytes := []byte(`{"name":"Alice","age":25}`)
obj, err := value.NewObjectFromJson(jsonBytes)
```

### Setting and Getting Values

```go
// Set values
obj.PutAny("name", "Alice")     // Auto-converts type
obj.PutAny("age", 25)
obj.PutAny("active", true)
obj.Put("key", value.NewText("value"))  // Use concrete type

// Get values
name := obj.GetString("name")         // "Alice"
age := obj.GetInt("age")              // 25
active := obj.GetBool("active")       // true
score := obj.GetNumber("score")       // 0.0 (zero value if not exists)

// Get with default values
count := obj.GetIntForDefault("count", 0)        // 0 (default)
name := obj.GetStringForDefault("name", "Anonymous")  // "Anonymous" (default)

// Check if key exists
if obj.HasKey("name") {
    // Key exists
}

// Check key-value pair
if obj.HasKeyValue("name", "Alice") {
    // Key exists and value equals "Alice"
}
```

### Nested Objects and Arrays

```go
// Create nested object
address := value.NewObject()
address.PutAny("city", "Beijing")
address.PutAny("street", "Chaoyang Road")
obj.Put("address", address)

// Get nested object
addr := obj.GetObject("address")
city := addr.GetString("city")  // "Beijing"

// Create array
tags := value.NewArray(
    value.NewText("go"),
    value.NewText("web"),
    value.NewText("framework"),
)
obj.Put("tags", tags)

// Get array
arr := obj.GetArray("tags")
arr.ForEach(func(i int, v value.Value) bool {
    fmt.Println(v.String())  // go, web, framework
    return true
})
```

## Advanced Usage

### Decoding to Struct

Use the `Decode` method to convert an Object to a Go struct:

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
obj.PutAny("name", "Alice")
obj.PutAny("age", 25)
obj.PutAny("email", "alice@example.com")

address := value.NewObject()
address.PutAny("city", "Beijing")
address.PutAny("zip_code", "100000")
obj.Put("address", address)

var user User
err := obj.Decode(&user)
if err != nil {
    log.Fatal(err)
}

fmt.Println(user.Name)           // Alice
fmt.Println(user.Address.City)   // Beijing
```

### Converting to Map

```go
obj := value.NewObject()
obj.PutAny("name", "Alice")
obj.PutAny("age", 25)

m := obj.ToMap()
// m = map[string]any{"name": "Alice", "age": int64(25)}
```

### JSON Serialization

```go
obj := value.NewObject()
obj.PutAny("name", "Alice")
obj.PutAny("age", 25)

// Convert to JSON
jsonBytes := obj.ToJSON()
fmt.Println(string(jsonBytes))  // {"name":"Alice","age":25}

// Implements json.Marshaler interface
data, err := json.Marshal(obj)

// Implements json.Unmarshaler interface
var obj2 value.Object
err := json.Unmarshal(data, &obj2)
```

### Any Type

The `Any` type is used for non-standard values (like custom structs):

```go
type Status struct {
    Code    int
    Message string
}

obj := value.NewObject()
obj.PutAny("status", Status{Code: 200, Message: "ok"})

// Get Any value
v := obj.Get("status")
if v.IsAny() {
    status := v.AsAny().Value().(Status)
    fmt.Println(status.Code)  // 200
}

// Any type supports JSON serialization
jsonBytes := obj.ToJSON()
// {"status":{"Code":200,"Message":"ok"}}
```

### Iterating Over Objects

```go
obj := value.NewObject()
obj.PutAny("name", "Alice")
obj.PutAny("age", 25)
obj.PutAny("city", "Beijing")

// Using ForEach
obj.ForEach(func(key string, value value.Value) bool {
    fmt.Printf("%s: %v\n", key, value.String())
    return true  // Return false to stop iteration
})

// Using Iter (Go 1.23+)
for k, v := range obj.Iter {
    fmt.Printf("%s: %v\n", k, v.String())
}
```

## Using in Web Framework

### JSONObject

`JSONObject` is a type alias for `value.Object`, widely used in web requests:

```go
// Get JSON request body
jsonObj, err := req.Json()
if err != nil {
    return nil, err
}

// Type-safe getters
name := jsonObj.GetString("name")
age := jsonObj.GetInt("age")

// Bind to struct
var input UserInput
err := req.BindJSON(&input)

// Get pagination parameters
page, err := req.JsonPage()
```

### Comparison with KV

| Feature | KV | value.Object |
|---------|-----|--------------|
| Type | `map[string]any` | Custom struct |
| Type safety | Runtime check | Compile-time check |
| Nested support | Manual conversion | Native support |
| JSON serialization | Standard library | Custom implementation |
| Struct decoding | Requires mapstructure | Built-in Decode method |

## Best Practices

1. **Prefer structs**: For known data structures, use structs instead of Object
2. **Use type-safe getters**: Like `GetString`, `GetInt`, avoid type assertions
3. **Handle defaults**: Use `GetXxxForDefault` methods for potentially missing fields
4. **Nested objects**: Use `GetObject` for nested objects, avoid type assertions
5. **JSON tags**: Use `json` tags to ensure consistency with JSON field names

## Complete Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/chuccp/go-web-frame/value"
)

func main() {
    // Create user object
    user := value.NewObject()
    user.PutAny("id", 1)
    user.PutAny("name", "Alice")
    user.PutAny("email", "alice@example.com")
    user.PutAny("active", true)

    // Create address object
    address := value.NewObject()
    address.PutAny("city", "Beijing")
    address.PutAny("street", "Chaoyang Road 123")
    address.PutAny("zip_code", "100000")
    user.Put("address", address)

    // Create tags array
    tags := value.NewArray(
        value.NewText("developer"),
        value.NewText("go"),
        value.NewText("backend"),
    )
    user.Put("tags", tags)

    // Get values
    fmt.Printf("ID: %d\n", user.GetInt("id"))
    fmt.Printf("Name: %s\n", user.GetString("name"))
    fmt.Printf("Email: %s\n", user.GetString("email"))
    fmt.Printf("Active: %v\n", user.GetBool("active"))
    fmt.Printf("City: %s\n", user.GetObject("address").GetString("city"))

    // Iterate tags
    fmt.Print("Tags: ")
    user.GetArray("tags").ForEach(func(i int, v value.Value) bool {
        if i > 0 {
            fmt.Print(", ")
        }
        fmt.Print(v.String())
        return true
    })
    fmt.Println()

    // Convert to JSON
    jsonBytes := user.ToJSON()
    fmt.Printf("JSON: %s\n", string(jsonBytes))

    // Decode to struct
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
