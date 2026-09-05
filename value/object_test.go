package value

import (
	"encoding/json"
	"strings"
	"testing"
)

// 命名类型用于验证 fromReflect 兜底（type switch 精确匹配会漏掉它们）。
type testLevel string
type testCount int

func TestObjectNilSafe(t *testing.T) {
	var o *Object
	if got := o.Get("x"); got != nil {
		t.Errorf("nil Object.Get 应返回 nil, got %#v", got)
	}
	if !o.IsEmpty() {
		t.Error("nil Object.IsEmpty 应返回 true")
	}
	if o.HasKey("x") {
		t.Error("nil Object.HasKey 应返回 false")
	}
	if got := o.GetString("x"); got != "" {
		t.Errorf("nil Object.GetString 应返回空串, got %q", got)
	}
	if got := o.GetMustString("x"); got != "" {
		t.Errorf("nil Object.GetMustString 应返回空串, got %q", got)
	}
	if got := o.GetBool("x"); got {
		t.Error("nil Object.GetBool 应返回 false")
	}
	if got := o.GetNumber("x"); got != 0 {
		t.Errorf("nil Object.GetNumber 应返回 0, got %v", got)
	}
	if got := o.GetInt("x"); got != 0 {
		t.Errorf("nil Object.GetInt 应返回 0, got %v", got)
	}
	if got := o.GetObject("x"); got != nil {
		t.Errorf("nil Object.GetObject 应返回 nil, got %#v", got)
	}
	if got := o.GetArray("x"); got != nil {
		t.Errorf("nil Object.GetArray 应返回 nil, got %#v", got)
	}
	// 这些方法在 nil 上不应 panic。
	o.Delete("x")
	o.ForEach(func(k string, v Value) bool { return true })
	o.Iter(func(k string, v Value) bool { return true })
	if o.HasAnyKey("x", "y") {
		t.Error("nil Object.HasAnyKey 应返回 false")
	}
	if got := o.ToMap(); got != nil {
		t.Errorf("nil Object.ToMap 应返回 nil, got %#v", got)
	}
}

func TestObjectPutAnyNamedType(t *testing.T) {
	obj := NewObject()
	obj.PutAny("level", testLevel("high"))
	obj.PutAny("count", testCount(5))

	if !obj.Get("level").IsText() {
		t.Errorf("命名 string 类型应转为 Text, got %T", obj.Get("level"))
	}
	if got := obj.GetString("level"); got != "high" {
		t.Errorf("命名 string 类型取值错误, got %q", got)
	}
	if !obj.Get("count").IsNumber() {
		t.Errorf("命名 int 类型应转为 Number, got %T", obj.Get("count"))
	}
	if got := obj.GetInt("count"); got != 5 {
		t.Errorf("命名 int 类型取值错误, got %v", got)
	}
}

func TestObjectPutAnySliceString(t *testing.T) {
	obj := NewObject()
	obj.PutAny("tags", []string{"a", "b"})

	arr := obj.GetArray("tags")
	if arr == nil || arr.Len() != 2 {
		t.Fatalf("[]string 应转为 Array(2), got %#v", obj.Get("tags"))
	}
	vals := arr.StringValues()
	if len(vals) != 2 || vals[0] != "a" || vals[1] != "b" {
		t.Errorf("StringValues 错误: %#v", vals)
	}
}

func TestArrayNilSafe(t *testing.T) {
	var a *Array
	if got := a.Len(); got != 0 {
		t.Errorf("nil Array.Len 应返回 0, got %v", got)
	}
	if got := a.Get(0); got != NullValue {
		t.Errorf("nil Array.Get 应返回 NullValue, got %#v", got)
	}
	a.Set(0, NewText("x")) // 不应 panic
	a.ForEach(func(i int, v Value) bool { return true })
	a.Iter(func(i int, v Value) bool { return true })
	if got := a.StringValues(); got != nil {
		t.Errorf("nil Array.StringValues 应返回 nil, got %#v", got)
	}
}

func TestNumberIntFloatPreserved(t *testing.T) {
	if got := string(NewInt(3).ToJSON()); got != "3" {
		t.Errorf("NewInt(3) 应序列化为 3, got %s", got)
	}
	if got := string(NewNumber(3.5).ToJSON()); got != "3.5" {
		t.Errorf("NewNumber(3.5) 应序列化为 3.5, got %s", got)
	}
	if NewInt(3).IsFloat() {
		t.Error("NewInt 不应标记为 float")
	}
	if !NewNumber(3.5).IsFloat() {
		t.Error("NewNumber 应标记为 float")
	}
}

func TestObjectJSONRoundTrip(t *testing.T) {
	obj := NewObject()
	obj.Put("text", NewText(`he said "hi"`))
	obj.Put("n", NewInt(42))
	obj.Put("arr", NewArray(NewText("a"), NewText("b")))

	data := obj.ToJSON()
	parsed, err := NewObjectFromJson(data)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if got := parsed.GetString("text"); got != `he said "hi"` {
		t.Errorf("文本往返出错(可能二次转义): %q", got)
	}
	if got := parsed.GetInt("n"); got != 42 {
		t.Errorf("整数往返出错: %v", got)
	}
	arr := parsed.GetArray("arr")
	if arr == nil || arr.Len() != 2 || arr.StringValues()[0] != "a" {
		t.Errorf("数组往返出错: %#v", arr)
	}
}

