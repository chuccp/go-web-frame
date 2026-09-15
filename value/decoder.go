package value

import (
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// durationType 用于在弱类型转换中单独识别 time.Duration。
var durationType = reflect.TypeOf(time.Duration(0))

var (
	jsonUnmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
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

		// 匿名内嵌结构体：把提升字段摊平到本层（与 encoding/json 一致），
		// 因此 struct{ Base; Name string } 可直接匹配 {"id":1,"name":"x"}。
		if field.Anonymous && !hasTag(field, cfg) {
			switch field.Type.Kind() {
			case reflect.Struct:
				if err := decodeStruct(data, fv, cfg); err != nil {
					return err
				}
				continue
			case reflect.Pointer:
				if field.Type.Elem().Kind() == reflect.Struct {
					// 没有任何 key 命中时不分配，保持 nil（与 encoding/json 一致）
					if !hasMatchingKey(data, field.Type.Elem(), cfg) {
						continue
					}
					if fv.IsNil() {
						fv.Set(reflect.New(field.Type.Elem()))
					}
					if err := decodeStruct(data, fv.Elem(), cfg); err != nil {
						return err
					}
					continue
				}
			}
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

// hasMatchingKey 判断 data 中是否存在能命中 t 字段的 key，
// 用于避免在没有字段可解时凭空分配内嵌指针结构体。
func hasMatchingKey(data map[string]any, t reflect.Type, cfg *DecoderConfig) bool {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		if field.Anonymous && !hasTag(field, cfg) && field.Type.Kind() == reflect.Struct {
			if hasMatchingKey(data, field.Type, cfg) {
				return true
			}
			continue
		}
		if _, ok := lookupKey(data, resolveFieldKeys(field, cfg.TagName, cfg.MatchFieldName)); ok {
			return true
		}
	}
	return false
}

// hasTag 判断字段是否带有任一配置的 tag（用于区分"内嵌结构体"与"普通具名字段"）。
func hasTag(field reflect.StructField, cfg *DecoderConfig) bool {
	for _, tagName := range cfg.TagName {
		if field.Tag.Get(tagName) != "" {
			return true
		}
	}
	return false
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
			lower: strings.ToLower(key),
			snake: snake,
			camel: snakeToCamel(key),
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

	// 指针字段：分配新元素后递归解码，成功后写回，因此 *int / *bool / *string / *struct / *slice 等均可用。
	// 解码失败时不修改原字段（val 为 nil 的情况已在上面处理）。
	if fv.Kind() == reflect.Pointer {
		elem := reflect.New(fv.Type().Elem())
		if err := setField(elem.Elem(), val, cfg); err != nil {
			return err
		}
		fv.Set(elem)
		return nil
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
		mapKey, err := convertMapKey(k, keyType)
		if err != nil {
			return err
		}
		mapVal := reflect.New(valType).Elem()
		if err := setField(mapVal, v, cfg); err != nil {
			return err
		}
		fv.SetMapIndex(mapKey, mapVal)
	}
	return nil
}

// convertMapKey 将 map 数据的字符串 key 转换为目标 map 的 key 类型。
// string 之外的类型（如 map[int]T）按 key 类型解析，无法解析时返回错误而非 panic。
func convertMapKey(k string, keyType reflect.Type) (reflect.Value, error) {
	kv := reflect.ValueOf(k)
	if kv.Type().AssignableTo(keyType) {
		return kv, nil
	}
	invalid := func() (reflect.Value, error) {
		return reflect.Value{}, fmt.Errorf("cannot convert map key %q to %s", k, keyType)
	}
	switch keyType.Kind() {
	case reflect.String:
		// 命名 string 类型（如 type Status string）
		return kv.Convert(keyType), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(k, 10, keyType.Bits())
		if err != nil {
			return invalid()
		}
		return reflect.ValueOf(n).Convert(keyType), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(k, 10, keyType.Bits())
		if err != nil {
			return invalid()
		}
		return reflect.ValueOf(n).Convert(keyType), nil
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(k, keyType.Bits())
		if err != nil {
			return invalid()
		}
		return reflect.ValueOf(n).Convert(keyType), nil
	case reflect.Bool:
		b, err := strconv.ParseBool(k)
		if err != nil {
			return invalid()
		}
		return reflect.ValueOf(b).Convert(keyType), nil
	}
	return invalid()
}

func assignValue(fv reflect.Value, val any, cfg *DecoderConfig) error {
	rv := reflect.ValueOf(val)

	// 整数 → string 在 Go 中是合法转换，但结果是 rune（65 → "A"），
	// 几乎不会是调用方想要的，因此单独按十进制格式化，避免静默变成字符。
	if fv.Kind() == reflect.String && isIntegerKind(rv.Kind()) {
		if !cfg.WeaklyTypedInput {
			return fmt.Errorf("cannot assign %T to %s", val, fv.Type())
		}
		fv.SetString(fmt.Sprintf("%v", val))
		return nil
	}

	if rv.Type().AssignableTo(fv.Type()) {
		fv.Set(rv)
		return nil
	}

	// 目标类型实现了 json.Unmarshaler / encoding.TextUnmarshaler（time.Time、
	// uuid.UUID、自定义枚举等）：交给类型自己的解析逻辑，不走弱类型转换。
	if handled, err := tryUnmarshaler(fv, val); handled {
		return err
	}

	// 数值类型之间（含 JSON 数字 float64 → int 的常规路径）统一走带范围检查的转换，
	// 避免 Convert 静默回绕：-5 → uint 巨值、300 → uint8 得 44、1e40 → float32 得 +Inf。
	if isNumericKind(rv.Kind()) && isNumericKind(fv.Kind()) {
		return convertNumber(fv, val)
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
	// time.Duration 需单独处理：否则 "30s" 会被当普通整数解析成 30 纳秒。
	if fv.Type() == durationType {
		if s, ok := val.(string); ok {
			d, err := time.ParseDuration(strings.TrimSpace(s))
			if err != nil {
				return fmt.Errorf("cannot parse %q as time.Duration", s)
			}
			fv.SetInt(int64(d))
			return nil
		}
	}
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(fmt.Sprintf("%v", val))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return convertNumber(fv, val)
	case reflect.Bool:
		b, err := toBool(val)
		if err != nil {
			return err
		}
		fv.SetBool(b)
	default:
		return fmt.Errorf("cannot convert %T to %s", val, fv.Type())
	}
	return nil
}

// convertNumber 在数值类型之间转换，越界时返回错误而不是静默回绕。
func convertNumber(fv reflect.Value, val any) error {
	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := toInt64(val)
		if err != nil {
			return err
		}
		if fv.OverflowInt(n) {
			return fmt.Errorf("cannot convert %v to %s (out of range)", val, fv.Type())
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := toUint64(val)
		if err != nil {
			return err
		}
		if fv.OverflowUint(n) {
			return fmt.Errorf("cannot convert %v to %s (out of range)", val, fv.Type())
		}
		fv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := toFloat64(val)
		if err != nil {
			return err
		}
		if fv.OverflowFloat(n) {
			return fmt.Errorf("cannot convert %v to %s (out of range)", val, fv.Type())
		}
		fv.SetFloat(n)
	default:
		return fmt.Errorf("cannot convert %T to %s", val, fv.Type())
	}
	return nil
}

// tryUnmarshaler 优先使用目标类型自身的解析逻辑（json.Unmarshaler、
// encoding.TextUnmarshaler），覆盖 time.Time / uuid.UUID / decimal / 自定义枚举等。
// handled 为 false 表示类型未实现这两个接口，调用方继续走常规转换。
func tryUnmarshaler(fv reflect.Value, val any) (handled bool, err error) {
	// 未命名类型（int/string/[]T/map 等）不可能有方法，快速跳过，避免热路径开销。
	if fv.Type().Name() == "" || !fv.CanAddr() {
		return false, nil
	}
	ptr := fv.Addr()
	if ptr.Type().Implements(jsonUnmarshalerType) {
		raw, err := json.Marshal(val)
		if err != nil {
			return true, err
		}
		return true, ptr.Interface().(json.Unmarshaler).UnmarshalJSON(raw)
	}
	if s, ok := val.(string); ok && ptr.Type().Implements(textUnmarshalerType) {
		return true, ptr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
	}
	return false, nil
}

// isIntegerKind 判断类型是否为（有符号或无符号）整数。
func isIntegerKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	return false
}

// isNumericKind 判断类型是否为整数或浮点数。
func isNumericKind(k reflect.Kind) bool {
	return isIntegerKind(k) || k == reflect.Float32 || k == reflect.Float64
}

// int64Upper / uint64Upper 是两个上界的浮点表示（2^63 / 2^64）。
// math.MaxInt64、math.MaxUint64 转成 float64 时会进位到这两个值，
// 若用 v > math.MaxInt64 判断，恰好等于上界的输入会被放过，随后的
// int64()/uint64() 转换结果由实现决定（amd64 上是 MinInt64 / 0），
// 因此浮点分支的边界必须写成 >=。
const (
	int64Upper  = float64(1 << 63)
	uint64Upper = float64(1 << 64)
)

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
		// 32 位平台上 uint 放得下，64 位平台上可能超过 int64 上限。
		if v > math.MaxInt64 {
			return 0, fmt.Errorf("cannot convert %d to int64 (out of range)", v)
		}
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if v > math.MaxInt64 {
			return 0, fmt.Errorf("cannot convert %d to int64 (out of range)", v)
		}
		return int64(v), nil
	case float32:
		return toInt64(float64(v))
	case float64:
		if math.IsNaN(v) || v >= int64Upper || v < math.MinInt64 {
			return 0, fmt.Errorf("cannot convert %v to int64 (out of range)", v)
		}
		// 带小数的值不静默截断："2.9" 写进 int 字段会让用户以为存的是 2.9。
		if v != math.Trunc(v) {
			return 0, fmt.Errorf("cannot convert %v to int64 (would lose decimal part)", v)
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		return strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	default:
		// 具名数值类型（如 type UserID int64、time.Duration）按其底层类型处理
		if n, ok := underlyingNumber(val); ok {
			return toInt64(n)
		}
		return 0, fmt.Errorf("cannot convert %T to int64", val)
	}
}

