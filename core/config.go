package core

import (
	"github.com/chuccp/go-web-frame/log"
)

// UnmarshalConfigKey unmarshals configuration under the given key into the specified type.
// T is typically a pointer to a struct (e.g. *ServerConfig) or a value type.
//
// Example:
//
//	cfg, err := core.UnmarshalConfigKey[*ServerConfig]("server", ctx)
func UnmarshalConfigKey[T any](key string, ctx *Context) (T, error) {
	var t T
	err := ctx.GetConfig().UnmarshalKey(key, &t)
	if err != nil {
		return t, err
	}
	return t, nil
}

// MustUnmarshalConfigKey is like UnmarshalConfigKey but panics if an error occurs.
// Useful during initialization when config loading failures should be fatal.
func MustUnmarshalConfigKey[T any](key string, ctx *Context) T {
	t, err := UnmarshalConfigKey[T](key, ctx)
	if err != nil {
		log.PanicErrors("MustUnmarshalConfigKey", err)
	}
	return t
}

// UnmarshalConfig unmarshals the entire configuration into the specified type.
// T is typically a pointer to a struct representing the full config schema.
//
// Example:
//
//	cfg, err := core.UnmarshalConfig[*AppConfig](ctx)
func UnmarshalConfig[T any](ctx *Context) (T, error) {
	var t T
	err := ctx.GetConfig().Unmarshal(&t)
	if err != nil {
		return t, err
	}
	return t, nil
}

// MustUnmarshalConfig is like UnmarshalConfig but panics if an error occurs.
// Useful during initialization when config loading failures should be fatal.
func MustUnmarshalConfig[T any](ctx *Context) T {
	t, err := UnmarshalConfig[T](ctx)
	if err != nil {
		log.PanicErrors("MustUnmarshalConfig", err)
	}
	return t
}

// UnmarshalKeyConfig is an alias for UnmarshalConfigKey, provided for backward compatibility.
// Prefer UnmarshalConfigKey in new code.
func UnmarshalKeyConfig[T any](key string, c *Context) (T, error) {
	return UnmarshalConfigKey[T](key, c)
}
