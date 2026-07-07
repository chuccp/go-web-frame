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

// Add sets the value for key in the KV.
func (o KV) Add(key string, value any) {
	(o)[key] = value
}
