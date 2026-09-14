package value

import (
	"math"
	"testing"
	"time"
)

// --- snakeToCamel / camelToSnake ---

func TestSnakeToCamel(t *testing.T) {
	tests := []struct{ in, want string }{
		{"max_open_conns", "MaxOpenConns"},
		{"filePath", "FilePath"},
		{"name", "Name"},
		{"conn_max_lifetime", "ConnMaxLifetime"},
		{"", ""},
		{"a", "A"},
	}
	for _, tt := range tests {
		got := snakeToCamel(tt.in)
		if got != tt.want {
			t.Errorf("snakeToCamel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCamelToSnake(t *testing.T) {
	tests := []struct{ in, want string }{
		{"MaxOpenConns", "max_open_conns"},
		{"FilePath", "file_path"},
		{"Name", "name"},
		{"ConnMaxLifetime", "conn_max_lifetime"},
		{"", ""},
		{"A", "a"},
	}
	for _, tt := range tests {
		got := camelToSnake(tt.in)
		if got != tt.want {
			t.Errorf("camelToSnake(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// --- Object.Unmarshal with json tag ---

func TestUnmarshal_JsonTag(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"user_name":"alice","age":30}`))
	type User struct {
		Name string `json:"user_name"`
		Age  int    `json:"age"`
	}
	var u User
	err := obj.Unmarshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "alice" || u.Age != 30 {
		t.Errorf("got %+v", u)
	}
}

// --- Object.Unmarshal with multiple tags ---

func TestUnmarshal_MultipleTags(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"name":"bob"}`))
	type User struct {
		Name string `yaml:"name" json:"name"`
	}
	var u User
	err := obj.Unmarshal(&u, WithTagName("yaml", "json"))
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "bob" {
		t.Errorf("got %+v", u)
	}
}

// --- Object.Unmarshal with snake_case field name match (no tag) ---

func TestUnmarshal_SnakeCaseFieldName(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"max_open_conns":10,"max_idle_conns":5,"conn_max_lifetime":3600}`))
	type Config struct {
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
	}
	var cfg Config
	err := obj.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxOpenConns != 10 || cfg.MaxIdleConns != 5 || cfg.ConnMaxLifetime != 3600 {
		t.Errorf("got %+v", cfg)
	}
}

// --- Object.Unmarshal with CamelCase key matching snake_case field ---

func TestUnmarshal_CamelCaseKeyMatchSnakeField(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"MaxOpenConns":20}`))
	type Config struct {
		MaxOpenConns int
	}
	var cfg Config
	err := obj.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxOpenConns != 20 {
		t.Errorf("got %+v", cfg)
	}
}

// --- Object.Unmarshal with case insensitive ---

func TestUnmarshal_CaseInsensitive(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"MAX_OPEN_CONNS":15}`))
	type Config struct {
		MaxOpenConns int
	}
	var cfg Config
	err := obj.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxOpenConns != 15 {
		t.Errorf("got %+v", cfg)
	}
}

// --- Object.Unmarshal with WeaklyTypedInput ---

func TestUnmarshal_WeaklyTypedInput(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"name":"alice","age":"25","active":"true"}`))
	type User struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}
	var u User
	err := obj.Unmarshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "alice" || u.Age != 25 || !u.Active {
		t.Errorf("got %+v", u)
	}
}

// --- Object.Unmarshal with WeaklyTypedInput disabled ---

func TestUnmarshal_StrictMode(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"age":"not_a_number"}`))
	type User struct {
		Age int `json:"age"`
	}
	var u User
	err := obj.Unmarshal(&u, WithWeaklyTypedInput(false))
	if err == nil {
		t.Error("expected error in strict mode")
	}
}

// --- Object.Unmarshal with nested struct ---

func TestUnmarshal_NestedStruct(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"user":{"name":"alice"},"max_open_conns":10}`))
	type Config struct {
		User struct {
			Name string `json:"name"`
		} `json:"user"`
		MaxOpenConns int
	}
	var cfg Config
	err := obj.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.User.Name != "alice" || cfg.MaxOpenConns != 10 {
		t.Errorf("got %+v", cfg)
	}
}

