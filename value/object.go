package value

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/chuccp/go-web-frame/util"
)

type Object struct {
	ValueBase
	data map[string]Value
}

func (o *Object) PutAny(key string, value any) {
	o.data[key] = fromInterface(value)
}
func (o *Object) PutMap(data map[string]any) {
	for k, v := range data {
		o.data[k] = fromInterface(v)
	}
}
func (o *Object) Put(key string, value Value) {
	o.data[key] = value
}
func (o *Object) Get(key string) Value {
	if o == nil {
		return nil
	}
	return o.data[key]
}
func (o *Object) GetByPath(path string) Value {
	parts := strings.Split(path, ".")
	var current Value = o
	for _, part := range parts {
		obj, ok := current.(*Object)
		if !ok {
			return nil
		}
		current = obj.Get(part)
		if current == nil {
			return nil
		}
	}
	return current
}

func (o *Object) PutByPath(path string, value any) {
	parts := strings.Split(path, ".")
	current := o
	for i := 0; i < len(parts)-1; i++ {
		child := current.Get(parts[i])
		childObj, ok := child.(*Object)
		if !ok {
			childObj = NewObject()
			current.Put(parts[i], childObj)
		}
		current = childObj
	}
	current.PutAny(parts[len(parts)-1], value)
}

// GetNative returns the native Go value for the given key.
// It converts Value types back to their original Go types (e.g., *Text -> string, *Number -> int/float64).
func (o *Object) GetNative(key string) any {
	if o == nil {
		return nil
	}
	v, ok := o.data[key]
	if !ok {
		return nil
	}
	return toAny(v)
}

func (o *Object) HasKeyValue(key string, value any) bool {
	if o == nil {
		return false
	}
	v, ok := o.data[key]
	if !ok {
		return false
	}
	return fromInterface(value).Equal(v)
}

func (o *Object) IsEmpty() bool {
	if o == nil {
		return true
	}
	return len(o.data) == 0
}
func (o *Object) GetMustString(key string) string {
	if o == nil {
		return ""
	}
	v := o.Get(key)
	if v == nil {
		log.Panic("GetString: " + key + " not found ")
	}

	if v.IsNull() {
		return ""
	}
	return v.String()

}

func (o *Object) GetString(key string) string {
	v := o.Get(key)
	if v == nil {
		return ""
	}
	if v.IsNull() {
		return ""
	}
	return v.String()
}

func (o *Object) HasKey(key string) bool {
	if o == nil {
		return false
	}
	_, ok := o.data[key]
	return ok
}

func (o *Object) HasAnyKey(key ...string) bool {
	if o == nil {
		return false
	}
	for _, k := range key {
		if _, ok := o.data[k]; ok {
			return true
		}
	}
	return false
}

func (o *Object) GetIntForDefault(key string, defaultValue int) int {
	v := o.Get(key)
	if v == nil || !v.IsNumber() {
		return defaultValue
	}
	return int(v.AsNumber().Int64())
}

func (o *Object) GetStringOrDefault(key string, defaultValue string) string {
	v := o.GetString(key)
	if util.IsBlank(v) {
		return defaultValue
	}
	return v
}

func (o *Object) GetBoolOrDefault(key string, defaultValue bool) bool {
	if o.HasKey(key) {
		return defaultValue
	}
	return o.GetBool(key)
}

func (o *Object) GetBool(key string) bool {
	v := o.Get(key)
	if v == nil || !v.IsBool() {
		return false
	}
	return v.AsBool().b
}

func (o *Object) GetNumber(key string) float64 {
	v := o.Get(key)
	if v == nil || !v.IsNumber() {
		return 0
	}
	return v.AsNumber().Float64()
}

func (o *Object) GetInt(key string) int {
	v := o.Get(key)
	if v == nil || !v.IsNumber() {
		return 0
	}
	return int(v.AsNumber().Int64())
}

