package value

import (
	"encoding/json"
	"testing"
	"time"
)

type benchItem struct {
	UserName string  `json:"user_name"`
	Age      int     `json:"age"`
	Active   bool    `json:"active"`
	Score    float64 `json:"score"`
}

type benchConfig struct {
	Items        []benchItem   `json:"items"`
	MaxOpenConns int           `json:"max_open_conns"`
	Timeout      time.Duration `json:"timeout"`
	Name         string        `json:"name"`
}

type benchFlat struct {
	Name         string  `json:"name"`
	Age          int     `json:"age"`
	MaxOpenConns int     `json:"max_open_conns"`
	Score        float64 `json:"score"`
	Active       bool    `json:"active"`
}

var benchNested = []byte(`{"items":[{"user_name":"alice","age":30,"active":true,"score":1.5},{"user_name":"bob","age":25,"active":false,"score":2.5}],"max_open_conns":10,"timeout":"30s","name":"x"}`)
var benchFlatData = []byte(`{"name":"alice","age":30,"max_open_conns":10,"score":1.5,"active":true}`)

func BenchmarkDecodeNested(b *testing.B) {
	obj, _ := ParseJSON(benchNested)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg benchConfig
		if err := obj.Unmarshal(&cfg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeFlat(b *testing.B) {
	obj, _ := ParseJSON(benchFlatData)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s benchFlat
		if err := obj.Unmarshal(&s); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStdJSON 作为参照：同样结构体走 encoding/json 的成本。
func BenchmarkStdJSONFlat(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s benchFlat
		if err := json.Unmarshal(benchFlatData, &s); err != nil {
			b.Fatal(err)
		}
	}
}
