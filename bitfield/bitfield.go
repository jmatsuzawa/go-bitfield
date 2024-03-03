package bitfield

import (
	"errors"
	"reflect"
	"strconv"
)

func Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	// Note: reflect.ValueOf(nil).Kind() == reflect.Invalid
	if rv.Kind() != reflect.Ptr || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return errors.New("v must be a non-nil pointer to a struct")
	}

	iData := 0
	iBitInData := 0
	rt := reflect.TypeOf(v).Elem()
	for iField := 0; iField < rt.NumField(); iField++ {
		tf := rt.Field(iField)
		vf := rv.Elem().Field(iField)
		if tag, ok := tf.Tag.Lookup("bit"); ok {
			bitLen, _ := strconv.Atoi(tag)
			var val uint64

			i := 0
			for i < bitLen && iData < len(data) {
				d := uint64(data[iData])
				for ; iBitInData < 8 && i < bitLen; iBitInData, i = iBitInData+1, i+1 {
					val |= (((d >> iBitInData) & 1) << i)
				}
				if iBitInData >= 8 {
					iData++
					iBitInData = 0
				}
			}

			if vf.CanUint() {
				vf.SetUint(val)
			} else if vf.CanInt() {
				vf.SetInt(signed(val, bitLen))
			}
		} else {
			if tf.Type.Kind() == reflect.Uint8 {
				vf.SetUint(uint64(data[iData]))
			}
			if tf.Type.Kind() == reflect.Uint32 {
				v := uint32(data[iData+3])<<24 | uint32(data[iData+2])<<16 | uint32(data[iData+1])<<8 | uint32(data[iData])
				vf.SetUint(uint64(v))
			}
		}
	}

	return nil
}

/**
 * Convert an unsigned integer with a specific bit length to a signed integer
 * For example, signed(val = 0b00101101, bitLen = 6) returns 0b11101101
 */
func signed(val uint64, bitLen int) int64 {
	msb := val >> (bitLen - 1)
	pattern := (0 - msb) << bitLen
	return int64(val | pattern)
}
