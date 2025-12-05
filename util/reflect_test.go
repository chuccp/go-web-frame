package util

import "testing"

func TestNewPtr(t *testing.T) {
	type User struct {
		Id   uint
		Name string
	}
	user := NewSlice(&User{}).([]*User)
	t.Log(user)
}