func toUint64(val any) (uint64, error) {
	switch v := val.(type) {
	case int:
		return uint64FromInt64(int64(v))
	case int8:
		return uint64FromInt64(int64(v))
	case int16:
		return uint64FromInt64(int64(v))
	case int32:
		return uint64FromInt64(int64(v))
	case int64:
		return uint64FromInt64(v)
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case float32:
		return toUint64(float64(v))
	case float64:
		if math.IsNaN(v) {
			return 0, fmt.Errorf("cannot convert %v to uint64", v)
		}
		if v < 0 {
			return 0, fmt.Errorf("cannot convert %v to unsigned (negative value)", v)
		}
		if v >= uint64Upper {
			return 0, fmt.Errorf("cannot convert %v to unsigned (out of range)", v)
		}
		// 同 toInt64：带小数的值不静默截断。
		if v != math.Trunc(v) {
			return 0, fmt.Errorf("cannot convert %v to uint64 (would lose decimal part)", v)
		}
		return uint64(v), nil
	case json.Number:
		return strconv.ParseUint(v.String(), 10, 64)
	case string:
		return strconv.ParseUint(strings.TrimSpace(v), 10, 64)
	default:
		// 具名数值类型按其底层类型处理
		if n, ok := underlyingNumber(val); ok {
			return toUint64(n)
		}
		return 0, fmt.Errorf("cannot convert %T to uint64", val)
	}
}