func (o *Object) GetObject(key string) *Object {
	v := o.Get(key)
	if v == nil || !v.IsObject() {
		return nil
	}
	return v.AsObject()
}

func (o *Object) GetArray(key string) *Array {
	v := o.Get(key)
	if v == nil || !v.IsArray() {
		return nil
	}
	return v.AsArray()
}

func (o *Object) AddAll(other *Object) {
	if other == nil {
		return
	}
	for k, v := range other.data {
		if existing, ok := o.data[k]; ok {
			if eObj, ok1 := existing.(*Object); ok1 {
				if vObj, ok2 := v.(*Object); ok2 {
					eObj.AddAll(vObj)
					continue
				}
			}
		}
		o.data[k] = v
	}
}

func (o *Object) Delete(key string) {
	if o == nil {
		return
	}
	delete(o.data, key)
}

func (o *Object) ForEach(fn func(key string, value Value) bool) {
	if o == nil {
		return
	}
	for k, v := range o.data {
		if !fn(k, v) {
			break
		}
	}
}

// Iter 返回一个迭代器函数，支持 Go 1.23+ 的 for-range 语法。
// 遍历顺序不确定（底层为 map）。
func (o *Object) Iter(yield func(k string, v Value) bool) {
	if o == nil {
		return
	}
	for k, v := range o.data {
		if !yield(k, v) {
			return
		}
	}
}

// ToMap 将对象转换为原生 map（递归转换嵌套的 Object/Array）。
func (o *Object) ToMap() map[string]any {
	if o == nil {
		return nil
	}
	m := make(map[string]any, len(o.data))
	for k, v := range o.data {
		m[k] = toAny(v)
	}
	return m
}

// Decode decodes the Object into the provided struct using json tags.
// It converts the Object to a map first, then uses the value decoder.
func (o *Object) Decode(v any, opts ...DecoderConfigOption) error {
	if o == nil {
		return nil
	}
	m := o.ToMap()
	cfg := newDecoderConfig(opts...)
	return decodeValue(m, v, cfg)
}

func (o *Object) IsObject() bool { return true }

func (o *Object) AsObject() *Object { return o }

func (o *Object) String() string { return string(o.ToJSON()) }

func (o *Object) ToJSON() json.RawMessage {
	m := make(map[string]json.RawMessage, len(o.data))
	for k, v := range o.data {
		if v == nil {
			m[k] = json.RawMessage("null")
		} else {
			m[k] = v.ToJSON()
		}
	}
	data, _ := json.Marshal(m)
	return data
}

func (o *Object) MarshalJSON() ([]byte, error) { return o.ToJSON(), nil }

func (o *Object) Unmarshal(v any, opts ...DecoderConfigOption) error {
	if o == nil {
		return nil
	}
	cfg := newDecoderConfig(opts...)
	m := o.ToMap()
	return decodeValue(m, v, cfg)
}

func (o *Object) Equal(other Value) bool {
	obj, ok := other.(*Object)
	if !ok {
		return false
	}
	if len(o.data) != len(obj.data) {
		return false
	}
	for k, v := range o.data {
		ov, exists := obj.data[k]
		if !exists {
			return false
		}
		if v == nil && ov == nil {
			continue
		}
		if v == nil || ov == nil {
			return false
		}
		if !v.Equal(ov) {
			return false
		}
	}
	return true
}

