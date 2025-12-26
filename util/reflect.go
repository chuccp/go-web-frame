package util

import (
	"reflect"
)

func NewPtr[T any](v T) T {
	type_ := reflect.TypeOf(v)
	switch type_.Kind() {
	case reflect.Ptr:
		return newPtr(type_.Elem()).(T)
	case reflect.Struct:
		return newPtr(type_).(T)
	default:
		panic("unhandled default case")
	}
}

func NewSlice[T any](v T) []T {
	elemType := reflect.TypeOf(v)
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	ptrType := reflect.PointerTo(elemType)
	sliceType := reflect.SliceOf(ptrType)
	slice := reflect.MakeSlice(sliceType, 0, 0)
	return slice.Interface().([]T)
}

func newPtr(type_ reflect.Type) any {
	value := reflect.New(type_)
	u := value.Interface()
	return u
}

func GetStructFullName(v any) string {
	if v == nil {
		return "<nil>"
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.String() // 带包路径和指针符号
}
func GetStructName(v any) string {
	if v == nil {
		return "<nil>"
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