func uint64FromInt64(n int64) (uint64, error) {
	if n < 0 {
		return 0, fmt.Errorf("cannot convert %d to uint64 (negative)", n)
	}
	return uint64(n), nil
}

// underlyingNumber 把具名数值类型（如 type UserID int64、time.Duration）
// 转换为对应的内建数值类型，ok 为 false 表示不是数值类型。
func underlyingNumber(val any) (any, bool) {
	rv := reflect.ValueOf(val)
	if !isNumericKind(rv.Kind()) {
		return nil, false
	}
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint(), true
	default:
		return rv.Float(), true
	}
}

// toBool 将输入转换为布尔值：字符串支持 true/false、yes/no、on/off、y/n、1/0（忽略大小写），
// 数字按非 0 为 true；其余类型报错。
func toBool(val any) (bool, error) {
	switch v := val.(type) {
	case bool:
		return v, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "t", "true", "yes", "y", "on":
			return true, nil
		case "0", "f", "false", "no", "n", "off", "":
			return false, nil
		}
		return false, fmt.Errorf("cannot convert %q to bool", v)
	default:
		n, err := toFloat64(val)
		if err != nil {
			return false, fmt.Errorf("cannot convert %T to bool", val)
		}
		return n != 0, nil
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
		return strconv.ParseFloat(strings.TrimSpace(v), 64)
	default:
		// 具名数值类型按其底层类型处理
		if n, ok := underlyingNumber(val); ok {
			return toFloat64(n)
		}
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}
