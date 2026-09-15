package value

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type Number struct {
	ValueBase
	i       int64
	f       float64
	isFloat bool
}

func (n *Number) IsNumber() bool { return true }

func (n *Number) AsNumber() *Number { return n }

func (n *Number) IsFloat() bool { return n.isFloat }

// Int64 返回整数值。若为浮点数则截断小数部分；
// 超出 int64 范围或为 NaN 时返回 0，避免 int64() 的实现相关结果
// （amd64 上 int64(1e30) 会得到 MinInt64）。
func (n *Number) Int64() int64 {
	if n.isFloat {
		if math.IsNaN(n.f) || n.f >= int64Upper || n.f < math.MinInt64 {
			return 0
		}
		return int64(n.f)
	}
	return n.i
}

// Float64 返回浮点值。若为整数则转换为 float64。
func (n *Number) Float64() float64 {
	if n.isFloat {
		return n.f
	}
	return float64(n.i)
}

func (n *Number) String() string {
	if n.isFloat {
		return fmt.Sprintf("%v", n.f)
	}
	return fmt.Sprintf("%d", n.i)
}

func (n *Number) ToJSON() json.RawMessage {
	if n.isFloat {
		data, _ := json.Marshal(n.f)
		return data
	}
	data, _ := json.Marshal(n.i)
	return data
}

func (n *Number) MarshalJSON() ([]byte, error) { return n.ToJSON(), nil }

func (n *Number) Unmarshal(v any, opts ...DecoderConfigOption) error {
	return json.Unmarshal(n.ToJSON(), v)
}

func (n *Number) Equal(other Value) bool {
	o, ok := other.(*Number)
	if !ok {
		return false
	}
	if n.isFloat != o.isFloat {
		return false
	}
	if n.isFloat {
		return n.f == o.f
	}
	return n.i == o.i
}

// NewNumber 从 float64 创建浮点数值。
func NewNumber(f float64) *Number {
	return &Number{f: f, isFloat: true}
}

// ToNumberE 把任意输入转换为 Number：整数形态优先走 int64 保精度，
// 带小数或超出 int64 范围的走 float64；无法识别的类型或文本返回错误。
func ToNumberE(val any) (*Number, error) {
	switch v := val.(type) {
	case float32:
		return NewNumber(float64(v)), nil
	case float64:
		// JSON 数字解码后就是 float64，整数值也保持浮点（与 fromInterface 一致）。
		return NewNumber(v), nil
	default:
		// 其余全部复用 decoder 的 toInt64 / toFloat64：int/int8/…/uint64、
		// string、json.Number 及具名数值类型它俩都已覆盖，不在这里另写一套解析。
		if i, err := toInt64(val); err == nil {
			return NewInt(i), nil
		}
		f, err := toFloat64(val)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %v (%T) to Number", val, val)
		}
		// "10.0"、"1e5" 这类整数值文本仍返回整型 Number（与 JSON 数字路径一致）；
		// 恰好 2^63 的保持浮点，否则 int64(f) 会回绕。
		if f == math.Trunc(f) && f >= math.MinInt64 && f < int64Upper {
			return NewInt(int64(f)), nil
		}
		return NewNumber(f), nil
	}
}

// NewInt 从 int64 创建整数值。
func NewInt(i int64) *Number {
	return &Number{i: i}
}

// newUintValue 把无符号整数转成 Number。32 位及以下的类型放得下 int64，
// 只有 uint/uint64 在 64 位平台上可能超过 int64 上限：此时降级为 float64
// （丢精度但保留数量级），而不是 int64() 回绕成负数（上限值会变成 -1）。
// 与 ToNumberE 对超范围输入的处理保持一致。
func newUintValue(u uint64) *Number {
	if u > math.MaxInt64 {
		return NewNumber(float64(u))
	}
	return NewInt(int64(u))
}

// value2Number 把 Value 转成 Number。JSON 数字直接返回；Text 交给 ToNumberE 转换
// （URL query 里的 id 经前端原样转发就是这种形态，与 web.Request 的 query/form
// 参数走同一套解析语义）。转换失败是为了拒绝 "12abc" 这类尾随垃圾，
// 而不是静默取 0；bool / object / array 等非文本类型返回 false。
func value2Number(v Value) (*Number, bool) {
	if v == nil {
		return nil, false
	}
	if v.IsNumber() {
		return v.AsNumber(), true
	}
	if !v.IsText() {
		return nil, false
	}
	s := strings.TrimSpace(v.String())
	if s == "" {
		return nil, false
	}
	// 整数形态（"502"、"10.0"）返回整型 Number，带小数的（"10.50"）保留小数位。
	n, err := ToNumberE(s)
	if err != nil {
		return nil, false
	}
	return n, true
}
