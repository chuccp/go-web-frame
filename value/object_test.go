package value

import (
	"encoding/json"
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
