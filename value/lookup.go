package value

import (
	"strconv"
	"strings"
)

// LookupOption controls how Lookup matches object keys.
// The zero value disables all fuzzy matching (equivalent to MatchExact);
// the framework default (defaultLookupOption) enables all three.
type LookupOption struct {
	// CaseInsensitive enables matching keys regardless of letter case,
	// e.g. "username" matches "userName", "USERNAME".
	CaseInsensitive bool
	// SnakeCase enables CamelCase → snake_case conversion,
	// e.g. "userName" matches "user_name".
	SnakeCase bool
	// CamelCase enables snake_case → CamelCase conversion,
	// e.g. "user_name" matches "UserName".
	CamelCase bool
}

// MatchExact returns a LookupOption that only matches keys exactly.
// No case folding, no snake/camel conversion.
func MatchExact() LookupOption {
	return LookupOption{CaseInsensitive: false, SnakeCase: false, CamelCase: false}
}

// MatchCaseInsensitive returns a LookupOption that matches keys case-insensitively
// but does not convert between snake_case and CamelCase.
func MatchCaseInsensitive() LookupOption {
	return LookupOption{CaseInsensitive: true, SnakeCase: false, CamelCase: false}
}

var defaultLookupOption = LookupOption{CaseInsensitive: true, SnakeCase: true, CamelCase: true}

// MatchFlexible returns a LookupOption with all matching strategies enabled.
// This is the default: case-insensitive + snake_case + CamelCase fallback.
func MatchFlexible() LookupOption {
	return defaultLookupOption
}

// preparePath trims whitespace and strips the optional "$." prefix.
// The second return value reports whether the path is exactly "$" (the root).
func preparePath(path string) (string, bool) {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "$.") {
		return path[2:], false
	}
	if path == "$" {
		return "", true
	}
	return path, false
}

// Lookup performs a JSONPath-like lookup on the given Value.
//
// Supported syntax:
//
//	store.book[0].title       — dot navigation + array index
//	store.book[*].title       — wildcard (all array elements)
//	store..title              — recursive descent (search all descendants)
//	store.book[-1].title      — negative index (last element)
//	store.book[0,2].title     — union (multiple indices)
//	store.book[0:2].title     — slice (start:end, end exclusive)
//	store.book[:3].title      — slice from beginning
//	store.book[2:].title      — slice to end
//	store['key-with.dot'].x   — bracket notation for keys with special chars
//
// The leading "$" prefix is optional and stripped automatically.
// Returns nil if the path does not resolve.
//
// Key matching options control how object keys are resolved:
//
//	Lookup(v, path)                   — default: case-insensitive + snake/camel fallback
//	Lookup(v, path, MatchExact())     — exact match only
//	Lookup(v, path, MatchCaseInsensitive())  — case-insensitive only
func Lookup(v Value, path string, opts ...LookupOption) Value {
	if v == nil || path == "" {
		return nil
	}
	cfg := defaultLookupOption
	if len(opts) > 0 {
		cfg = opts[0]
	}
	path, isRoot := preparePath(path)
	if isRoot {
		return v
	}
	segments := parsePath(path)
	var current Value = v
	for _, seg := range segments {
		if current == nil {
			return nil
		}
		if isMultiSegment(seg) {
			matches := evalSegmentAllCfg(current, seg, cfg)
			if len(matches) == 0 {
				return nil
			}
			current = NewArray(matches...)
		} else {
			if seg.hasIdx {
				arr := asArraySafe(current)
				if arr == nil {
					return nil
				}
				current = getByIndex(arr, seg.index)
			} else if arr := asArraySafe(current); arr != nil {
				var results []Value
				for _, item := range arr.data {
					if r := evalSegmentOneCfg(item, seg, cfg); r != nil {
						results = append(results, r)
					}
				}
				if len(results) == 0 {
					return nil
				}
				current = NewArray(results...)
			} else {
				current = evalSegmentOneCfg(current, seg, cfg)
			}
		}
	}
	return current
}

// LookupFirst is like Lookup but returns the first non-nil result from
// multiple alternative paths. Useful for fallback matching:
//
//	LookupFirst(v, "data.name", "name")  // try data.name first, then name
func LookupFirst(v Value, paths ...string) Value {
	for _, p := range paths {
		if r := Lookup(v, p); r != nil {
			return r
		}
	}
	return nil
}