// --- Object.Unmarshal with slice ---

func TestUnmarshal_Slice(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"tags":["go","web"]}`))
	type Config struct {
		Tags []string `json:"tags"`
	}
	var cfg Config
	err := obj.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "go" || cfg.Tags[1] != "web" {
		t.Errorf("got %+v", cfg)
	}
}

// --- Object.Unmarshal with tag priority ---

func TestUnmarshal_TagPriority(t *testing.T) {
	// json tag 优先于字段名
	obj, _ := ParseJSON([]byte(`{"user_name":"alice","Name":"bob"}`))
	type User struct {
		Name string `json:"user_name"`
	}
	var u User
	err := obj.Unmarshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "alice" {
		t.Errorf("tag should take priority, got %+v", u)
	}
}

// --- Object.Unmarshal with "-" tag ---

func TestUnmarshal_SkipField(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"name":"alice","secret":"hidden"}`))
	type User struct {
		Name   string `json:"name"`
		Secret string `json:"-"`
	}
	var u User
	err := obj.Unmarshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "alice" || u.Secret != "" {
		t.Errorf("got %+v", u)
	}
}

// --- Object.Unmarshal nil ---

func TestUnmarshal_NilObject(t *testing.T) {
	var obj *Object
	type User struct {
		Name string
	}
	var u User
	err := obj.Unmarshal(&u)
	if err != nil {
		t.Errorf("nil object should return nil error, got %v", err)
	}
}

// --- Object.Decode with options ---

func TestDecode_WithOptions(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"max_open_conns":10}`))
	type Config struct {
		MaxOpenConns int
	}
	var cfg Config
	err := obj.AsObject().Decode(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxOpenConns != 10 {
		t.Errorf("got %+v", cfg)
	}
}

// --- Full SQLite-like config ---

func TestUnmarshal_SQLiteConfig(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{
		"filePath": "/tmp/test.db",
		"max_open_conns": 20,
		"max_idle_conns": 10,
		"conn_max_lifetime": 7200
	}`))
	type SQLiteConfig struct {
		FilePath        string `json:"filePath"`
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
	}
	var cfg SQLiteConfig
	err := obj.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FilePath != "/tmp/test.db" {
		t.Errorf("FilePath: got %q", cfg.FilePath)
	}
	if cfg.MaxOpenConns != 20 {
		t.Errorf("MaxOpenConns: got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 10 {
		t.Errorf("MaxIdleConns: got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 7200 {
		t.Errorf("ConnMaxLifetime: got %d", cfg.ConnMaxLifetime)
	}
}

// --- Array.Unmarshal to []string ---

func TestArrayUnmarshal_Strings(t *testing.T) {
	arr, _ := ParseJSON([]byte(`["go","web","frame"]`))
	var out []string
	err := arr.Unmarshal(&out)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 || out[0] != "go" || out[1] != "web" || out[2] != "frame" {
		t.Errorf("got %+v", out)
	}
}

// --- Array.Unmarshal to []int ---

func TestArrayUnmarshal_Ints(t *testing.T) {
	arr, _ := ParseJSON([]byte(`[1,2,3]`))
	var out []int
	err := arr.Unmarshal(&out)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 || out[0] != 1 || out[1] != 2 || out[2] != 3 {
		t.Errorf("got %+v", out)
	}
}

// --- Array.Unmarshal to []float64 ---

func TestArrayUnmarshal_Floats(t *testing.T) {
	arr, _ := ParseJSON([]byte(`[1.5,2.5,3.5]`))
	var out []float64
	err := arr.Unmarshal(&out)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 || out[0] != 1.5 || out[1] != 2.5 || out[2] != 3.5 {
		t.Errorf("got %+v", out)
	}
}

// --- Array.Unmarshal to []struct with snake_case fields ---

func TestArrayUnmarshal_Structs(t *testing.T) {
	arr, _ := ParseJSON([]byte(`[{"user_name":"alice","max_open_conns":10},{"user_name":"bob","max_open_conns":20}]`))
	type Item struct {
		UserName    string `json:"user_name"`
		MaxOpenConns int
	}
	var out []Item
	err := arr.Unmarshal(&out)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}
	if out[0].UserName != "alice" || out[0].MaxOpenConns != 10 {
		t.Errorf("item[0]: got %+v", out[0])
	}
	if out[1].UserName != "bob" || out[1].MaxOpenConns != 20 {
		t.Errorf("item[1]: got %+v", out[1])
	}
}

