package value

import (
	"encoding/json"
	"math"
	"testing"
)

func TestToNumberE(t *testing.T) {
	tests := []struct {
		name    string
		in      any
		wantErr bool
		isFloat bool
		wantI   int64
		wantF   float64
	}{
		{"整数文本", "502", false, false, 502, 502},
		{"两端空格", " 8 ", false, false, 8, 8},
		{"负整数文本", "-3", false, false, -3, -3},
		{"整数形态的浮点文本", "10.0", false, false, 10, 10},
		{"科学计数法文本", "1e5", false, false, 100000, 100000},
		{"带小数的文本", "10.50", false, true, 10, 10.5},
		{"int64 上限文本", "9223372036854775807", false, false, math.MaxInt64, float64(math.MaxInt64)},
		{"超出 int64 的文本", "9223372036854775808", false, true, math.MinInt64, float64(1 << 63)},
		{"尾随垃圾", "12abc", true, false, 0, 0},
		{"空串", "", true, false, 0, 0},
		{"bool", true, true, false, 0, 0},
		{"nil", nil, true, false, 0, 0},

		{"int", 42, false, false, 42, 42},
		{"int64", int64(-7), false, false, -7, -7},
		{"float32", float32(1.5), false, true, 1, 1.5},
		{"float64", 2.5, false, true, 2, 2.5},
		{"uint64 上限", uint64(math.MaxUint64), false, true, 0, float64(1 << 64)},

		{"json.Number 整数", json.Number("42"), false, false, 42, 42},
		{"json.Number 小数", json.Number("2.9"), false, true, 2, 2.9},
		{"json.Number 非法", json.Number("abc"), true, false, 0, 0},
		{"json.Number 超 int64", json.Number("9223372036854775808"), false, true, 0, float64(1 << 63)},

		{"具名整数类型", namedInt(9), false, false, 9, 9},
		{"具名浮点类型", namedFloat(2.5), false, true, 2, 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := ToNumberE(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("%#v 应报错, got %v", tt.in, n)
				}
				return
			}
			if err != nil {
				t.Fatalf("%#v 不应报错: %v", tt.in, err)
			}
			if n.IsFloat() != tt.isFloat {
				t.Errorf("IsFloat = %v, want %v", n.IsFloat(), tt.isFloat)
			}
			if tt.isFloat {
				if n.Float64() != tt.wantF {
					t.Errorf("Float64 = %v, want %v", n.Float64(), tt.wantF)
				}
				return
			}
			if n.Int64() != tt.wantI {
				t.Errorf("Int64 = %v, want %v", n.Int64(), tt.wantI)
			}
		})
	}
}

type namedInt int64

type namedFloat float32

// TestNumericBoundaries 浮点→整数的边界：恰好 2^63 / 2^64 的输入
// 若用 v > math.MaxInt64 判断会被放过，转换结果由实现决定（amd64 上是
// MinInt64），因此必须报错而不是静默回绕。
func TestNumericBoundaries(t *testing.T) {
	if v, err := toInt64(float64(1 << 63)); err == nil {
		t.Errorf("2^63 → int64 应报错, got %d", v)
	}
	if v, err := toInt64(float64(-1 << 63)); err != nil || v != math.MinInt64 {
		t.Errorf("-2^63 应正常转换, got %d %v", v, err)
	}
	if v, err := toUint64(float64(1 << 64)); err == nil {
		t.Errorf("2^64 → uint64 应报错, got %d", v)
	}
	// uint64 上限本身必须保真（不经过 float64）
	if v, err := toUint64(uint64(math.MaxUint64)); err != nil || v != math.MaxUint64 {
		t.Errorf("uint64 上限应保真, got %d %v", v, err)
	}
	if v, err := toInt64(uint(math.MaxUint64)); err == nil {
		t.Errorf("超 int64 上限的 uint 应报错, got %d", v)
	}
}

// TestNumberGetterBoundaries Number 取值层的同类回绕：越界浮点经 int64() 在
// amd64 上会得到 MinInt64，负数经 uint() 会变成巨大值，都必须挡住。
func TestNumberGetterBoundaries(t *testing.T) {
	if got := NewNumber(1e30).Int64(); got != 0 {
		t.Errorf("Int64(1e30) = %d, want 0", got)
	}
	if got := NewNumber(math.NaN()).Int64(); got != 0 {
		t.Errorf("Int64(NaN) = %d, want 0", got)
	}

	o := NewObject()
	o.PutAny("big", 1e30)
	if got := o.GetInt("big"); got != 0 {
		t.Errorf("GetInt(1e30) = %d, want 0", got)
	}
	o.PutAny("neg", -3)
	if got := o.GetUint("neg"); got != 0 {
		t.Errorf("GetUint(-3) = %d, want 0（回绕成 18446744073709551613）", got)
	}
	// uint64 上限：Number 装不下，降级 float64 保留数量级，不变成 -1
	o.PutAny("u", uint64(math.MaxUint64))
	if got := o.GetInt("u"); got != 0 {
		t.Errorf("GetInt(uint64 上限) = %d, want 0（回绕成 -1）", got)
	}
	if got := o.GetNumber("u"); got != float64(1<<64) {
		t.Errorf("GetNumber(uint64 上限) = %v, want %v", got, float64(1<<64))
	}

	// JSON 路径（大数、负数同源）
	v, err := ParseJSON([]byte(`{"n":1e30,"m":-3}`))
	if err != nil {
		t.Fatal(err)
	}
	obj := v.AsObject()
	if got := obj.GetInt("n"); got != 0 {
		t.Errorf("JSON 1e30 GetInt = %d, want 0", got)
	}
	if got := obj.GetUint("m"); got != 0 {
		t.Errorf("JSON -3 GetUint = %d, want 0", got)
	}
}