func (o *Object) ReplaceKey(key string, newKey string) {
	if o.HasKey(key) {
		o.PutAny(newKey, o.Get(key))
	}
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (o *Object) UnmarshalJSON(data []byte) error {
	return o.PutJson(data)
}

// PutJson 解析 JSON 并填充到对象中。
func (o *Object) PutJson(dataJson []byte) error {
	var m map[string]any
	if err := json.Unmarshal(dataJson, &m); err != nil {
		return err
	}
	o.data = make(map[string]Value, len(m))
	for k, v := range m {
		o.data[k] = fromInterface(v)
	}
	return nil
}

func NewObject() *Object {
	return &Object{
		data: make(map[string]Value),
	}
}

// NewObjectFromMap 从原生 map 构建对象。
func NewObjectFromMap(m map[string]any) *Object {
	obj := NewObject()
	for k, v := range m {
		obj.data[k] = fromInterface(v)
	}
	return obj
}
func NewObjectFromJson(dataJson json.RawMessage) (*Object, error) {
	var m map[string]any
	err := json.Unmarshal(dataJson, &m)
	if err != nil {
		return NewObject(), err
	}
	return NewObjectFromMap(m), nil
}

// fromInterface 将 JSON 反序列化得到的原生值转换为对应的 Value 类型。
func fromInterface(v any) Value {
	switch val := v.(type) {
	case Value:
		return val
	case nil:
		return NullValue
	case bool:
		return NewBool(val)
	case map[string]any:
		obj := NewObject()
		for k, item := range val {
			obj.data[k] = fromInterface(item)
		}
		return obj
	case []any:
		arr := make([]Value, len(val))
		for index, item := range val {
			arr[index] = fromInterface(item)
		}
		return NewArray(arr...)
	case []string:
		arr := make([]Value, len(val))
		for index, item := range val {
			arr[index] = &Text{text: item}
		}
		return NewArray(arr...)
	case float64:
		return NewNumber(val)
	case float32:
		return NewNumber(float64(val))
	case int:
		return NewInt(int64(val))
	case int8:
		return NewInt(int64(val))
	case int16:
		return NewInt(int64(val))
	case int32:
		return NewInt(int64(val))
	case int64:
		return NewInt(val)
	case uint:
		return NewInt(int64(val))
	case uint8:
		return NewInt(int64(val))
	case uint16:
		return NewInt(int64(val))
	case uint32:
		return NewInt(int64(val))
	case uint64:
		return NewInt(int64(val))
	case string:
		return &Text{text: val}
	default:
		return fromReflect(v)
	}
}

// fromReflect 处理底层为原生类型的命名类型（如 ThinkingLevel、Role 等自定义 string/int 类型）。
// 类型 switch 只做精确匹配，命名类型会漏进 default；这里按 Kind 兜底转换为对应 Value，
// 无法识别的类型返回 NullValue。
func fromReflect(v any) Value {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return NullValue
	}
	switch rv.Kind() {
	case reflect.String:
		return &Text{text: rv.String()}
	case reflect.Bool:
		return NewBool(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewInt(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewInt(int64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		return NewNumber(rv.Float())
	default:
		return NewAny(v)
	}
}

// toAny 将 Value 还原为原生 Go 值（ToMap 的递归辅助函数）。
func toAny(v Value) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case *Object:
		return val.ToMap()
	case *Array:
		out := make([]any, 0, len(val.data))
		for _, item := range val.data {
			out = append(out, toAny(item))
		}
		return out
	case *Text:
		return val.text
	case *Number:
		if val.isFloat {
			return val.f
		}
		return val.i
	case *Bool:
		return val.b
	case *Any:
		return val.value
	default:
		return nil
	}
}

type Any struct {
	ValueBase
	value any
}

func NewAny(value any) *Any {
	return &Any{
		value: value,
	}
}

func (a *Any) IsAny() bool { return true }

func (a *Any) AsAny() *Any { return a }

func (a *Any) Value() any { return a.value }

func (a *Any) String() string {
	return fmt.Sprintf("%v", a.value)
}

func (a *Any) ToJSON() json.RawMessage {
	data, _ := json.Marshal(a.value)
	return data
}

func (a *Any) MarshalJSON() ([]byte, error) { return a.ToJSON(), nil }

func (a *Any) Unmarshal(v any, opts ...DecoderConfigOption) error {
	return json.Unmarshal(a.ToJSON(), v)
}

func (a *Any) Equal(other Value) bool {
	if o, ok := other.(*Any); ok {
		return reflect.DeepEqual(a.value, o.value)
	}
	return false
}