// --- Array.Unmarshal with weakly typed input ---

func TestArrayUnmarshal_WeaklyTyped(t *testing.T) {
	arr, _ := ParseJSON([]byte(`["10","20","30"]`))
	var out []int
	err := arr.Unmarshal(&out)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 || out[0] != 10 || out[1] != 20 || out[2] != 30 {
		t.Errorf("got %+v", out)
	}
}

// --- Array.Unmarshal with weak type disabled ---

func TestArrayUnmarshal_StrictMode(t *testing.T) {
	arr, _ := ParseJSON([]byte(`["not_a_number"]`))
	var out []int
	err := arr.Unmarshal(&out, WithWeaklyTypedInput(false))
	if err == nil {
		t.Error("expected error in strict mode")
	}
}

// --- Array.Unmarshal nil ---

func TestArrayUnmarshal_Nil(t *testing.T) {
	var arr *Array
	var out []string
	err := arr.Unmarshal(&out)
	if err != nil {
		t.Errorf("nil array should return nil error, got %v", err)
	}
}

// --- Array.Unmarshal empty ---

func TestArrayUnmarshal_Empty(t *testing.T) {
	arr, _ := ParseJSON([]byte(`[]`))
	var out []string
	err := arr.Unmarshal(&out)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty slice, got %+v", out)
	}
}

// --- Object.Unmarshal with map field ---

func TestUnmarshal_MapField(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"labels":{"env":"prod","tier":"web"}}`))
	type Config struct {
		Labels map[string]string `json:"labels"`
	}
	var cfg Config
	err := obj.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Labels["env"] != "prod" || cfg.Labels["tier"] != "web" {
		t.Errorf("got %+v", cfg)
	}
}

// --- Object.Unmarshal with pointer field ---

func TestUnmarshal_PointerField(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"user":{"name":"alice"}}`))
	type User struct {
		Name string `json:"name"`
	}
	type Config struct {
		User *User `json:"user"`
	}
	var cfg Config
	err := obj.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.User == nil || cfg.User.Name != "alice" {
		t.Errorf("got %+v", cfg)
	}
}

// --- Object.Unmarshal with nil pointer field ---

func TestUnmarshal_NilPointerField(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"user":null}`))
	type User struct {
		Name string `json:"name"`
	}
	type Config struct {
		User *User `json:"user"`
	}
	var cfg Config
	err := obj.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.User != nil {
		t.Errorf("expected nil, got %+v", cfg.User)
	}
}

// --- Object.Unmarshal with pointer-to-scalar fields ---

func TestUnmarshal_PointerScalarFields(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"settingStatus":2,"enabled":true,"note":"hi","ratio":1.5}`))
	type Req struct {
		SettingStatus *int     `json:"settingStatus"`
		Enabled       *bool    `json:"enabled"`
		Note          *string  `json:"note"`
		Ratio         *float64 `json:"ratio"`
	}
	var r Req
	err := obj.Unmarshal(&r)
	if err != nil {
		t.Fatal(err)
	}
	if r.SettingStatus == nil || *r.SettingStatus != 2 {
		t.Errorf("SettingStatus: got %v", r.SettingStatus)
	}
	if r.Enabled == nil || !*r.Enabled {
		t.Errorf("Enabled: got %v", r.Enabled)
	}
	if r.Note == nil || *r.Note != "hi" {
		t.Errorf("Note: got %v", r.Note)
	}
	if r.Ratio == nil || *r.Ratio != 1.5 {
		t.Errorf("Ratio: got %v", r.Ratio)
	}
}

