package value

import (
	"encoding/json"
	"testing"
)

// buildTestValue creates a nested Value for testing:
//
//	{
//	  "store": {
//	    "name": "My Store",
//	    "book": [
//	      { "category": "reference", "author": "Nigel", "title": "Sayings", "price": 8.95 },
//	      { "category": "fiction",   "author": "Evelyn", "title": "Sword",  "price": 12.99 },
//	      { "category": "fiction",   "author": "Herman", "title": "Moby",   "price": 8.99 },
//	      { "category": "fiction",   "author": "Tolkien","title": "Hobbit", "price": 22.99 }
//	    ],
//	    "bicycle": { "color": "red", "price": 19.95 }
//	  }
//	}
func buildTestValue() Value {
	raw := `{
		"store": {
			"name": "My Store",
			"book": [
				{"category":"reference","author":"Nigel","title":"Sayings","price":8.95},
				{"category":"fiction","author":"Evelyn","title":"Sword","price":12.99},
				{"category":"fiction","author":"Herman","title":"Moby","price":8.99},
				{"category":"fiction","author":"Tolkien","title":"Hobbit","price":22.99}
			],
			"bicycle": {"color":"red","price":19.95}
		}
	}`
	v, _ := ParseJSON(json.RawMessage(raw))
	return v
}

// --- Lookup: basic dot navigation ---

func TestLookup_BasicDotNavigation(t *testing.T) {
	v := buildTestValue()

	// Single level
	name := Lookup(v, "store.name")
	if name == nil || name.String() != "My Store" {
		t.Errorf("store.name = %v, want 'My Store'", name)
	}

	// Nested object
	color := Lookup(v, "store.bicycle.color")
	if color == nil || color.String() != "red" {
		t.Errorf("store.bicycle.color = %v, want 'red'", color)
	}
}

func TestLookup_WithRootPrefix(t *testing.T) {
	v := buildTestValue()

	name := Lookup(v, "$.store.name")
	if name == nil || name.String() != "My Store" {
		t.Errorf("$.store.name = %v, want 'My Store'", name)
	}

	color := Lookup(v, "$.store.bicycle.color")
	if color == nil || color.String() != "red" {
		t.Errorf("$.store.bicycle.color = %v, want 'red'", color)
	}
}

func TestLookup_RootOnly(t *testing.T) {
	v := buildTestValue()
	result := Lookup(v, "$")
	if result != v {
		t.Error("Lookup('$') should return the root value")
	}
}

func TestLookup_MissingPath(t *testing.T) {
	v := buildTestValue()

	if Lookup(v, "store.missing") != nil {
		t.Error("missing key should return nil")
	}
	if Lookup(v, "store.missing.deep.path") != nil {
		t.Error("deep missing path should return nil")
	}
	if Lookup(v, "nonexistent") != nil {
		t.Error("nonexistent top-level key should return nil")
	}
}

func TestLookup_NilSafe(t *testing.T) {
	if Lookup(nil, "any.path") != nil {
		t.Error("nil value should return nil")
	}
	if Lookup(buildTestValue(), "") != nil {
		t.Error("empty path should return nil")
	}
}

// --- Lookup: array index ---

func TestLookup_ArrayIndex(t *testing.T) {
	v := buildTestValue()

	// Positive index
	title := Lookup(v, "store.book[0].title")
	if title == nil || title.String() != "Sayings" {
		t.Errorf("store.book[0].title = %v, want 'Sayings'", title)
	}

	title = Lookup(v, "store.book[2].title")
	if title == nil || title.String() != "Moby" {
		t.Errorf("store.book[2].title = %v, want 'Moby'", title)
	}

	// Last element
	title = Lookup(v, "store.book[3].title")
	if title == nil || title.String() != "Hobbit" {
		t.Errorf("store.book[3].title = %v, want 'Hobbit'", title)
	}
}

func TestLookup_NegativeIndex(t *testing.T) {
	v := buildTestValue()

	// -1 = last element
	title := Lookup(v, "store.book[-1].title")
	if title == nil || title.String() != "Hobbit" {
		t.Errorf("store.book[-1].title = %v, want 'Hobbit'", title)
	}

	// -2 = second to last
	title = Lookup(v, "store.book[-2].title")
	if title == nil || title.String() != "Moby" {
		t.Errorf("store.book[-2].title = %v, want 'Moby'", title)
	}
}

func TestLookup_IndexOutOfBounds(t *testing.T) {
	v := buildTestValue()

	if Lookup(v, "store.book[99]") != nil {
		t.Error("out of bounds index should return nil")
	}
	if Lookup(v, "store.book[-99]") != nil {
		t.Error("large negative index should return nil")
	}
}

