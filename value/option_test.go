package value

import (
	"testing"
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

// --- lookupKey ---

func TestLookupKey_ExactMatch(t *testing.T) {
	data := map[string]any{"name": "alice"}
	val, ok := lookupKey(data, []string{"name"})
	if !ok || val != "alice" {
		t.Errorf("expected alice, got %v (ok=%v)", val, ok)
	}
}

func TestLookupKey_CaseInsensitive(t *testing.T) {
	data := map[string]any{"Name": "alice"}
	val, ok := lookupKey(data, []string{"name"})
	if !ok || val != "alice" {
		t.Errorf("expected alice, got %v (ok=%v)", val, ok)
	}
}

func TestLookupKey_SnakeToCamel(t *testing.T) {
	data := map[string]any{"MaxOpenConns": 10}
	val, ok := lookupKey(data, []string{"max_open_conns"})
	if !ok || val != 10 {
		t.Errorf("expected 10, got %v (ok=%v)", val, ok)
	}
}

func TestLookupKey_CamelToSnake(t *testing.T) {
	data := map[string]any{"max_open_conns": 10}
	val, ok := lookupKey(data, []string{"MaxOpenConns"})
	if !ok || val != 10 {
		t.Errorf("expected 10, got %v (ok=%v)", val, ok)
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