// --- Object.Unmarshal 缺省与 null 的指针字段保持 nil ---

func TestUnmarshal_PointerScalarNil(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"settingStatus":null,"other":1}`))
	type Req struct {
		SettingStatus *int `json:"settingStatus"`
		StoryStatus   *int `json:"storyStatus"`
	}
	var r Req
	if err := obj.Unmarshal(&r); err != nil {
		t.Fatal(err)
	}
	if r.SettingStatus != nil {
		t.Errorf("null 应保持 nil, got %v", *r.SettingStatus)
	}
	if r.StoryStatus != nil {
		t.Errorf("缺省应保持 nil, got %v", *r.StoryStatus)
	}
}

// --- Object.Unmarshal 弱类型转换到指针字段 ---

func TestUnmarshal_PointerScalarWeaklyTyped(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"settingStatus":"2"}`))
	type Req struct {
		SettingStatus *int `json:"settingStatus"`
	}
	var r Req
	if err := obj.Unmarshal(&r); err != nil {
		t.Fatal(err)
	}
	if r.SettingStatus == nil || *r.SettingStatus != 2 {
		t.Errorf("got %v", r.SettingStatus)
	}
}

// --- Array.Unmarshal 指针元素 ---

func TestArrayUnmarshal_PointerElements(t *testing.T) {
	arr, _ := ParseJSON([]byte(`[1,2,3]`))
	var out []*int
	if err := arr.Unmarshal(&out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 || out[0] == nil || *out[0] != 1 || *out[2] != 3 {
		t.Errorf("got %+v", out)
	}
}

// --- Object.Unmarshal with MatchFieldName disabled ---

func TestUnmarshal_MatchFieldNameDisabled(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"max_open_conns":10}`))
	type Config struct {
		MaxOpenConns int `json:"max_open_conns"`
	}
	var cfg Config
	// 禁用字段名匹配，只靠 json tag
	err := obj.Unmarshal(&cfg, WithTagName("json"), WithMatchFieldName(false))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxOpenConns != 10 {
		t.Errorf("got %+v", cfg)
	}
}

// --- Object.Unmarshal with custom tag only ---

func TestUnmarshal_CustomTagOnly(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"name":"alice"}`))
	type User struct {
		Name string `yaml:"name"`
	}
	var u User
	err := obj.Unmarshal(&u, WithTagName("yaml"), WithMatchFieldName(false))
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "alice" {
		t.Errorf("got %+v", u)
	}
}

// --- Array of objects with case insensitive keys ---

func TestArrayUnmarshal_StructsCaseInsensitive(t *testing.T) {
	arr, _ := ParseJSON([]byte(`[{"USER_NAME":"alice"},{"USER_NAME":"bob"}]`))
	type Item struct {
		UserName string
	}
	var out []Item
	err := arr.Unmarshal(&out)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0].UserName != "alice" || out[1].UserName != "bob" {
		t.Errorf("got %+v", out)
	}
}

// --- Any.Unmarshal basic (no opts, falls back to json.Unmarshal) ---

func TestAnyUnmarshal_Basic(t *testing.T) {
	a := NewAny(map[string]any{"name": "alice", "age": float64(30)})
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var u User
	err := a.Unmarshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "alice" || u.Age != 30 {
		t.Errorf("got %+v", u)
	}
}

// --- Any.Unmarshal with custom json tag ---

func TestAnyUnmarshal_WithJsonTag(t *testing.T) {
	a := NewAny(map[string]any{"user_name": "bob"})
	type User struct {
		Name string `json:"user_name"`
	}
	var u User
	err := a.Unmarshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "bob" {
		t.Errorf("got %+v", u)
	}
}

