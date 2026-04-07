package util

import (
	"testing"
)

func GetValue[T any]() T {
	var t T

	return NewPtr(t)
}

func TestNewPtr(t *testing.T) {
	type User struct {
		Id   uint
		Name string
	}
	t.Log(GetValue[*User]())

}
