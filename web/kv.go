// Package web: key-value map type with type-safe accessors.
package web

import "github.com/spf13/cast"

// KV is a string-keyed map with type-safe accessor methods.
type KV map[string]any

// GetString returns the value for key as a string.
func (o KV) GetString(key string) string {
	return cast.ToString((o)[key])
}

// GetInt returns the value for key as an int.
func (o KV) GetInt(key string) int {
	return cast.ToInt((o)[key])
}

// GetIntForDefault returns the value for key as an int, or defaultValue if the key is not present.
func (o KV) GetIntForDefault(key string, defaultValue int) int {
	if _, ok := (o)[key]; ok {
		return o.GetInt(key)
	}
	return defaultValue
}

// GetFloat returns the value for key as a float64.
func (o KV) GetFloat(key string) float64 {
	return cast.ToFloat64((o)[key])
}

// GetBool returns the value for key as a bool.
func (o KV) GetBool(key string) bool {
	return cast.ToBool((o)[key])
}

// GetUint returns the value for key as a uint.
func (o KV) GetUint(key string) uint {
	return cast.ToUint((o)[key])
}

// GetInt64 returns the value for key as an int64.
func (o KV) GetInt64(key string) int64 {
	return cast.ToInt64((o)[key])
}

// GetUint64 returns the value for key as an uint64.
func (o KV) GetUint64(key string) uint64 {
	return cast.ToUint64((o)[key])
}

// GetStringForDefault returns the value for key as a string, or defaultValue if the key is not present.
func (o KV) GetStringForDefault(key string, defaultValue string) string {
	if _, ok := (o)[key]; ok {
		return o.GetString(key)
	}
	return defaultValue
}

// GetBoolForDefault returns the value for key as a bool, or defaultValue if the key is not present.
func (o KV) GetBoolForDefault(key string, defaultValue bool) bool {
	if _, ok := (o)[key]; ok {
		return o.GetBool(key)
	}
	return defaultValue
}

// GetMap returns the value for key as a KV (map), or nil.
func (o KV) GetMap(key string) KV {
	v, ok := (o)[key]
	if !ok {
		return nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	return KV(m)
}

// GetArray returns the value for key as a slice, or nil.
func (o KV) GetArray(key string) []any {
	v, ok := (o)[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	return arr
}

// GetStrings returns the value for key as a []string, or nil.
func (o KV) GetStrings(key string) []string {
	v, ok := (o)[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		result = append(result, cast.ToString(item))
	}
	return result
}

// Has returns true if the key exists in the KV.
func (o KV) Has(key string) bool {
	_, ok := (o)[key]
	return ok
}

// Add sets the value for key in the KV.
func (o KV) Add(key string, value any) {
	(o)[key] = value
}