func TestObjectFromJSONNumber(t *testing.T) {
	// JSON 数字默认按 float64 解码，应保持为 Number 且 GetInt 可用。
	raw := json.RawMessage(`{"n": 3}`)
	obj, err := NewObjectFromJson(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := obj.GetInt("n"); got != 3 {
		t.Errorf("GetInt 应为 3, got %v", got)
	}
}

// 自定义类型用于测试 Any 类型的处理。
type testStatus struct {
	Code    int
	Message string
}

func TestAnyTypeBasic(t *testing.T) {
	// 测试创建 Any 类型
	status := testStatus{Code: 200, Message: "ok"}
	anyVal := NewAny(status)

	if !anyVal.IsAny() {
		t.Error("Any.IsAny() 应返回 true")
	}

	if anyVal.AsAny() != anyVal {
		t.Error("Any.AsAny() 应返回自身")
	}

	if anyVal.Value() != status {
		t.Error("Any.Value() 应返回原始值")
	}
}

func TestAnyTypeString(t *testing.T) {
	// 测试 String() 方法
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"struct", testStatus{Code: 200, Message: "ok"}, "{200 ok}"},
		{"slice", []int{1, 2, 3}, "[1 2 3]"},
		{"map", map[string]int{"a": 1}, "map[a:1]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anyVal := NewAny(tt.value)
			if got := anyVal.String(); got != tt.want {
				t.Errorf("Any.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAnyTypeJSON(t *testing.T) {
	// 测试 ToJSON() 方法
	status := testStatus{Code: 200, Message: "ok"}
	anyVal := NewAny(status)

	jsonData := anyVal.ToJSON()
	if len(jsonData) == 0 {
		t.Error("Any.ToJSON() 不应返回空")
	}

	// 验证 JSON 可以被解析
	var parsed testStatus
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Any.ToJSON() 返回的 JSON 无法解析: %v", err)
	}
	if parsed.Code != 200 || parsed.Message != "ok" {
		t.Errorf("JSON 解析结果错误: %+v", parsed)
	}
}

func TestAnyTypeInObject(t *testing.T) {
	// 测试 Any 类型在 Object 中的使用
	obj := NewObject()
	status := testStatus{Code: 404, Message: "not found"}
	obj.PutAny("status", status)

	// 获取值并验证类型
	v := obj.Get("status")
	if v == nil {
		t.Fatal("Get('status') 不应返回 nil")
	}

	if !v.IsAny() {
		t.Errorf("期望 Any 类型, got %T", v)
	}

	anyVal := v.AsAny()
	if anyVal.Value() != status {
		t.Errorf("Any.Value() 返回错误值: %+v", anyVal.Value())
	}
}

func TestAnyTypeToMap(t *testing.T) {
	// 测试 Any 类型在 ToMap() 时的处理
	obj := NewObject()
	status := testStatus{Code: 500, Message: "error"}
	obj.PutAny("status", status)
	obj.PutAny("tags", []string{"a", "b"})

	m := obj.ToMap()
	if m == nil {
		t.Fatal("ToMap() 不应返回 nil")
	}

	// 验证 Any 类型能正确还原
	parsedStatus, ok := m["status"].(testStatus)
	if !ok {
		t.Fatalf("ToMap() 中 status 类型错误: %T", m["status"])
	}
	if parsedStatus.Code != 500 || parsedStatus.Message != "error" {
		t.Errorf("ToMap() 中 status 值错误: %+v", parsedStatus)
	}

	// 验证切片能正确处理（注意：[]string 会转为 []interface{}）
	tags, ok := m["tags"].([]interface{})
	if !ok {
		t.Fatalf("ToMap() 中 tags 类型错误: %T", m["tags"])
	}
	if len(tags) != 2 || tags[0] != "a" || tags[1] != "b" {
		t.Errorf("ToMap() 中 tags 值错误: %v", tags)
	}
}

func TestHasKeyValueWithAny(t *testing.T) {
	// 测试 HasKeyValue 方法对 Any 类型的支持
	obj := NewObject()
	status := testStatus{Code: 200, Message: "ok"}
	obj.PutAny("status", status)

	// 测试匹配的情况
	if !obj.HasKeyValue("status", status) {
		t.Error("HasKeyValue 应匹配相同的 Any 值")
	}

	// 测试不匹配的情况
	otherStatus := testStatus{Code: 404, Message: "not found"}
	if obj.HasKeyValue("status", otherStatus) {
		t.Error("HasKeyValue 不应匹配不同的值")
	}

	// 测试不存在的键
	if obj.HasKeyValue("nonexistent", status) {
		t.Error("HasKeyValue 不应匹配不存在的键")
	}

	// 测试 nil 对象
	var nilObj *Object
	if nilObj.HasKeyValue("key", "value") {
		t.Error("nil Object.HasKeyValue 应返回 false")
	}
}

func TestHasKeyValuePrimitives(t *testing.T) {
	obj := NewObject()
	obj.PutAny("name", "alice")
	obj.PutAny("age", 30)
	obj.PutAny("score", 95.5)
	obj.PutAny("active", true)
	obj.PutAny("note", nil)

	tests := []struct {
		key   string
		value any
		want  bool
	}{
		{"name", "alice", true},
		{"name", "bob", false},
		{"age", 30, true},
		{"age", 99, false},
		{"score", 95.5, true},
		{"score", 1.0, false},
		{"active", true, true},
		{"active", false, false},
		{"note", nil, true},
		{"note", "x", false},
		{"missing", "x", false},
	}
	for _, tt := range tests {
		got := obj.HasKeyValue(tt.key, tt.value)
		if got != tt.want {
			t.Errorf("HasKeyValue(%q, %v) = %v, want %v", tt.key, tt.value, got, tt.want)
		}
	}
}

func TestHasKeyValueMap(t *testing.T) {
	obj := NewObject()
	obj.PutAny("info", map[string]any{"city": "beijing", "zip": "100000"})

	if !obj.HasKeyValue("info", map[string]any{"city": "beijing", "zip": "100000"}) {
		t.Error("HasKeyValue 应匹配相同的 map")
	}
	if obj.HasKeyValue("info", map[string]any{"city": "shanghai"}) {
		t.Error("HasKeyValue 不应匹配不同的 map")
	}
	if obj.HasKeyValue("info", map[string]any{"city": "beijing", "zip": "100000", "extra": 1}) {
		t.Error("HasKeyValue 不应匹配不同长度的 map")
	}
}

func TestHasKeyValueSlice(t *testing.T) {
	obj := NewObject()
	obj.PutAny("tags", []string{"go", "web"})

	if !obj.HasKeyValue("tags", []string{"go", "web"}) {
		t.Error("HasKeyValue 应匹配相同的 slice")
	}
	if obj.HasKeyValue("tags", []string{"go", "rust"}) {
		t.Error("HasKeyValue 不应匹配不同的 slice")
	}
}

func TestObjectDecode(t *testing.T) {
	// 测试 Decode 方法 - 将 Object 转换为结构体
	type User struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	obj := NewObject()
	obj.PutAny("name", "张三")
	obj.PutAny("age", 25)
	obj.PutAny("email", "zhangsan@example.com")

	var user User
	err := obj.Decode(&user)
	if err != nil {
		t.Fatalf("Decode 失败: %v", err)
	}

	if user.Name != "张三" {
		t.Errorf("Name 错误: got %q, want %q", user.Name, "张三")
	}
	if user.Age != 25 {
		t.Errorf("Age 错误: got %d, want %d", user.Age, 25)
	}
	if user.Email != "zhangsan@example.com" {
		t.Errorf("Email 错误: got %q, want %q", user.Email, "zhangsan@example.com")
	}
}

func TestObjectDecodeNested(t *testing.T) {
	// 测试嵌套结构体的 Decode
	type Address struct {
		City  string `json:"city"`
		Phone string `json:"phone"`
	}
	type User struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	obj := NewObject()
	obj.PutAny("name", "李四")
	addressObj := NewObject()
	addressObj.PutAny("city", "北京")
	addressObj.PutAny("phone", "12345678901")
	obj.Put("address", addressObj)

	var user User
	err := obj.Decode(&user)
	if err != nil {
		t.Fatalf("Decode 失败: %v", err)
	}

	if user.Name != "李四" {
		t.Errorf("Name 错误: got %q, want %q", user.Name, "李四")
	}
	if user.Address.City != "北京" {
		t.Errorf("Address.City 错误: got %q, want %q", user.Address.City, "北京")
	}
	if user.Address.Phone != "12345678901" {
		t.Errorf("Address.Phone 错误: got %q, want %q", user.Address.Phone, "12345678901")
	}
}

func TestObjectDecodeNil(t *testing.T) {
	// 测试 nil Object 的 Decode
	var nilObj *Object
	type User struct {
		Name string `json:"name"`
	}
	var user User
	err := nilObj.Decode(&user)
	if err != nil {
		t.Errorf("nil Object.Decode 应返回 nil, got %v", err)
	}
}

func TestObjectDecodeToMap(t *testing.T) {
	// 测试 Decode 到 map
	obj := NewObject()
	obj.PutAny("key1", "value1")
	obj.PutAny("key2", 42)

	var m map[string]any
	err := obj.Decode(&m)
	if err != nil {
		t.Fatalf("Decode 失败: %v", err)
	}

	if m["key1"] != "value1" {
		t.Errorf("key1 错误: got %v, want %q", m["key1"], "value1")
	}
	// 验证数字类型能正确解码
	key2Val := m["key2"]
	if key2Val == nil {
		t.Error("key2 不应为 nil")
	} else {
		// 尝试转换为 int64 或 float64
		switch v := key2Val.(type) {
		case int64:
			if v != 42 {
				t.Errorf("key2 值错误: got %v, want 42", v)
			}
		case float64:
			if v != 42 {
				t.Errorf("key2 值错误: got %v, want 42", v)
			}
		default:
			t.Errorf("key2 类型错误: got %T, want int64 or float64", key2Val)
		}
	}
}

func TestDecodeJSONObject(t *testing.T) {
	// 测试解码 JSON 对象
	jsonStr := `{"name":"张三","age":25,"active":true}`
	r := strings.NewReader(jsonStr)

	val, err := DecodeJSON(r)
	if err != nil {
		t.Fatalf("DecodeJSON 失败: %v", err)
	}

	if !val.IsObject() {
		t.Fatalf("期望 Object 类型, got %T", val)
	}

	obj := val.AsObject()
	if obj.GetString("name") != "张三" {
		t.Errorf("name 错误: got %q", obj.GetString("name"))
	}
	if obj.GetInt("age") != 25 {
		t.Errorf("age 错误: got %d", obj.GetInt("age"))
	}
	if !obj.GetBool("active") {
		t.Error("active 应为 true")
	}
}

func TestDecodeJSONArray(t *testing.T) {
	// 测试解码 JSON 数组
	jsonStr := `["a","b","c"]`
	r := strings.NewReader(jsonStr)

	val, err := DecodeJSON(r)
	if err != nil {
		t.Fatalf("DecodeJSON 失败: %v", err)
	}

	if !val.IsArray() {
		t.Fatalf("期望 Array 类型, got %T", val)
	}

	arr := val.AsArray()
	if arr.Len() != 3 {
		t.Fatalf("数组长度错误: got %d, want 3", arr.Len())
	}
	if arr.Get(0).String() != "a" {
		t.Errorf("第一个元素错误: got %q", arr.Get(0).String())
	}
}

func TestDecodeJSONNested(t *testing.T) {
	// 测试解码嵌套 JSON
	jsonStr := `{
		"user": {
			"name": "李四",
			"address": {
				"city": "北京"
			}
		},
		"tags": ["go", "web"]
	}`
	r := strings.NewReader(jsonStr)

	val, err := DecodeJSON(r)
	if err != nil {
		t.Fatalf("DecodeJSON 失败: %v", err)
	}

	obj := val.AsObject()
	user := obj.GetObject("user")
	if user == nil {
		t.Fatal("user 不应为 nil")
	}
	if user.GetString("name") != "李四" {
		t.Errorf("user.name 错误: got %q", user.GetString("name"))
	}

	address := user.GetObject("address")
	if address == nil {
		t.Fatal("address 不应为 nil")
	}
	if address.GetString("city") != "北京" {
		t.Errorf("address.city 错误: got %q", address.GetString("city"))
	}

	tags := obj.GetArray("tags")
	if tags == nil || tags.Len() != 2 {
		t.Fatalf("tags 错误: %v", tags)
	}
}

func TestParseJSONObject(t *testing.T) {
	raw := json.RawMessage(`{"name":"张三","age":25,"active":true}`)
	val, err := ParseJSON(raw)
	if err != nil {
		t.Fatalf("ParseJSON 失败: %v", err)
	}
	if !val.IsObject() {
		t.Fatalf("期望 Object 类型, got %T", val)
	}
	obj := val.AsObject()
	if obj.GetString("name") != "张三" {
		t.Errorf("name 错误: got %q", obj.GetString("name"))
	}
	if obj.GetInt("age") != 25 {
		t.Errorf("age 错误: got %d", obj.GetInt("age"))
	}
	if !obj.GetBool("active") {
		t.Error("active 应为 true")
	}
}

func TestParseJSONArray(t *testing.T) {
	raw := json.RawMessage(`["a","b","c"]`)
	val, err := ParseJSON(raw)
	if err != nil {
		t.Fatalf("ParseJSON 失败: %v", err)
	}
	if !val.IsArray() {
		t.Fatalf("期望 Array 类型, got %T", val)
	}
	arr := val.AsArray()
	if arr.Len() != 3 {
		t.Fatalf("数组长度错误: got %d, want 3", arr.Len())
	}
	if arr.Get(0).String() != "a" {
		t.Errorf("第一个元素错误: got %q", arr.Get(0).String())
	}
}

func TestParseJSONNested(t *testing.T) {
	raw := json.RawMessage(`{"user":{"name":"李四","address":{"city":"北京"}},"tags":["go","web"]}`)
	val, err := ParseJSON(raw)
	if err != nil {
		t.Fatalf("ParseJSON 失败: %v", err)
	}
	obj := val.AsObject()
	user := obj.GetObject("user")
	if user == nil {
		t.Fatal("user 不应为 nil")
	}
	if user.GetString("name") != "李四" {
		t.Errorf("user.name 错误: got %q", user.GetString("name"))
	}
	address := user.GetObject("address")
	if address == nil {
		t.Fatal("address 不应为 nil")
	}
	if address.GetString("city") != "北京" {
		t.Errorf("address.city 错误: got %q", address.GetString("city"))
	}
	tags := obj.GetArray("tags")
	if tags == nil || tags.Len() != 2 {
		t.Fatalf("tags 错误: %v", tags)
	}
}

func TestParseJSONPrimitives(t *testing.T) {
	tests := []struct {
		name  string
		input json.RawMessage
		check func(Value) bool
	}{
		{"string", json.RawMessage(`"hello"`), func(v Value) bool { return v.IsText() && v.String() == "hello" }},
		{"number", json.RawMessage(`42`), func(v Value) bool { return v.IsNumber() && v.AsNumber().Int64() == 42 }},
		{"float", json.RawMessage(`3.14`), func(v Value) bool { return v.IsNumber() && v.AsNumber().Float64() == 3.14 }},
		{"bool", json.RawMessage(`true`), func(v Value) bool { return v.IsBool() && v.AsBool().String() == "true" }},
		{"null", json.RawMessage(`null`), func(v Value) bool { return v.IsNull() }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := ParseJSON(tt.input)
			if err != nil {
				t.Fatalf("ParseJSON 失败: %v", err)
			}
			if !tt.check(val) {
				t.Errorf("类型或值校验失败: %T %v", val, val)
			}
		})
	}
}

func TestParseJSONInvalid(t *testing.T) {
	_, err := ParseJSON(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("无效 JSON 应返回错误")
	}
}

func TestEqualText(t *testing.T) {
	if !NewText("hello").Equal(NewText("hello")) {
		t.Error("相同文本应相等")
	}
	if NewText("hello").Equal(NewText("world")) {
		t.Error("不同文本不应相等")
	}
	if NewText("hello").Equal(NewInt(1)) {
		t.Error("不同类型不应相等")
	}
}

func TestEqualNumber(t *testing.T) {
	if !NewInt(42).Equal(NewInt(42)) {
		t.Error("相同整数应相等")
	}
	if NewInt(42).Equal(NewInt(99)) {
		t.Error("不同整数不应相等")
	}
	if !NewNumber(3.14).Equal(NewNumber(3.14)) {
		t.Error("相同浮点数应相等")
	}
	if NewInt(1).Equal(NewNumber(1.0)) {
		t.Error("int 与 float 不应相等")
	}
	if NewInt(1).Equal(NewText("1")) {
		t.Error("不同类型不应相等")
	}
}

func TestEqualBool(t *testing.T) {
	if !NewBool(true).Equal(NewBool(true)) {
		t.Error("相同布尔值应相等")
	}
	if NewBool(true).Equal(NewBool(false)) {
		t.Error("不同布尔值不应相等")
	}
	if NewBool(true).Equal(NewText("true")) {
		t.Error("不同类型不应相等")
	}
}

func TestEqualNull(t *testing.T) {
	if !NullValue.Equal(NullValue) {
		t.Error("null 与 null 应相等")
	}
	if NullValue.Equal(NewText("")) {
		t.Error("null 与 text 不应相等")
	}
}

func TestEqualObject(t *testing.T) {
	a := NewObject()
	a.PutAny("name", "alice")
	a.PutAny("age", 30)

	b := NewObject()
	b.PutAny("name", "alice")
	b.PutAny("age", 30)

	if !a.Equal(b) {
		t.Error("相同对象应相等")
	}

	c := NewObject()
	c.PutAny("name", "bob")
	if a.Equal(c) {
		t.Error("不同对象不应相等")
	}

	d := NewObject()
	d.PutAny("name", "alice")
	if a.Equal(d) {
		t.Error("不同长度对象不应相等")
	}

	if a.Equal(NewText("x")) {
		t.Error("不同类型不应相等")
	}
}

func TestEqualArray(t *testing.T) {
	a := NewArray(NewText("a"), NewText("b"))
	b := NewArray(NewText("a"), NewText("b"))
	if !a.Equal(b) {
		t.Error("相同数组应相等")
	}

	c := NewArray(NewText("a"), NewText("c"))
	if a.Equal(c) {
		t.Error("不同数组不应相等")
	}

	d := NewArray(NewText("a"))
	if a.Equal(d) {
		t.Error("不同长度数组不应相等")
	}

	if a.Equal(NewText("x")) {
		t.Error("不同类型不应相等")
	}
}

func TestEqualNested(t *testing.T) {
	a := NewObject()
	a.Put("user", NewObject())
	a.AsObject().GetObject("user").PutAny("name", "alice")
	a.Put("tags", NewArray(NewText("go"), NewText("web")))

	b := NewObject()
	b.Put("user", NewObject())
	b.AsObject().GetObject("user").PutAny("name", "alice")
	b.Put("tags", NewArray(NewText("go"), NewText("web")))

	if !a.Equal(b) {
		t.Error("嵌套结构相同应相等")
	}

	b.AsObject().GetObject("user").PutAny("name", "bob")
	if a.Equal(b) {
		t.Error("嵌套结构不同不应相等")
	}
}

func TestHasAnyKey(t *testing.T) {
	obj := NewObject()
	obj.PutAny("name", "alice")
	obj.PutAny("age", 30)

	if !obj.HasAnyKey("name") {
		t.Error("单个存在的 key 应返回 true")
	}
	if !obj.HasAnyKey("name", "age") {
		t.Error("多个存在的 key 应返回 true")
	}
	if !obj.HasAnyKey("missing", "age") {
		t.Error("部分存在的 key 应返回 true")
	}
	if obj.HasAnyKey("x", "y", "z") {
		t.Error("全部不存在的 key 应返回 false")
	}
	if obj.HasAnyKey() {
		t.Error("无参数应返回 false")
	}
}

// --- Object.GetByPath / PutByPath ---

func TestObjectGetByPath(t *testing.T) {
	obj := NewObject()
	dbObj := NewObject()
	dbObj.PutAny("type", "mysql")
	dbObj.PutAny("host", "localhost")
	obj.Put("db", dbObj)

	// Simple key
	if got := obj.GetByPath("db"); got == nil || !got.IsObject() {
		t.Error("GetByPath('db') 应返回 Object")
	}

	// Nested key
	if got := obj.GetByPath("db.type"); got == nil || got.String() != "mysql" {
		t.Errorf("GetByPath('db.type') 应返回 'mysql', got %v", got)
	}
	if got := obj.GetByPath("db.host"); got == nil || got.String() != "localhost" {
		t.Errorf("GetByPath('db.host') 应返回 'localhost', got %v", got)
	}

	// Missing key
	if got := obj.GetByPath("db.missing"); got != nil {
		t.Errorf("GetByPath('db.missing') 应返回 nil, got %v", got)
	}

	// Missing top-level key
	if got := obj.GetByPath("missing.key"); got != nil {
		t.Errorf("GetByPath('missing.key') 应返回 nil, got %v", got)
	}

	// Deep nested
	serverObj := NewObject()
	sslObj := NewObject()
	sslObj.PutAny("enabled", true)
	serverObj.Put("ssl", sslObj)
	obj.Put("server", serverObj)
	if got := obj.GetByPath("server.ssl.enabled"); got == nil || got.String() != "true" {
		t.Errorf("GetByPath('server.ssl.enabled') 应返回 true, got %v", got)
	}
}

func TestObjectGetByPathNilSafe(t *testing.T) {
	var obj *Object
	if got := obj.GetByPath("any.path"); got != nil {
		t.Errorf("nil Object.GetByPath 应返回 nil, got %v", got)
	}
}

func TestObjectPutByPath(t *testing.T) {
	obj := NewObject()

	// Simple key
	obj.PutByPath("name", "alice")
	if got := obj.GetString("name"); got != "alice" {
		t.Errorf("PutByPath('name') 失败, got %q", got)
	}

	// Nested key - creates intermediate objects
	obj.PutByPath("db.type", "sqlite")
	obj.PutByPath("db.path", ":memory:")

	dbVal := obj.GetByPath("db")
	if dbVal == nil || !dbVal.IsObject() {
		t.Fatal("db 应为 Object")
	}
	db := dbVal.AsObject()
	if db.GetString("type") != "sqlite" {
		t.Errorf("db.type 应为 'sqlite', got %q", db.GetString("type"))
	}
	if db.GetString("path") != ":memory:" {
		t.Errorf("db.path 应为 ':memory:', got %q", db.GetString("path"))
	}

	// Deep nested
	obj.PutByPath("server.ssl.cert", "/path/to/cert")
	if got := obj.GetByPath("server.ssl.cert"); got == nil || got.String() != "/path/to/cert" {
		t.Errorf("PutByPath deep nested 失败, got %v", got)
	}

	// Overwrite existing
	obj.PutByPath("db.type", "mysql")
	if got := obj.GetByPath("db.type"); got == nil || got.String() != "mysql" {
		t.Errorf("PutByPath overwrite 失败, got %v", got)
	}
	// Verify other keys in db preserved
	if got := obj.GetByPath("db.path"); got == nil || got.String() != ":memory:" {
		t.Errorf("PutByPath overwrite 不应影响其他 key, got %v", got)
	}
}

// --- Object.AddAll deep merge ---

func TestObjectAddAll_Shallow(t *testing.T) {
	a := NewObject()
	a.PutAny("x", 1)

	b := NewObject()
	b.PutAny("y", 2)

	a.AddAll(b)
	if a.GetInt("x") != 1 || a.GetInt("y") != 2 {
		t.Errorf("AddAll shallow 合并错误: %+v", a.ToMap())
	}
}

func TestObjectAddAll_DeepMerge(t *testing.T) {
	a := NewObject()
	dbA := NewObject()
	dbA.PutAny("type", "mysql")
	dbA.PutAny("host", "localhost")
	a.Put("db", dbA)

	b := NewObject()
	dbB := NewObject()
	dbB.PutAny("host", "production")
	dbB.PutAny("port", 3306)
	b.Put("db", dbB)

	a.AddAll(b)

	db := a.GetObject("db")
	if db == nil {
		t.Fatal("db 不应为 nil")
	}
	// type preserved from a
	if db.GetString("type") != "mysql" {
		t.Errorf("db.type 应保留 'mysql', got %q", db.GetString("type"))
	}
	// host overwritten by b
	if db.GetString("host") != "production" {
		t.Errorf("db.host 应被覆盖为 'production', got %q", db.GetString("host"))
	}
	// port added from b
	if db.GetInt("port") != 3306 {
		t.Errorf("db.port 应为 3306, got %d", db.GetInt("port"))
	}
}

func TestObjectAddAll_NilSafe(t *testing.T) {
	a := NewObject()
	a.PutAny("x", 1)
	a.AddAll(nil) // should not panic
	if a.GetInt("x") != 1 {
		t.Error("AddAll(nil) 不应影响原对象")
	}
}

// --- Object.GetNative ---

func TestObjectGetNative(t *testing.T) {
	obj := NewObject()
	obj.PutAny("name", "alice")
	obj.PutAny("age", 30)
	obj.PutAny("score", 95.5)
	obj.PutAny("active", true)

	if got := obj.GetNative("name"); got != "alice" {
		t.Errorf("GetNative('name') = %v, want 'alice'", got)
	}
	if got := obj.GetNative("age"); got != int64(30) {
		t.Errorf("GetNative('age') = %v (type %T), want int64(30)", got, got)
	}
	if got := obj.GetNative("score"); got != 95.5 {
		t.Errorf("GetNative('score') = %v, want 95.5", got)
	}
	if got := obj.GetNative("active"); got != true {
		t.Errorf("GetNative('active') = %v, want true", got)
	}
	if got := obj.GetNative("missing"); got != nil {
		t.Errorf("GetNative('missing') = %v, want nil", got)
	}
}

func TestObjectGetNativeNilSafe(t *testing.T) {
	var obj *Object
	if got := obj.GetNative("any"); got != nil {
		t.Errorf("nil GetNative 应返回 nil, got %v", got)
	}
}

// --- Object.ForEach / Iter ---

func TestObjectForEach(t *testing.T) {
	obj := NewObject()
	obj.PutAny("a", 1)
	obj.PutAny("b", 2)
	obj.PutAny("c", 3)

	collected := make(map[string]int)
	obj.ForEach(func(k string, v Value) bool {
		collected[k] = int(v.AsNumber().Int64())
		return true
	})
	if len(collected) != 3 {
		t.Errorf("ForEach 应遍历 3 个 key, got %d", len(collected))
	}
	if collected["a"] != 1 || collected["b"] != 2 || collected["c"] != 3 {
		t.Errorf("ForEach 值错误: %v", collected)
	}
}

func TestObjectForEach_Break(t *testing.T) {
	obj := NewObject()
	obj.PutAny("a", 1)
	obj.PutAny("b", 2)
	obj.PutAny("c", 3)

	count := 0
	obj.ForEach(func(k string, v Value) bool {
		count++
		return false // break after first
	})
	if count != 1 {
		t.Errorf("ForEach break 应只迭代 1 次, got %d", count)
	}
}

// --- Object.ReplaceKey ---

func TestObjectReplaceKey(t *testing.T) {
	obj := NewObject()
	obj.PutAny("old_name", "alice")

	obj.ReplaceKey("old_name", "new_name")
	// new key should have the value
	if obj.GetString("new_name") != "alice" {
		t.Errorf("ReplaceKey 应复制值到 new_name, got %q", obj.GetString("new_name"))
	}
	// old key still exists (ReplaceKey copies, doesn't move)
	if !obj.HasKey("old_name") {
		t.Error("ReplaceKey 后 old_name 应仍存在")
	}
	if obj.GetString("old_name") != "alice" {
		t.Errorf("old_name 值应保留, got %q", obj.GetString("old_name"))
	}
}

func TestObjectReplaceKey_NotExist(t *testing.T) {
	obj := NewObject()
	obj.ReplaceKey("missing", "new")
	if obj.HasKey("new") {
		t.Error("ReplaceKey 对不存在的 key 不应创建新 key")
	}
}

// --- Array.Add / AddAny / Set ---

func TestArrayAdd(t *testing.T) {
	arr := NewArray(NewText("a"))
	arr.Add(NewText("b"), NewText("c"))
	if arr.Len() != 3 {
		t.Fatalf("Add 后长度应为 3, got %d", arr.Len())
	}
	if arr.Get(1).String() != "b" || arr.Get(2).String() != "c" {
		t.Errorf("Add 值错误: [%s, %s]", arr.Get(1).String(), arr.Get(2).String())
	}
}

func TestArrayAddAny(t *testing.T) {
	arr := NewArray()
	arr.AddAny("hello")
	arr.AddAny(42)
	arr.AddAny(true)

	if arr.Len() != 3 {
		t.Fatalf("AddAny 后长度应为 3, got %d", arr.Len())
	}
	if !arr.Get(0).IsText() || arr.Get(0).String() != "hello" {
		t.Errorf("AddAny string 错误: %T %v", arr.Get(0), arr.Get(0))
	}
	if !arr.Get(1).IsNumber() || arr.Get(1).AsNumber().Int64() != 42 {
		t.Errorf("AddAny int 错误: %T %v", arr.Get(1), arr.Get(1))
	}
	if !arr.Get(2).IsBool() {
		t.Errorf("AddAny bool 错误: %T", arr.Get(2))
	}
}

func TestArraySet(t *testing.T) {
	arr := NewArray(NewText("a"), NewText("b"), NewText("c"))
	arr.Set(1, NewText("X"))
	if arr.Get(1).String() != "X" {
		t.Errorf("Set 后值应为 'X', got %q", arr.Get(1).String())
	}
	// Out of bounds should not panic
	arr.Set(10, NewText("y"))
	arr.Set(-1, NewText("z"))
}

// --- Array.Get out of bounds ---

func TestArrayGetOutOfBounds(t *testing.T) {
	arr := NewArray(NewText("a"))
	if got := arr.Get(5); got != NullValue {
		t.Errorf("越界 Get 应返回 NullValue, got %v", got)
	}
	if got := arr.Get(-1); got != NullValue {
		t.Errorf("负数 Get 应返回 NullValue, got %v", got)
	}
}

// --- Array.ToJSON ---

func TestArrayToJSON(t *testing.T) {
	arr := NewArray(NewText("a"), NewInt(1), NewBool(true), NullValue)
	data := arr.ToJSON()
	var out []any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("ToJSON 解析失败: %v", err)
	}
	if len(out) != 4 {
		t.Fatalf("长度应为 4, got %d", len(out))
	}
	if out[0] != "a" {
		t.Errorf("元素 0 应为 'a', got %v", out[0])
	}
	// JSON numbers default to float64
	if out[1] != float64(1) {
		t.Errorf("元素 1 应为 1, got %v", out[1])
	}
	if out[2] != true {
		t.Errorf("元素 2 应为 true, got %v", out[2])
	}
	if out[3] != nil {
		t.Errorf("元素 3 应为 nil, got %v", out[3])
	}
}