// LookupAll returns all values matching the given path. Unlike Lookup which
// returns only the first match for single-value segments, LookupAll collects
// every match into an Array.
func LookupAll(v Value, path string, opts ...LookupOption) *Array {
	if v == nil || path == "" {
		return nil
	}
	cfg := defaultLookupOption
	if len(opts) > 0 {
		cfg = opts[0]
	}
	path, isRoot := preparePath(path)
	if isRoot {
		return NewArray(v)
	}
	segments := parsePath(path)
	results := []Value{v}
	for _, seg := range segments {
		if len(results) == 0 {
			return nil
		}
		var next []Value
		for _, cur := range results {
			matches := evalSegmentAllCfg(cur, seg, cfg)
			next = append(next, matches...)
		}
		results = next
	}
	if len(results) == 0 {
		return nil
	}
	return NewArray(results...)
}

// matchKey looks up a key in an Object using the given matching config.
// It follows the same strategy chain as Object.matchKey but with configurable steps.
func matchKey(obj *Object, key string, cfg LookupOption) Value {
	if obj == nil {
		return nil
	}

	// 1. Exact match
	if v, ok := obj.data[key]; ok {
		return v
	}

	if !cfg.CaseInsensitive && !cfg.SnakeCase && !cfg.CamelCase {
		return nil
	}

	lowerKey := strings.ToLower(key)

	// 2. Lowercase match
	if cfg.CaseInsensitive && lowerKey != key {
		if v, ok := obj.data[lowerKey]; ok {
			return v
		}
	}

	// 3. snake_case conversion
	if cfg.SnakeCase {
		snakeKey := camelToSnake(key)
		if snakeKey != key && snakeKey != lowerKey {
			if v, ok := obj.data[snakeKey]; ok {
				return v
			}
		}
	}

	// 4–6. Fallback scans merged into a single pass over the keys.
	// Priority: case-insensitive (4) > camel (5) > normalized (6).
	camelKey := ""
	if cfg.CamelCase {
		camelKey = strings.ToLower(snakeToCamel(key))
	}
	normKey := ""
	doNormalize := cfg.SnakeCase || cfg.CamelCase
	if doNormalize {
		normKey = normalizeKey(lowerKey)
	}

	var camelMatch, normMatch Value
	camelSet, normSet := false, false
	for k, v := range obj.data {
		lk := strings.ToLower(k)
		if cfg.CaseInsensitive && lk == lowerKey {
			return v
		}
		if cfg.CamelCase && !camelSet && lk == camelKey {
			camelMatch, camelSet = v, true
		}
		if doNormalize && !normSet && normalizeKey(lk) == normKey {
			normMatch, normSet = v, true
		}
	}
	if camelSet {
		return camelMatch
	}
	if normSet {
		return normMatch
	}
	return nil
}

// normalizeKey removes underscores, spaces, and hyphens for fuzzy key matching.
func normalizeKey(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if c := s[i]; c != '_' && c != ' ' && c != '-' {
			b = append(b, c)
		}
	}
	return string(b)
}

// isMultiSegment returns true if the segment produces multiple values.
func isMultiSegment(seg pathSegment) bool {
	return seg.wild || seg.recurse || seg.isSlice || seg.indices != nil
}

// evalSegmentOneCfg applies a single-value segment with matching config.
func evalSegmentOneCfg(v Value, seg pathSegment, cfg LookupOption) Value {
	if v == nil {
		return nil
	}
	if seg.hasIdx {
		arr := asArraySafe(v)
		if arr == nil {
			return nil
		}
		return getByIndex(arr, seg.index)
	}
	if obj := asObjectSafe(v); obj != nil {
		return matchKey(obj, seg.key, cfg)
	}
	return nil
}

// pathSegment represents one segment of a parsed path.
type pathSegment struct {
	key     string // object key (empty for wildcard/recursive descent without key)
	index   int    // single array index (used when indices is nil)
	indices []int  // multiple indices for union [0,2,4]
	hasIdx  bool   // whether an index or slice was specified
	start   int    // slice start
	end     int    // slice end (exclusive, -1 means to end)
	isSlice bool   // whether this is a slice segment
	wild    bool   // wildcard [*]
	recurse bool   // recursive descent [..]
}

