package bitfield

import (
	"errors"
	"reflect"
)

func Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	// Note: reflect.ValueOf(nil).Kind() == reflect.Invalid
	if rv.Kind() != reflect.Ptr || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return errors.New("v must be a non-nil pointer to a struct")
	}

	rt := reflect.TypeOf(v).Elem()
	for i := 0; i < rt.NumField(); i++ {
		tf := rt.Field(i)
		vf := rv.Elem().Field(i)
		if tf.Type.Kind() == reflect.Uint8 {
			vf.SetUint(uint64(data[i]))
		}
		if tf.Type.Kind() == reflect.Uint32 {
			v := uint32(data[i+3])<<24 | uint32(data[i+2])<<16 | uint32(data[i+1])<<8 | uint32(data[i])
			vf.SetUint(uint64(v))
		}
	}

	return nil
}