// --- Any.Unmarshal with custom tag via opts ---

func TestAnyUnmarshal_WithCustomTag(t *testing.T) {
	a := NewAny(map[string]any{"name": "carol"})
	type User struct {
		Name string `yaml:"name"`
	}
	var u User
	err := a.Unmarshal(&u, WithTagName("yaml"), WithMatchFieldName(false))
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "carol" {
		t.Errorf("got %+v", u)
	}
}

// --- Any.Unmarshal with snake_case field matching ---

func TestAnyUnmarshal_SnakeCase(t *testing.T) {
	a := NewAny(map[string]any{"max_open_conns": 50})
	type Config struct {
		MaxOpenConns int
	}
	var cfg Config
	err := a.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxOpenConns != 50 {
		t.Errorf("got %+v", cfg)
	}
}

// --- Any.Unmarshal with case insensitive keys ---

func TestAnyUnmarshal_CaseInsensitive(t *testing.T) {
	a := NewAny(map[string]any{"MAX_OPEN_CONNS": 25})
	type Config struct {
		MaxOpenConns int
	}
	var cfg Config
	err := a.Unmarshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxOpenConns != 25 {
		t.Errorf("got %+v", cfg)
	}
}

// --- Any.Unmarshal with weakly typed input ---

func TestAnyUnmarshal_WeaklyTyped(t *testing.T) {
	a := NewAny(map[string]any{"name": "alice", "age": "25", "active": "true"})
	type User struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}
	var u User
	err := a.Unmarshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "alice" || u.Age != 25 || !u.Active {
		t.Errorf("got %+v", u)
	}
}

// --- Any.Unmarshal with strict mode ---

func TestAnyUnmarshal_StrictMode(t *testing.T) {
	a := NewAny(map[string]any{"age": "not_a_number"})
	type User struct {
		Age int `json:"age"`
	}
	var u User
	err := a.Unmarshal(&u, WithWeaklyTypedInput(false))
	if err == nil {
		t.Error("expected error in strict mode")
	}
}

// --- Any.Unmarshal nil ---

func TestAnyUnmarshal_Nil(t *testing.T) {
	var a *Any
	type User struct {
		Name string
	}
	var u User
	err := a.Unmarshal(&u)
	if err != nil {
		t.Errorf("nil any should return nil error, got %v", err)
	}
}

// --- Any.Unmarshal nil value ---

func TestAnyUnmarshal_NilValue(t *testing.T) {
	a := NewAny(nil)
	type User struct {
		Name string
	}
	var u User
	err := a.Unmarshal(&u)
	if err != nil {
		t.Errorf("nil value should return nil error, got %v", err)
	}
}

// --- Any.Unmarshal without opts uses json.Unmarshal ---

func TestAnyUnmarshal_NoOptsJsonFallback(t *testing.T) {
	a := NewAny(map[string]any{"count": float64(42)})
	var out map[string]any
	err := a.Unmarshal(&out)
	if err != nil {
		t.Fatal(err)
	}
	if out["count"] != float64(42) {
		t.Errorf("got %+v", out)
	}
}

// --- 非 string key 的 map：按 key 类型解析，不 panic ---

func TestUnmarshal_MapNonStringKey(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"m":{"1":"a","2":"b"}}`))
	type Config struct {
		M map[int]string `json:"m"`
	}
	var cfg Config
	if err := obj.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.M[1] != "a" || cfg.M[2] != "b" {
		t.Errorf("got %+v", cfg.M)
	}
}

func TestUnmarshal_MapUint64Key(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"m":{"10":true}}`))
	type Config struct {
		M map[uint64]bool `json:"m"`
	}
	var cfg Config
	if err := obj.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if !cfg.M[10] {
		t.Errorf("got %+v", cfg.M)
	}
}