// --- Stream ---

func TestStreamBasic(t *testing.T) {
	s := NewStream()
	if !s.IsStream() {
		t.Error("IsStream 应返回 true")
	}
	if !s.IsEmpty() {
		t.Error("新 Stream 应为空")
	}
	if s.Len() != 0 {
		t.Errorf("新 Stream 长度应为 0, got %d", s.Len())
	}

	s.WriteString("hello")
	s.WriteString(" world")
	if s.Text() != "hello world" {
		t.Errorf("Text() = %q, want 'hello world'", s.Text())
	}
	if s.Len() != 11 {
		t.Errorf("Len() = %d, want 11", s.Len())
	}
	if s.IsEmpty() {
		t.Error("写入后不应为空")
	}
	if s.String() != "hello world" {
		t.Errorf("String() = %q, want 'hello world'", s.String())
	}
}

func TestStreamReset(t *testing.T) {
	s := NewStream()
	s.WriteString("data")
	s.Reset()
	if !s.IsEmpty() {
		t.Error("Reset 后应为空")
	}
	if s.Text() != "" {
		t.Errorf("Reset 后 Text() = %q, want ''", s.Text())
	}
}

func TestStreamToJSON(t *testing.T) {
	s := NewStream()
	s.WriteString(`{"key":"value"}`)
	data := s.ToJSON()
	if string(data) != `{"key":"value"}` {
		t.Errorf("ToJSON = %q", string(data))
	}
}

