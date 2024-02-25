package bitfield

import (
	"errors"
	"reflect"
)

func Unmarshal(data []byte, v any) error {
	// Check v is a pointer to a struct
	if v == nil {
		return errors.New("v must be a pointer to a struct")
	}
	rt := reflect.TypeOf(v)
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Struct {
		return errors.New("v must be a pointer to a struct")
	}
	return nil
}