func TestUnmarshal_MapInvalidKey(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"m":{"abc":"a"}}`))
	type Config struct {
		M map[int]string `json:"m"`
	}
	var cfg Config
	err := obj.Unmarshal(&cfg)
	if err == nil {
		t.Fatalf("非法 key 应返回错误，got %+v", cfg.M)
	}
}

// --- bool 弱类型转换 ---

func TestUnmarshal_BoolWeaklyTyped(t *testing.T) {
	tests := []struct {
		raw  string
		want bool
	}{
		{`{"b":true}`, true},
		{`{"b":false}`, false},
		{`{"b":1}`, true},
		{`{"b":0}`, false},
		{`{"b":1.5}`, true},
		{`{"b":"1"}`, true},
		{`{"b":"0"}`, false},
		{`{"b":"true"}`, true},
		{`{"b":"True"}`, true},
		{`{"b":"TRUE"}`, true},
		{`{"b":"false"}`, false},
		{`{"b":"yes"}`, true},
		{`{"b":"no"}`, false},
		{`{"b":"on"}`, true},
		{`{"b":"off"}`, false},
		{`{"b":""}`, false},
	}
	type Config struct {
		B bool `json:"b"`
	}
	for _, tt := range tests {
		var cfg Config
		obj, _ := ParseJSON([]byte(tt.raw))
		if err := obj.Unmarshal(&cfg); err != nil {
			t.Errorf("%s: unexpected error %v", tt.raw, err)
			continue
		}
		if cfg.B != tt.want {
			t.Errorf("%s: got %v, want %v", tt.raw, cfg.B, tt.want)
		}
	}
}

func TestUnmarshal_BoolInvalid(t *testing.T) {
	type Config struct {
		B bool `json:"b"`
	}
	for _, raw := range []string{`{"b":"abc"}`, `{"b":[]}`, `{"b":{}}`} {
		var cfg Config
		obj, _ := ParseJSON([]byte(raw))
		if err := obj.Unmarshal(&cfg); err == nil {
			t.Errorf("%s: 期望报错，got %v", raw, cfg.B)
		}
	}
}

// --- time.Duration 字符串解析 ---

func TestUnmarshal_DurationFromString(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"timeout":"30s","idle":"1m30s","zero":"0"}`))
	type Config struct {
		Timeout time.Duration `json:"timeout"`
		Idle    time.Duration `json:"idle"`
		Zero    time.Duration `json:"zero"`
	}
	var cfg Config
	if err := obj.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout: got %v", cfg.Timeout)
	}
	if cfg.Idle != 90*time.Second {
		t.Errorf("Idle: got %v", cfg.Idle)
	}
	if cfg.Zero != 0 {
		t.Errorf("Zero: got %v", cfg.Zero)
	}
}

func TestUnmarshal_PointerDurationFromString(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"timeout":"5m"}`))
	type Config struct {
		Timeout *time.Duration `json:"timeout"`
	}
	var cfg Config
	if err := obj.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Timeout == nil || *cfg.Timeout != 5*time.Minute {
		t.Errorf("got %v", cfg.Timeout)
	}
}

func TestUnmarshal_DurationFromNumber(t *testing.T) {
	// 数字仍按纳秒处理（与 time.Duration 字面量一致）
	obj, _ := ParseJSON([]byte(`{"timeout":1500000000}`))
	type Config struct {
		Timeout time.Duration `json:"timeout"`
	}
	var cfg Config
	if err := obj.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Timeout != 1500*time.Millisecond {
		t.Errorf("got %v", cfg.Timeout)
	}
}

func TestUnmarshal_DurationInvalid(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"timeout":"abc"}`))
	type Config struct {
		Timeout time.Duration `json:"timeout"`
	}
	var cfg Config
	if err := obj.Unmarshal(&cfg); err == nil {
		t.Errorf("期望报错，got %v", cfg.Timeout)
	}
}

// --- 整数写入 string 字段按十进制格式化，而非 rune ---