func TestStreamEqual(t *testing.T) {
	a := NewStream()
	a.WriteString("hello")
	b := NewStream()
	b.WriteString("hello")
	c := NewStream()
	c.WriteString("world")

	if !a.Equal(b) {
		t.Error("相同 Stream 应相等")
	}
	if a.Equal(c) {
		t.Error("不同 Stream 不应相等")
	}
	if a.Equal(NewText("hello")) {
		t.Error("不同类型不应相等")
	}
}

func TestStreamUnmarshal(t *testing.T) {
	s := NewStream()
	s.WriteString(`{"name":"alice","age":30}`)
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var u User
	err := s.Unmarshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "alice" || u.Age != 30 {
		t.Errorf("got %+v", u)
	}
}

// --- Text ---

func TestTextBasic(t *testing.T) {
	text := NewText("hello")
	if !text.IsText() {
		t.Error("IsText 应返回 true")
	}
	if text.AsText() != text {
		t.Error("AsText 应返回自身")
	}
	if text.String() != "hello" {
		t.Errorf("String() = %q, want 'hello'", text.String())
	}
}

func TestTextToJSON(t *testing.T) {
	text := NewText(`he said "hi"`)
	data := text.ToJSON()
	// Should be JSON-escaped
	if string(data) != `"he said \"hi\""` {
		t.Errorf("ToJSON = %q", string(data))
	}
}