// parsePath splits a path string into segments.
// When ".." is followed by a key (e.g. "..title"), they are merged into one
// segment {recurse:true, key:"title"} for efficient lookup.
func parsePath(path string) []pathSegment {
	var segs []pathSegment
	i := 0
	for i < len(path) {
		if path[i] == '.' {
			i++
			if i < len(path) && path[i] == '.' {
				// Recursive descent — merge with following key if present
				i++
				if i < len(path) && path[i] != '.' && path[i] != '[' && path[i] != ']' {
					start := i
					for i < len(path) && path[i] != '.' && path[i] != '[' {
						i++
					}
					segs = append(segs, pathSegment{recurse: true, key: path[start:i]})
				} else {
					segs = append(segs, pathSegment{recurse: true})
				}
			}
			continue
		}
		if path[i] == '[' {
			i++
			segs = append(segs, parseBracket(path, &i))
			if i < len(path) && path[i] == ']' {
				i++
			}
			continue
		}
		start := i
		for i < len(path) && path[i] != '.' && path[i] != '[' {
			i++
		}
		key := path[start:i]
		if key != "" {
			segs = append(segs, pathSegment{key: key})
		}
	}
	return segs
}

// parseBracket parses the content inside [...].
func parseBracket(path string, i *int) pathSegment {
	for *i < len(path) && path[*i] == ' ' {
		*i++
	}
	if *i >= len(path) {
		return pathSegment{}
	}

	// Quoted key: ['key'] or ["key"]
	if path[*i] == '\'' || path[*i] == '"' {
		return parseQuotedKey(path, i)
	}

	// Wildcard
	if path[*i] == '*' {
		*i++
		return pathSegment{wild: true}
	}

	// Parse number (possibly negative)
	negative := false
	if *i < len(path) && path[*i] == '-' {
		negative = true
		*i++
	}
	digitStart := *i
	for *i < len(path) && path[*i] >= '0' && path[*i] <= '9' {
		*i++
	}
	for *i < len(path) && path[*i] == ' ' {
		*i++
	}

	// Slice [start:end]
	if *i < len(path) && path[*i] == ':' {
		return parseSlice(path, i, negative, digitStart)
	}

	// Union [a,b,c]
	if *i < len(path) && path[*i] == ',' {
		return parseUnion(path, i, negative, digitStart)
	}

	// Single number index
	if *i > digitStart {
		n, _ := strconv.Atoi(path[digitStart:*i])
		if negative {
			n = -n
		}
		return pathSegment{index: n, hasIdx: true}
	}

	// Bare key inside brackets
	*i = digitStart
	start := *i
	for *i < len(path) && path[*i] != ']' && path[*i] != ' ' {
		*i++
	}
	return pathSegment{key: path[start:*i]}
}

func parseQuotedKey(path string, i *int) pathSegment {
	quote := path[*i]
	*i++
	start := *i
	for *i < len(path) && path[*i] != quote {
		if path[*i] == '\\' {
			*i++
		}
		*i++
	}
	key := path[start:*i]
	if *i < len(path) {
		*i++
	}
	return pathSegment{key: key}
}

func parseSlice(path string, i *int, negative bool, digitStart int) pathSegment {
	startVal := 0
	if *i > digitStart {
		startVal, _ = strconv.Atoi(path[digitStart:*i])
		if negative {
			startVal = -startVal
		}
	}
	*i++ // skip ':'
	for *i < len(path) && path[*i] == ' ' {
		*i++
	}
	endVal := -1 // -1 means "to end"
	endNeg := false
	if *i < len(path) && path[*i] == '-' {
		endNeg = true
		*i++
	}
	endDigitStart := *i
	for *i < len(path) && path[*i] >= '0' && path[*i] <= '9' {
		*i++
	}
	if *i > endDigitStart {
		endVal, _ = strconv.Atoi(path[endDigitStart:*i])
		if endNeg {
			endVal = -endVal
		}
	}
	return pathSegment{isSlice: true, start: startVal, end: endVal, hasIdx: true}
}

func parseUnion(path string, i *int, negative bool, digitStart int) pathSegment {
	first, _ := strconv.Atoi(path[digitStart:*i])
	if negative {
		first = -first
	}
	indices := []int{first}

	for *i < len(path) && path[*i] == ',' {
		*i++
		for *i < len(path) && path[*i] == ' ' {
			*i++
		}
		neg := false
		if *i < len(path) && path[*i] == '-' {
			neg = true
			*i++
		}
		ds := *i
		for *i < len(path) && path[*i] >= '0' && path[*i] <= '9' {
			*i++
		}
		if *i > ds {
			n, _ := strconv.Atoi(path[ds:*i])
			if neg {
				n = -n
			}
			indices = append(indices, n)
		}
		for *i < len(path) && path[*i] == ' ' {
			*i++
		}
	}
	return pathSegment{indices: indices, hasIdx: true}
}