func TestUnmarshal_IntToString(t *testing.T) {
	type Config struct {
		Name  string `json:"name"`
		Count string `json:"count"`
	}
	// 程序化构造的 map（如数据库返回的 int64）不能变成 rune
	var cfg Config
	a := NewAny(map[string]any{"name": 65, "count": int64(1024)})
	if err := a.Unmarshal(&cfg, WithTagName("json")); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "65" {
		t.Errorf("Name: got %q, want %q", cfg.Name, "65")
	}
	if cfg.Count != "1024" {
		t.Errorf("Count: got %q, want %q", cfg.Count, "1024")
	}
}

func TestUnmarshal_IntToStringStrict(t *testing.T) {
	type Config struct {
		Name string `json:"name"`
	}
	var cfg Config
	a := NewAny(map[string]any{"name": 65})
	if err := a.Unmarshal(&cfg, WithTagName("json"), WithWeaklyTypedInput(false)); err == nil {
		t.Errorf("严格模式下应报错，got %q", cfg.Name)
	}
}

// --- 无符号字段不接受负数、越界值 ---

func TestUnmarshal_NegativeToUint(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"u":-5}`))
	type Config struct {
		U uint `json:"u"`
	}
	var cfg Config
	if err := obj.Unmarshal(&cfg); err == nil {
		t.Errorf("负数写入 uint 应报错，got %d", cfg.U)
	}
}

func TestUnmarshal_UintOverflow(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"u8":300}`))
	type Config struct {
		U8 uint8 `json:"u8"`
	}
	var cfg Config
	if err := obj.Unmarshal(&cfg); err == nil {
		t.Errorf("300 写入 uint8 应报错，got %d", cfg.U8)
	}
}

func TestUnmarshal_Uint64Large(t *testing.T) {
	// uint64 上限值写入 uint64 字段应正常
	type Config struct {
		U uint64 `json:"u"`
	}
	var cfg Config
	a := NewAny(map[string]any{"u": uint64(math.MaxUint64)})
	if err := a.Unmarshal(&cfg, WithTagName("json")); err != nil {
		t.Fatal(err)
	}
	if cfg.U != math.MaxUint64 {
		t.Errorf("got %d", cfg.U)
	}
}

func TestUnmarshal_Uint64ToInt64Overflow(t *testing.T) {
	type Config struct {
		I int64 `json:"i"`
	}
	var cfg Config
	a := NewAny(map[string]any{"i": uint64(1) << 63})
	if err := a.Unmarshal(&cfg, WithTagName("json")); err == nil {
		t.Errorf("溢出写入 int64 应报错，got %d", cfg.I)
	}
}

func TestUnmarshal_Float32Overflow(t *testing.T) {
	obj, _ := ParseJSON([]byte(`{"f":1e40}`))
	type Config struct {
		F float32 `json:"f"`
	}
	var cfg Config
	if err := obj.Unmarshal(&cfg); err == nil {
		t.Errorf("1e40 写入 float32 应报错，got %v", cfg.F)
	}
}

// --- 具名数值类型（如 type UserID int64）按底层类型处理 ---

func TestUnmarshal_NamedNumericTypes(t *testing.T) {
	type UserID int64
	type Level uint8
	type Ratio float32
	type Status string

	type Config struct {
		ID     int64  `json:"id"`
		UID    UserID `json:"uid"`
		Level  Level  `json:"level"`
		Ratio  Ratio  `json:"ratio"`
		Status Status `json:"status"`
	}
	var cfg Config
	a := NewAny(map[string]any{
		"id":     UserID(7),  // 具名 int64 → int64
		"uid":    int32(9),   // int32 → 具名 int64
		"level":  float64(3), // float64 → 具名 uint8
		"ratio":  int(2),     // int → 具名 float32
		"status": "active",   // string → 具名 string
	})
	if err := a.Unmarshal(&cfg, WithTagName("json")); err != nil {
		t.Fatal(err)
	}
	if cfg.ID != 7 || cfg.UID != 9 || cfg.Level != 3 || cfg.Ratio != 2 || cfg.Status != "active" {
		t.Errorf("got %+v", cfg)
	}
}