func TestLookup_IndexOnNonArray(t *testing.T) {
	v := buildTestValue()

	// store.name is a string, not an array
	if Lookup(v, "store.name[0]") != nil {
		t.Error("index on non-array should return nil")
	}
}

// --- Lookup: wildcard ---

func TestLookup_Wildcard(t *testing.T) {
	v := buildTestValue()

	// Wildcard on array returns all elements
	all := Lookup(v, "store.book[*]")
	if all == nil || !all.IsArray() {
		t.Fatalf("store.book[*] should return Array, got %T", all)
	}
	arr := all.AsArray()
	if arr.Len() != 4 {
		t.Errorf("store.book[*] length = %d, want 4", arr.Len())
	}

	// Wildcard on object returns all values
	colors := Lookup(v, "store.bicycle[*]")
	if colors == nil || !colors.IsArray() {
		t.Fatalf("store.bicycle[*] should return Array, got %T", colors)
	}
}

func TestLookup_WildcardThenKey(t *testing.T) {
	v := buildTestValue()

	// Get all authors
	authors := Lookup(v, "store.book[*].author")
	if authors == nil || !authors.IsArray() {
		t.Fatalf("store.book[*].author should return Array, got %T", authors)
	}
	arr := authors.AsArray()
	if arr.Len() != 4 {
		t.Errorf("author count = %d, want 4", arr.Len())
	}
	expected := []string{"Nigel", "Evelyn", "Herman", "Tolkien"}
	for i, exp := range expected {
		if arr.Get(i).String() != exp {
			t.Errorf("author[%d] = %q, want %q", i, arr.Get(i).String(), exp)
		}
	}
}

// --- Lookup: recursive descent ---

func TestLookup_RecursiveDescent(t *testing.T) {
	v := buildTestValue()

	// Find all "author" fields anywhere in the tree
	authors := LookupAll(v, "..author")
	if authors == nil || authors.Len() != 4 {
		t.Fatalf("..author count = %d, want 4", authors.Len())
	}

	// Find all "price" fields (bicycle.price + 4 book prices)
	prices := LookupAll(v, "..price")
	if prices == nil || prices.Len() != 5 {
		t.Fatalf("..price count = %d, want 5", prices.Len())
	}
}

func TestLookup_RecursiveDescentFromKey(t *testing.T) {
	v := buildTestValue()

	// Recursive descent returns Array of all matches
	titles := Lookup(v, "store..title")
	if titles == nil || !titles.IsArray() {
		t.Fatalf("store..title should return Array, got %T", titles)
	}
	arr := titles.AsArray()
	if arr.Len() != 4 {
		t.Errorf("store..title count = %d, want 4", arr.Len())
	}

	// LookupAll also works
	titles2 := LookupAll(v, "store..title")
	if titles2 == nil || titles2.Len() != 4 {
		t.Fatalf("LookupAll store..title count = %d, want 4", titles2.Len())
	}
}