func TestTextEqual(t *testing.T) {
	if !NewText("x").Equal(NewText("x")) {
		t.Error("相同 Text 应相等")
	}
	if NewText("x").Equal(NewText("y")) {
		t.Error("不同 Text 不应相等")
	}
	if NewText("x").Equal(NewInt(1)) {
		t.Error("不同类型不应相等")
	}
}

func TestTextUnmarshal(t *testing.T) {
	text := NewText(`{"key":"value"}`)
	var m map[string]string
	err := text.Unmarshal(&m)
	if err != nil {
		t.Fatal(err)
	}
	if m["key"] != "value" {
		t.Errorf("got %+v", m)
	}
}

// --- Number ---

func TestNumberString(t *testing.T) {
	if NewInt(42).String() != "42" {
		t.Errorf("NewInt(42).String() = %q", NewInt(42).String())
	}
	if NewNumber(3.14).String() != "3.14" {
		t.Errorf("NewNumber(3.14).String() = %q", NewNumber(3.14).String())
	}
}

func TestNumberInt64Float64(t *testing.T) {
	n := NewInt(42)
	if n.Int64() != 42 {
		t.Errorf("Int64() = %d", n.Int64())
	}
	if n.Float64() != 42.0 {
		t.Errorf("Float64() = %f", n.Float64())
	}

	f := NewNumber(3.5)
	if f.Int64() != 3 {
		t.Errorf("Float Int64() = %d", f.Int64())
	}
	if f.Float64() != 3.5 {
		t.Errorf("Float64() = %f", f.Float64())
	}
}