func TestUnmarshal_DurationToInt64(t *testing.T) {
	// time.Duration 作为来源值也应可写入普通整数字段
	type Config struct {
		N int64 `json:"n"`
	}
	var cfg Config
	a := NewAny(map[string]any{"n": 3 * time.Second})
	if err := a.Unmarshal(&cfg, WithTagName("json")); err != nil {
		t.Fatal(err)
	}
	if cfg.N != int64(3*time.Second) {
		t.Errorf("got %d", cfg.N)
	}
}

// --- 字符串数字解析：拒绝尾随垃圾 ---

func TestUnmarshal_NumberStringStrict(t *testing.T) {
	type Config struct {
		A int     `json:"a"`
		F float64 `json:"f"`
	}
	for _, raw := range []string{`{"a":"12abc"}`, `{"a":"1.9"}`} {
		var cfg Config
		obj, _ := ParseJSON([]byte(raw))
		if err := obj.Unmarshal(&cfg); err == nil {
			t.Errorf("%s: 期望报错，got %d", raw, cfg.A)
		}
	}
	var cfg Config
	obj, _ := ParseJSON([]byte(`{"f":"1.5e3"}`))
	if err := obj.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.F != 1500 {
		t.Errorf("got %v", cfg.F)
	}
}

// --- 匿名内嵌结构体：提升字段可直接匹配 ---

func TestUnmarshal_EmbeddedStruct(t *testing.T) {
	type Base struct {
		Id int `json:"id"`
	}
	type Config struct {
		Base
		Name string `json:"name"`
	}
	var cfg Config
	obj, _ := ParseJSON([]byte(`{"id":7,"name":"alice"}`))
	if err := obj.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Id != 7 || cfg.Name != "alice" {
		t.Errorf("got %+v", cfg)
	}
}

func TestUnmarshal_EmbeddedPointerStruct(t *testing.T) {
	type Base struct {
		Id int `json:"id"`
	}
	type Config struct {
		*Base
		Name string `json:"name"`
	}

	// 有字段命中：分配并写入
	var cfg Config
	obj, _ := ParseJSON([]byte(`{"id":7,"name":"alice"}`))
	if err := obj.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Base == nil || cfg.Id != 7 || cfg.Name != "alice" {
		t.Errorf("got %+v", cfg)
	}

	// 无字段命中：保持 nil，不凭空分配
	var cfg2 Config
	obj2, _ := ParseJSON([]byte(`{"name":"bob"}`))
	if err := obj2.Unmarshal(&cfg2); err != nil {
		t.Fatal(err)
	}
	if cfg2.Base != nil {
		t.Errorf("内嵌指针应保持 nil, got %+v", cfg2.Base)
	}
	if cfg2.Name != "bob" {
		t.Errorf("got %+v", cfg2)
	}
}

// --- time.Time 从字符串解析 ---

func TestUnmarshal_TimeFromString(t *testing.T) {
	type Config struct {
		When time.Time `json:"when"`
	}
	var cfg Config
	obj, _ := ParseJSON([]byte(`{"when":"2026-09-14T10:00:00Z"}`))
	if err := obj.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	if !cfg.When.Equal(want) {
		t.Errorf("got %v, want %v", cfg.When, want)
	}
}

// --- encoding.TextUnmarshaler ---

type textStatus string

func (s *textStatus) UnmarshalText(b []byte) error {
	*s = textStatus("text:" + string(b))
	return nil
}

func TestUnmarshal_TextUnmarshaler(t *testing.T) {
	type Config struct {
		Status textStatus `json:"status"`
	}
	var cfg Config
	obj, _ := ParseJSON([]byte(`{"status":"active"}`))
	if err := obj.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Status != "text:active" {
		t.Errorf("got %q", cfg.Status)
	}
}