func TestLookup_RecursiveDescentNested(t *testing.T) {
	raw := `{"a":{"b":{"c":{"d":"deep"}}}}`
	v, _ := ParseJSON(json.RawMessage(raw))

	// Recursive descent returns Array
	result := Lookup(v, "..d")
	if result == nil || !result.IsArray() {
		t.Fatalf("..d should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 1 || arr.Get(0).String() != "deep" {
		t.Errorf("..d = %v, want ['deep']", result)
	}

	result = Lookup(v, "a..d")
	if result == nil || !result.IsArray() {
		t.Fatalf("a..d should return Array, got %T", result)
	}
	arr = result.AsArray()
	if arr.Len() != 1 || arr.Get(0).String() != "deep" {
		t.Errorf("a..d = %v, want ['deep']", result)
	}
}

// --- Lookup: slice ---

func TestLookup_Slice(t *testing.T) {
	v := buildTestValue()

	// book[0:2] → first two books
	result := Lookup(v, "store.book[0:2]")
	if result == nil || !result.IsArray() {
		t.Fatalf("store.book[0:2] should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 2 {
		t.Errorf("store.book[0:2] length = %d, want 2", arr.Len())
	}
	if arr.Get(0).AsObject().GetString("title") != "Sayings" {
		t.Error("first element should be 'Sayings'")
	}
	if arr.Get(1).AsObject().GetString("title") != "Sword" {
		t.Error("second element should be 'Sword'")
	}
}

func TestLookup_SliceFromBeginning(t *testing.T) {
	v := buildTestValue()

	// [:2] → first two
	result := Lookup(v, "store.book[:2]")
	if result == nil || !result.IsArray() {
		t.Fatalf("store.book[:2] should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 2 {
		t.Errorf("store.book[:2] length = %d, want 2", arr.Len())
	}
}

func TestLookup_SliceToEnd(t *testing.T) {
	v := buildTestValue()

	// [2:] → last two
	result := Lookup(v, "store.book[2:]")
	if result == nil || !result.IsArray() {
		t.Fatalf("store.book[2:] should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 2 {
		t.Errorf("store.book[2:] length = %d, want 2", arr.Len())
	}
	if arr.Get(0).AsObject().GetString("title") != "Moby" {
		t.Error("first element should be 'Moby'")
	}
}

func TestLookup_SliceNegative(t *testing.T) {
	v := buildTestValue()

	// [-2:] → last two books
	result := Lookup(v, "store.book[-2:]")
	if result == nil || !result.IsArray() {
		t.Fatalf("store.book[-2:] should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 2 {
		t.Errorf("store.book[-2:] length = %d, want 2", arr.Len())
	}
	if arr.Get(0).AsObject().GetString("title") != "Moby" {
		t.Error("first element should be 'Moby'")
	}
	if arr.Get(1).AsObject().GetString("title") != "Hobbit" {
		t.Error("second element should be 'Hobbit'")
	}
}

// --- Lookup: union ---

func TestLookup_Union(t *testing.T) {
	v := buildTestValue()

	// [0,3] → first and last
	result := Lookup(v, "store.book[0,3]")
	if result == nil || !result.IsArray() {
		t.Fatalf("store.book[0,3] should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 2 {
		t.Errorf("store.book[0,3] length = %d, want 2", arr.Len())
	}
	if arr.Get(0).AsObject().GetString("title") != "Sayings" {
		t.Error("first element should be 'Sayings'")
	}
	if arr.Get(1).AsObject().GetString("title") != "Hobbit" {
		t.Error("second element should be 'Hobbit'")
	}
}

func TestLookup_UnionNegative(t *testing.T) {
	v := buildTestValue()

	// [-1,0] → last and first
	result := Lookup(v, "store.book[-1,0]")
	if result == nil || !result.IsArray() {
		t.Fatalf("store.book[-1,0] should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 2 {
		t.Errorf("length = %d, want 2", arr.Len())
	}
	if arr.Get(0).AsObject().GetString("title") != "Hobbit" {
		t.Error("first should be 'Hobbit' (last element)")
	}
}

func TestLookup_SingleUnion(t *testing.T) {
	v := buildTestValue()

	// [0,0] returns the same element twice → Array of 2
	result := Lookup(v, "store.book[0,0]")
	if result == nil || !result.IsArray() {
		t.Fatalf("store.book[0,0] should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 2 {
		t.Errorf("length = %d, want 2", arr.Len())
	}
}

// --- Lookup: bracket notation for keys ---

func TestLookup_BracketNotation(t *testing.T) {
	raw := `{"key.with.dots": "value1", "key[bracket]": "value2"}`
	v, _ := ParseJSON(json.RawMessage(raw))

	// Quoted key with dots
	result := Lookup(v, `['key.with.dots']`)
	if result == nil || result.String() != "value1" {
		t.Errorf("bracket key with dots = %v, want 'value1'", result)
	}

	// Double-quoted variant
	result = Lookup(v, `["key.with.dots"]`)
	if result == nil || result.String() != "value1" {
		t.Errorf("bracket key with double quotes = %v, want 'value1'", result)
	}
}

// --- LookupFirst ---

func TestLookupFirst(t *testing.T) {
	v := buildTestValue()

	// First path matches
	result := LookupFirst(v, "store.name", "name")
	if result == nil || result.String() != "My Store" {
		t.Errorf("LookupFirst = %v, want 'My Store'", result)
	}

	// Second path matches (fallback)
	result = LookupFirst(v, "name", "store.name")
	if result == nil || result.String() != "My Store" {
		t.Errorf("LookupFirst fallback = %v, want 'My Store'", result)
	}

	// No match
	result = LookupFirst(v, "a", "b", "c")
	if result != nil {
		t.Error("LookupFirst with no match should return nil")
	}
}

// --- LookupAll ---

func TestLookupAll_Simple(t *testing.T) {
	v := buildTestValue()

	// All prices in the tree
	prices := LookupAll(v, "store.book[*].price")
	if prices == nil || prices.Len() != 4 {
		t.Fatalf("store.book[*].price count = %d, want 4", prices.Len())
	}
}

func TestLookupAll_RecursiveDescent(t *testing.T) {
	v := buildTestValue()

	// All "price" anywhere
	prices := LookupAll(v, "..price")
	if prices == nil || prices.Len() != 5 {
		t.Fatalf("..price count = %d, want 5", prices.Len())
	}
}

func TestLookupAll_NilSafe(t *testing.T) {
	if LookupAll(nil, "any") != nil {
		t.Error("nil value should return nil")
	}
	if LookupAll(buildTestValue(), "") != nil {
		t.Error("empty path should return nil")
	}
}

// --- Lookup: complex combined paths ---

func TestLookup_CombinedPath(t *testing.T) {
	v := buildTestValue()

	// Get all fiction books' authors via slice + key
	result := Lookup(v, "store.book[1:3].author")
	if result == nil || !result.IsArray() {
		t.Fatalf("store.book[1:3].author should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 2 {
		t.Errorf("length = %d, want 2", arr.Len())
	}
	if arr.Get(0).String() != "Evelyn" {
		t.Errorf("first author = %q, want 'Evelyn'", arr.Get(0).String())
	}
	if arr.Get(1).String() != "Herman" {
		t.Errorf("second author = %q, want 'Herman'", arr.Get(1).String())
	}
}

// --- parsePath edge cases ---

func TestParsePath_Whitespace(t *testing.T) {
	v := buildTestValue()

	// Spaces around brackets
	result := Lookup(v, "store.book[ 0 ].title")
	if result == nil || result.String() != "Sayings" {
		t.Errorf("whitespace in brackets = %v, want 'Sayings'", result)
	}
}

func TestParsePath_ConsecutiveDots(t *testing.T) {
	v := buildTestValue()

	// Double dots = recursive descent (returns Array of all matches)
	result := Lookup(v, "store..title")
	if result == nil || !result.IsArray() {
		t.Fatalf("store..title should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 4 {
		t.Errorf("store..title count = %d, want 4", arr.Len())
	}
}

// --- LookupOption ---

func TestLookup_MatchExact(t *testing.T) {
	raw := `{"userName":"alice","user_name":"bob","USERNAME":"charlie"}`
	v, _ := ParseJSON(json.RawMessage(raw))

	// Exact match — only "userName" works
	if got := Lookup(v, "userName", MatchExact()); got == nil || got.String() != "alice" {
		t.Errorf("exact 'userName' = %v, want 'alice'", got)
	}

	// Exact miss — "username" (lowercase) does NOT match "userName"
	if got := Lookup(v, "username", MatchExact()); got != nil {
		t.Errorf("exact 'username' should be nil, got %v", got)
	}

	// Exact miss — "USERNAME" does NOT match "userName"
	if got := Lookup(v, "USERNAME", MatchExact()); got == nil || got.String() != "charlie" {
		// "USERNAME" exists as its own key, so exact match succeeds
		t.Errorf("exact 'USERNAME' = %v, want 'charlie'", got)
	}
}

func TestLookup_MatchCaseInsensitive(t *testing.T) {
	raw := `{"userName":"alice","user_name":"bob"}`
	v, _ := ParseJSON(json.RawMessage(raw))

	// Case-insensitive: "username" matches "userName"
	if got := Lookup(v, "username", MatchCaseInsensitive()); got == nil || got.String() != "alice" {
		t.Errorf("ci 'username' = %v, want 'alice'", got)
	}

	// Case-insensitive: "USERNAME" matches "userName"
	if got := Lookup(v, "USERNAME", MatchCaseInsensitive()); got == nil || got.String() != "alice" {
		t.Errorf("ci 'USERNAME' = %v, want 'alice'", got)
	}

	// Case-insensitive WITHOUT snake_case: "user_name" does NOT match "userName"
	if got := Lookup(v, "user_name", MatchCaseInsensitive()); got == nil || got.String() != "bob" {
		// "user_name" exists as its own key, so it matches directly
		t.Errorf("ci 'user_name' = %v, want 'bob'", got)
	}
}

func TestLookup_MatchFlexible(t *testing.T) {
	raw := `{"user_name":"alice","MaxOpenConns":100}`
	v, _ := ParseJSON(json.RawMessage(raw))

	// Flexible: "userName" → snake_case "user_name" matches
	if got := Lookup(v, "userName", MatchFlexible()); got == nil || got.String() != "alice" {
		t.Errorf("flex 'userName' = %v, want 'alice'", got)
	}

	// Flexible: "max_open_conns" → CamelCase "MaxOpenConns" matches
	if got := Lookup(v, "max_open_conns", MatchFlexible()); got == nil || got.AsNumber().Int64() != 100 {
		t.Errorf("flex 'max_open_conns' = %v, want 100", got)
	}
}

func TestLookup_DefaultIsFlexible(t *testing.T) {
	raw := `{"user_name":"alice"}`
	v, _ := ParseJSON(json.RawMessage(raw))

	// Default behavior should be the same as MatchFlexible()
	if got := Lookup(v, "userName"); got == nil || got.String() != "alice" {
		t.Errorf("default 'userName' = %v, want 'alice'", got)
	}
}

func TestLookupAll_WithOptions(t *testing.T) {
	raw := `{"items":[{"UserName":"a"},{"UserName":"b"}]}`
	v, _ := ParseJSON(json.RawMessage(raw))

	// Flexible: "username" → case-insensitive matches "UserName"
	all := LookupAll(v, "items[*].username", MatchFlexible())
	if all == nil || all.Len() != 2 {
		t.Fatalf("flex LookupAll count = %d, want 2", all.Len())
	}

	// Exact: "username" does NOT match "UserName"
	all = LookupAll(v, "items[*].username", MatchExact())
	if all != nil {
		t.Error("exact LookupAll should return nil for case mismatch")
	}
}

func TestLookup_RecursiveDescent_WithOptions(t *testing.T) {
	raw := `{"a":{"UserName":"deep"}}`
	v, _ := ParseJSON(json.RawMessage(raw))

	// Flexible: "username" → case-insensitive matches "UserName"
	result := Lookup(v, "..username", MatchFlexible())
	if result == nil || !result.IsArray() {
		t.Fatalf("..username flex should return Array, got %T", result)
	}
	arr := result.AsArray()
	if arr.Len() != 1 || arr.Get(0).String() != "deep" {
		t.Errorf("..username flex = %v, want ['deep']", result)
	}

	// Exact: "username" does NOT match "UserName"
	result = Lookup(v, "..username", MatchExact())
	if result != nil {
		t.Error("..username exact should return nil")
	}
}

func TestLookup_NormalizedMatch(t *testing.T) {
	raw := `{"User_Name":"alice","max-open-conns":100}`
	v, _ := ParseJSON(json.RawMessage(raw))

	// "userName" normalizes to "username", "User_Name" normalizes to "username"
	if got := Lookup(v, "userName", MatchFlexible()); got == nil || got.String() != "alice" {
		t.Errorf("normalized 'userName' = %v, want 'alice'", got)
	}

	// "user_name" direct match (also normalizes)
	if got := Lookup(v, "user_name", MatchFlexible()); got == nil || got.String() != "alice" {
		t.Errorf("normalized 'user_name' = %v, want 'alice'", got)
	}

	// "maxOpenConns" → camelToSnake → "max_open_conns" → normalize → "maxopenconns"
	// "max-open-conns" → normalize → "maxopenconns"
	if got := Lookup(v, "maxOpenConns", MatchFlexible()); got == nil || got.AsNumber().Int64() != 100 {
		t.Errorf("normalized 'maxOpenConns' = %v, want 100", got)
	}

	// Exact: "userName" does NOT match "User_Name"
	if got := Lookup(v, "userName", MatchExact()); got != nil {
		t.Error("exact 'userName' should not match 'User_Name'")
	}
}

// --- Object.LookupByPath vs package-level Lookup ---

func TestLookupVsObjectLookupByPath(t *testing.T) {
	raw := `{"a":{"b":{"c":"hello"}}}`
	v, _ := ParseJSON(json.RawMessage(raw))
	obj := v.AsObject()

	// Both should give the same result for simple dot paths
	r1 := obj.LookupByPath("a.b.c")
	r2 := Lookup(v, "a.b.c")
	if r1 == nil || r2 == nil || r1.String() != r2.String() {
		t.Errorf("LookupByPath and Lookup should agree: %v vs %v", r1, r2)
	}
}

// --- Benchmark ---

func BenchmarkLookup_SimplePath(b *testing.B) {
	v := buildTestValue()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lookup(v, "store.name")
	}
}

func BenchmarkLookup_DeepPath(b *testing.B) {
	v := buildTestValue()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lookup(v, "store.book[2].title")
	}
}

func BenchmarkLookup_Wildcard(b *testing.B) {
	v := buildTestValue()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lookup(v, "store.book[*].author")
	}
}

func BenchmarkLookup_RecursiveDescent(b *testing.B) {
	v := buildTestValue()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LookupAll(v, "..price")
	}
}