func TestNumberMarshalJSON(t *testing.T) {
	n := NewInt(42)
	data, err := n.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "42" {
		t.Errorf("MarshalJSON = %q", string(data))
	}

	f := NewNumber(3.14)
	data, err = f.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "3.14" {
		t.Errorf("MarshalJSON = %q", string(data))
	}
}

// --- Bool ---

func TestBoolBasic(t *testing.T) {
	if !NewBool(true).IsBool() {
		t.Error("IsBool 应返回 true")
	}
	if NewBool(true).String() != "true" {
		t.Errorf("String() = %q", NewBool(true).String())
	}
	if NewBool(false).String() != "false" {
		t.Errorf("String() = %q", NewBool(false).String())
	}
}

func TestBoolToJSON(t *testing.T) {
	if string(NewBool(true).ToJSON()) != "true" {
		t.Errorf("ToJSON = %q", string(NewBool(true).ToJSON()))
	}
	if string(NewBool(false).ToJSON()) != "false" {
		t.Errorf("ToJSON = %q", string(NewBool(false).ToJSON()))
	}
}

// --- Null ---

func TestNullBasic(t *testing.T) {
	if !NullValue.IsNull() {
		t.Error("IsNull 应返回 true")
	}
	if NullValue.String() != "null" {
		t.Errorf("String() = %q", NullValue.String())
	}
	if string(NullValue.ToJSON()) != "null" {
		t.Errorf("ToJSON = %q", string(NullValue.ToJSON()))
	}
}

