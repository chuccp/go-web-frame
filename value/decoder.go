package value

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// DecoderConfig 保存 Unmarshal 的解码配置。
type DecoderConfig struct {
	// TagName 指定用于匹配 struct 字段的 tag 名称，默认 "json"。
	TagName []string
	// WeaklyTypedInput 启用弱类型输入，允许字符串到数字、布尔值等的自动转换。
	WeaklyTypedInput bool
	// MatchFieldName 启用字段名匹配（忽略 tag），支持 snake_case 到 CamelCase 的自动转换和大小写忽略。
	// 例如 map key "max_open_conns" 可匹配字段 "MaxOpenConns"。
	MatchFieldName bool
}

// DecoderConfigOption 用于配置 Unmarshal 的解码行为。
type DecoderConfigOption func(*DecoderConfig)

// WithTagName 设置解码时使用的 struct tag 名称，默认使用 "json" tag。
func WithTagName(tag ...string) DecoderConfigOption {
	return func(c *DecoderConfig) {
		c.TagName = tag
	}
}

// WithWeaklyTypedInput 启用弱类型输入模式，允许字符串到数字、布尔值等的自动转换。
func WithWeaklyTypedInput(enabled bool) DecoderConfigOption {
	return func(c *DecoderConfig) {
		c.WeaklyTypedInput = enabled
	}
}

// WithMatchFieldName 启用字段名匹配模式，支持 snake_case 到 CamelCase 的自动转换和大小写忽略。
// 启用后，map key "max_open_conns" 可自动匹配字段 "MaxOpenConns"，无需 struct tag。
func WithMatchFieldName(enabled bool) DecoderConfigOption {
	return func(c *DecoderConfig) {
		c.MatchFieldName = enabled
	}
}

func newDecoderConfig(opts ...DecoderConfigOption) *DecoderConfig {
	c := &DecoderConfig{
		TagName:          []string{"json"},
		WeaklyTypedInput: true,
		MatchFieldName:   true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// decodeValue 使用指定 tag 将 map 数据解码到目标结构体。
func decodeValue(data map[string]any, output any, cfg *DecoderConfig) error {
	rv := reflect.ValueOf(output)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("value: Unmarshal requires a non-nil pointer")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, output)
	}
	return decodeStruct(data, rv, cfg)
}

func decodeStruct(data map[string]any, rv reflect.Value, cfg *DecoderConfig) error {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)

		if !fv.CanSet() {
			continue
		}

		keys := resolveFieldKeys(field, cfg.TagName, cfg.MatchFieldName)
		if len(keys) == 0 {
			continue
		}

		val, ok := lookupKey(data, keys)
		if !ok {
			continue
		}

		if err := setField(fv, val, cfg); err != nil {
			return fmt.Errorf("value: field %s: %w", field.Name, err)
		}
	}
	return nil
}

// resolveFieldKeys 返回所有候选的 key 列表：先按 tag 顺序，再按字段名匹配。
func resolveFieldKeys(field reflect.StructField, tagNames []string, matchFieldName bool) []string {
	var keys []string
	for _, tagName := range tagNames {
		tag := field.Tag.Get(tagName)
		key := parseTagKey(tag)
		if key == "-" {
			return nil
		}
		if key != "" {
			keys = append(keys, key)
		}
	}
	if matchFieldName {
		keys = append(keys, field.Name)
	}
	return keys
}

// lookupKey 依次尝试候选 key，优先精确匹配，再忽略大小写匹配，最后 snake/camel 转换匹配。
func lookupKey(data map[string]any, keys []string) (any, bool) {
	// 精确匹配
	for _, key := range keys {
		if val, ok := data[key]; ok {
			return val, true
		}
	}
	// 构建所有候选的变体（小写、snake、camel），一次遍历 data 完成匹配
	type variant struct{ lower, snake, camel string }
	variants := make([]variant, len(keys))
	for i, key := range keys {
		snake := camelToSnake(key)
		variants[i] = variant{
			lower:  strings.ToLower(key),
			snake:  snake,
			camel:  snakeToCamel(key),
		}
		// 补充 snake 形式的小写（如 MaxOpenConns → max_open_conns）
		if snake != variants[i].lower {
			variants = append(variants, variant{lower: strings.ToLower(snake), snake: snake})
		}
	}
	for k, val := range data {
		lowerK := strings.ToLower(k)
		for _, v := range variants {
			if lowerK == v.lower || lowerK == v.snake || k == v.snake || k == v.camel {
				return val, true
			}
		}
	}
	return nil, false
}

