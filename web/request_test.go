package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// TestQueryFloat64 覆盖 QueryFloat64 的解析语义：与同组 QueryInt/QueryUint 一致，
// 走 cast，非法或缺失的输入静默返回 0，不报错。
func TestQueryFloat64(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.Get("/q", func(r *Request) (any, error) {
			return map[string]any{
				"price":   r.QueryFloat64("price"),
				"count":   r.QueryFloat64("count"),
				"neg":     r.QueryFloat64("neg"),
				"sci":     r.QueryFloat64("sci"),
				"missing": r.QueryFloat64("missing"),
				"empty":   r.QueryFloat64("empty"),
				"garbage": r.QueryFloat64("garbage"),
				"trail":   r.QueryFloat64("trail"),
				// NaN/Inf 会被 strconv 接受，先转成字符串对比，避免 JSON 序列化失败
				"nan": strconv.FormatFloat(r.QueryFloat64("nan"), 'g', -1, 64),
				"inf": strconv.FormatFloat(r.QueryFloat64("inf"), 'g', -1, 64),
			}, nil
		})
	})

	resp, err := http.Get(ts.URL + "/q?price=3.14&count=10&neg=-2.5&sci=1e3&empty=&garbage=abc&trail=12abc&nan=NaN&inf=Inf")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var msg struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &msg); err != nil {
		t.Fatalf("unmarshal %q: %v", body, err)
	}

	want := map[string]float64{
		"price":   3.14,
		"count":   10,
		"neg":     -2.5,
		"sci":     1000,
		"missing": 0,
		"empty":   0,
		"garbage": 0,
		"trail":   0,
	}
	for key, w := range want {
		if got := msg.Data[key].(float64); got != w {
			t.Errorf("QueryFloat64(%q) = %v, want %v", key, got, w)
		}
	}

	if got := msg.Data["nan"].(string); got != "NaN" {
		t.Errorf("QueryFloat64(\"nan\") = %v, want NaN", got)
	}
	if got := msg.Data["inf"].(string); got != "+Inf" {
		t.Errorf("QueryFloat64(\"inf\") = %v, want +Inf", got)
	}
}

// TestGetJsonFloat64Value 覆盖 JSON body 的 float64 取数：GetJsonFloat64Value
// 对缺失/非法值静默返回 0，只在 body 不是合法 JSON 时报错；
// GetJsonFloat64ValueOrDefault 在同样情况下回退到默认值。
func TestGetJsonFloat64Value(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.Post("/j", func(r *Request) (any, error) {
			price, err := r.GetJsonFloat64Value("price")
			if err != nil {
				return nil, err
			}
			count, _ := r.GetJsonFloat64Value("count")
			garbage, _ := r.GetJsonFloat64Value("garbage")
			missing, _ := r.GetJsonFloat64Value("missing")
			return map[string]any{
				"price":            price,
				"count":            count,
				"garbage":          garbage,
				"missing":          missing,
				"priceOrDefault":   r.GetJsonFloat64ValueOrDefault("price", -1),
				"missingOrDefault": r.GetJsonFloat64ValueOrDefault("missing", -1),
				"garbageOrDefault": r.GetJsonFloat64ValueOrDefault("garbage", -1),
			}, nil
		})
	})

	resp, err := http.Post(ts.URL+"/j", "application/json",
		strings.NewReader(`{"price":"3.14","count":10,"garbage":"abc"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var msg struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &msg); err != nil {
		t.Fatalf("unmarshal %q: %v", body, err)
	}

	want := map[string]float64{
		"price":            3.14, // 数字字符串也解析
		"count":            10,
		"garbage":          0,
		"missing":          0,
		"priceOrDefault":   3.14,
		"missingOrDefault": -1,
		"garbageOrDefault": -1,
	}
	for key, w := range want {
		if got := msg.Data[key].(float64); got != w {
			t.Errorf("%s = %v, want %v", key, got, w)
		}
	}

	// body 不是合法 JSON 时，GetJsonFloat64Value 报错，OrDefault 回退默认值
	badResp, err := http.Post(ts.URL+"/j", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatal(err)
	}
	defer badResp.Body.Close()
	if badResp.StatusCode == http.StatusOK {
		t.Errorf("非法 JSON body 应返回错误状态码, got %d", badResp.StatusCode)
	}
}