func TestNullEqual(t *testing.T) {
	if !NullValue.Equal(NullValue) {
		t.Error("NullValue 应等于自身")
	}
	if NullValue.Equal(NewText("")) {
		t.Error("Null 不应等于 Text")
	}
}

// --- Object.PutJson ---

func TestObjectPutJson(t *testing.T) {
	obj := NewObject()
	err := obj.PutJson([]byte(`{"name":"alice","age":30}`))
	if err != nil {
		t.Fatal(err)
	}
	if obj.GetString("name") != "alice" {
		t.Errorf("name = %q", obj.GetString("name"))
	}
	if obj.GetInt("age") != 30 {
		t.Errorf("age = %d", obj.GetInt("age"))
	}
}

func TestObjectPutJson_Invalid(t *testing.T) {
	obj := NewObject()
	err := obj.PutJson([]byte(`{invalid`))
	if err == nil {
		t.Error("无效 JSON 应返回错误")
	}
}

// --- NewObjectFromMap ---

func TestNewObjectFromMap(t *testing.T) {
	m := map[string]any{
		"name":   "alice",
		"age":    30,
		"active": true,
	}
	obj := NewObjectFromMap(m)
	if obj.GetString("name") != "alice" {
		t.Errorf("name = %q", obj.GetString("name"))
	}
	if obj.GetInt("age") != 30 {
		t.Errorf("age = %d", obj.GetInt("age"))
	}
	if !obj.GetBool("active") {
		t.Error("active 应为 true")
	}
}