// snakeToCamel 将 snake_case 转换为 CamelCase（如 max_open_conns → MaxOpenConns）。
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// camelToSnake 将 CamelCase 转换为 snake_case（如 MaxOpenConns → max_open_conns）。
func camelToSnake(s string) string {
	var result []byte
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, byte(r-'A'+'a'))
		} else {
			result = append(result, byte(r))
		}
	}
	return string(result)
}

func parseTagKey(tag string) string {
	if idx := strings.IndexByte(tag, ','); idx >= 0 {
		return tag[:idx]
	}
	return tag
}

func setField(fv reflect.Value, val any, cfg *DecoderConfig) error {
	if val == nil {
		fv.Set(reflect.Zero(fv.Type()))
		return nil
	}

	if fv.Kind() == reflect.Struct {
		if m, ok := val.(map[string]any); ok {
			return decodeStruct(m, fv, cfg)
		}
	}

	if fv.Kind() == reflect.Pointer && fv.Type().Elem().Kind() == reflect.Struct {
		if m, ok := val.(map[string]any); ok {
			if fv.IsNil() {
				fv.Set(reflect.New(fv.Type().Elem()))
			}
			return decodeStruct(m, fv.Elem(), cfg)
		}
	}

	if fv.Kind() == reflect.Slice {
		if arr, ok := val.([]any); ok {
			return setSlice(fv, arr, cfg)
		}
	}

	if fv.Kind() == reflect.Map {
		if m, ok := val.(map[string]any); ok {
			return setMap(fv, m, cfg)
		}
	}

	return assignValue(fv, val, cfg)
}

func setSlice(fv reflect.Value, arr []any, cfg *DecoderConfig) error {
	elemType := fv.Type().Elem()
	slice := reflect.MakeSlice(fv.Type(), len(arr), len(arr))
	for i, item := range arr {
		elem := reflect.New(elemType).Elem()
		if err := setField(elem, item, cfg); err != nil {
			return err
		}
		slice.Index(i).Set(elem)
	}
	fv.Set(slice)
	return nil
}

func setMap(fv reflect.Value, m map[string]any, cfg *DecoderConfig) error {
	mapType := fv.Type()
	if fv.IsNil() {
		fv.Set(reflect.MakeMap(mapType))
	}
	keyType := mapType.Key()
	valType := mapType.Elem()
	for k, v := range m {
		mapKey := reflect.ValueOf(k).Convert(keyType)
		mapVal := reflect.New(valType).Elem()
		if err := setField(mapVal, v, cfg); err != nil {
			return err
		}
		fv.SetMapIndex(mapKey, mapVal)
	}
	return nil
}

func assignValue(fv reflect.Value, val any, cfg *DecoderConfig) error {
	rv := reflect.ValueOf(val)

	if rv.Type().AssignableTo(fv.Type()) {
		fv.Set(rv)
		return nil
	}

	if rv.Type().ConvertibleTo(fv.Type()) {
		fv.Set(rv.Convert(fv.Type()))
		return nil
	}

	if cfg.WeaklyTypedInput {
		return weakConvert(fv, val)
	}

	return fmt.Errorf("cannot assign %T to %s", val, fv.Type())
}

func weakConvert(fv reflect.Value, val any) error {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(fmt.Sprintf("%v", val))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := toInt64(val)
		if err != nil {
			return err
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := toInt64(val)
		if err != nil {
			return err
		}
		fv.SetUint(uint64(n))
	case reflect.Float32, reflect.Float64:
		n, err := toFloat64(val)
		if err != nil {
			return err
		}
		fv.SetFloat(n)
	case reflect.Bool:
		switch v := val.(type) {
		case bool:
			fv.SetBool(v)
		case string:
			fv.SetBool(v == "true" || v == "1" || v == "yes")
		default:
			fv.SetBool(reflect.ValueOf(val).Kind() != reflect.Invalid)
		}
	default:
		return fmt.Errorf("cannot convert %T to %s", val, fv.Type())
	}
	return nil
}

func toInt64(val any) (int64, error) {
	switch v := val.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		var n int64
		_, err := fmt.Sscanf(v, "%d", &n)
		return n, err
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", val)
	}
}

func toFloat64(val any) (float64, error) {
	switch v := val.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case json.Number:
		return v.Float64()
	case string:
		var n float64
		_, err := fmt.Sscanf(v, "%f", &n)
		return n, err
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}