// evalSegmentAllCfg applies a single path segment to a Value, returning all matches.
func evalSegmentAllCfg(v Value, seg pathSegment, cfg LookupOption) []Value {
	if v == nil {
		return nil
	}

	// Recursive descent
	if seg.recurse {
		if seg.key != "" {
			return collectRecurseByKeyCfg(v, seg.key, cfg)
		}
		return collectAllDescendants(v)
	}

	// Wildcard
	if seg.wild {
		return expandWildcard(v)
	}

	// Array index / slice / union
	if seg.hasIdx {
		arr := asArraySafe(v)
		if arr == nil {
			return nil
		}
		if seg.isSlice {
			r := evalSlice(arr, seg)
			if r == nil {
				return nil
			}
			if a, ok := r.(*Array); ok {
				return a.data
			}
			return []Value{r}
		}
		if seg.indices != nil {
			r := evalUnion(arr, seg.indices)
			if r == nil {
				return nil
			}
			if a, ok := r.(*Array); ok {
				return a.data
			}
			return []Value{r}
		}
		el := getByIndex(arr, seg.index)
		if el == nil {
			return nil
		}
		return []Value{el}
	}

	// Object key lookup
	if obj := asObjectSafe(v); obj != nil {
		r := matchKey(obj, seg.key, cfg)
		if r == nil {
			return nil
		}
		return []Value{r}
	}
	return nil
}

// collectRecurseByKeyCfg finds all descendants matching a key (recursive descent).
func collectRecurseByKeyCfg(v Value, key string, cfg LookupOption) []Value {
	var results []Value
	if obj := asObjectSafe(v); obj != nil {
		if r := matchKey(obj, key, cfg); r != nil {
			results = append(results, r)
		}
	}
	switch val := v.(type) {
	case *Object:
		for _, child := range val.data {
			results = append(results, collectRecurseByKeyCfg(child, key, cfg)...)
		}
	case *Array:
		for _, child := range val.data {
			results = append(results, collectRecurseByKeyCfg(child, key, cfg)...)
		}
	}
	return results
}

// collectAllDescendants returns all leaf and non-leaf values in the tree.
func collectAllDescendants(v Value) []Value {
	var results []Value
	switch val := v.(type) {
	case *Object:
		for _, child := range val.data {
			results = append(results, child)
			results = append(results, collectAllDescendants(child)...)
		}
	case *Array:
		for _, child := range val.data {
			results = append(results, child)
			results = append(results, collectAllDescendants(child)...)
		}
	}
	return results
}

// getByIndex gets a single element from an Array by index, supporting negative indices.
func getByIndex(arr *Array, index int) Value {
	if arr == nil {
		return nil
	}
	if index < 0 {
		index = len(arr.data) + index
	}
	if index < 0 || index >= len(arr.data) {
		return nil
	}
	return arr.data[index]
}

// evalSlice evaluates a slice operation on an Array.
func evalSlice(arr *Array, seg pathSegment) Value {
	if arr == nil {
		return nil
	}
	n := len(arr.data)
	start := seg.start
	end := seg.end

	if start < 0 {
		start = n + start
	}
	if end == -1 {
		end = n
	} else if end < 0 {
		end = n + end
	}
	if start < 0 {
		start = 0
	}
	if start > n {
		start = n
	}
	if end > n {
		end = n
	}
	if start >= end {
		return NewArray()
	}
	return NewArray(arr.data[start:end]...)
}

// evalUnion evaluates a union (multiple indices) on an Array.
func evalUnion(arr *Array, indices []int) Value {
	if arr == nil {
		return nil
	}
	n := len(arr.data)
	var result []Value
	for _, idx := range indices {
		if idx < 0 {
			idx = n + idx
		}
		if idx >= 0 && idx < n {
			result = append(result, arr.data[idx])
		}
	}
	if len(result) == 0 {
		return nil
	}
	if len(result) == 1 {
		return result[0]
	}
	return NewArray(result...)
}

// expandWildcard returns all child values of a Value as a slice.
func expandWildcard(v Value) []Value {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case *Object:
		result := make([]Value, 0, len(val.data))
		for _, child := range val.data {
			result = append(result, child)
		}
		return result
	case *Array:
		return val.data
	}
	return nil
}

// asObjectSafe returns *Object if v is an Object, nil otherwise.
func asObjectSafe(v Value) *Object {
	if obj, ok := v.(*Object); ok {
		return obj
	}
	return nil
}

// asArraySafe returns *Array if v is an Array, nil otherwise.
func asArraySafe(v Value) *Array {
	if arr, ok := v.(*Array); ok {
		return arr
	}
	return nil
}