// --- Object.MarshalJSON ---

func TestObjectMarshalJSON(t *testing.T) {
	obj := NewObject()
	obj.PutAny("name", "alice")
	obj.PutAny("count", 5)

	data, err := obj.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["name"] != "alice" {
		t.Errorf("name = %v", m["name"])
	}
	if m["count"] != float64(5) {
		t.Errorf("count = %v", m["count"])
	}
}

// --- Object.Delete ---

func TestObjectDelete(t *testing.T) {
	obj := NewObject()
	obj.PutAny("name", "alice")
	obj.PutAny("age", 30)

	obj.Delete("name")
	if obj.HasKey("name") {
		t.Error("Delete 后 key 不应存在")
	}
	if obj.GetString("name") != "" {
		t.Errorf("Delete 后 GetString 应返回空, got %q", obj.GetString("name"))
	}
	// Other key preserved
	if obj.GetInt("age") != 30 {
		t.Errorf("Delete 不应影响其他 key, age = %d", obj.GetInt("age"))
	}
}

// --- Object.IsEmpty ---

func TestObjectIsEmpty(t *testing.T) {
	obj := NewObject()
	if !obj.IsEmpty() {
		t.Error("新 Object 应为空")
	}
	obj.PutAny("key", "value")
	if obj.IsEmpty() {
		t.Error("有 key 后不应为空")
	}
}

// --- Type assertion panics ---

func TestValueBaseAsPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("AsObject on non-Object should panic")
		}
	}()
	NewText("x").AsObject()
}

// --- Array nil safety extended ---

func TestArrayNilSafeExtended(t *testing.T) {
	var arr *Array
	// These should not panic
	arr.Set(0, NewText("x"))
	if arr.Len() != 0 {
		t.Errorf("nil Array.Len 应为 0, got %d", arr.Len())
	}
	if arr.Get(0) != NullValue {
		t.Error("nil Array.Get 应返回 NullValue")
	}
	arr.ForEach(func(i int, v Value) bool { return true })
	arr.Iter(func(i int, v Value) bool { return true })
}

// --- Object with null value ---

func TestObjectNullValue(t *testing.T) {
	obj := NewObject()
	obj.Put("key", NullValue)

	if !obj.Get("key").IsNull() {
		t.Error("Get('key') 应为 Null")
	}
	if obj.GetString("key") != "" {
		t.Errorf("Null GetString 应返回空, got %q", obj.GetString("key"))
	}
}

// --- NewArraySize ---

func TestNewArraySize(t *testing.T) {
	arr := NewArraySize(5)
	if arr.Len() != 5 {
		t.Errorf("NewArraySize(5).Len() = %d, want 5", arr.Len())
	}
	// Elements should be nil initially
	if arr.Get(0) != nil {
		t.Errorf("初始元素应为 nil, got %v", arr.Get(0))
	}
	// Can set values
	arr.Set(0, NewText("hello"))
	if arr.Get(0).String() != "hello" {
		t.Errorf("Set 后值应为 'hello', got %q", arr.Get(0).String())
	}
}
