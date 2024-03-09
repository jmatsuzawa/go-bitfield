package bitfield

import (
	"errors"
	"reflect"
	"strconv"
)

// An InvalidUnmarshalError describes an invalid argument passed to [Unmarshal].
// (The argument to [Unmarshal] must be a non-nil pointer to a struct.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "go-bitfield: Unmarshal(nil)"
	}
	if e.Type.Kind() != reflect.Pointer {
		return "go-bitfield: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	if e.Type.Elem().Kind() != reflect.Struct {
		return "go-bitfield: Unmarshal(pointer to non-struct " + e.Type.String() + ")"
	}
	return "go-bitfield: Unmarshal(nil " + e.Type.String() + ")"
}

func Unmarshal(data []byte, v any) error {
	if err := validateUnmarshalType(v); err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	iData := 0
	iBitInData := 0
	rt := reflect.TypeOf(v).Elem()
	for iField := 0; iField < rt.NumField(); iField++ {
		tf := rt.Field(iField)
		vf := rv.Elem().Field(iField)
		if tag, ok := tf.Tag.Lookup("bit"); ok {
			// Already checked error
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
				iData++
			} else if tf.Type.Kind() == reflect.Uint32 {
				v := uint32(data[iData+3])<<24 | uint32(data[iData+2])<<16 | uint32(data[iData+1])<<8 | uint32(data[iData])
				vf.SetUint(uint64(v))
				iData += 4
			} else if tf.Type.Kind() == reflect.Int8 {
				vf.SetInt(int64(int8(data[iData])))
				iData++
			}
		}
	}

	return nil
}

func isNonNilPointerToStruct(v any) bool {
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Pointer && !rv.IsNil() && rv.Elem().Kind() == reflect.Struct
}

func isInteger(field reflect.StructField) bool {
	switch field.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func validateStruct(v any) error {
	rt := reflect.TypeOf(v).Elem()
	for iField := 0; iField < rt.NumField(); iField++ {
		field := rt.Field(iField)
		tag, ok := field.Tag.Lookup("bit")
		if !ok {
			continue
		}

		bitLen, err := strconv.Atoi(tag)
		if err != nil {
			return err
		}
		if !isInteger(field) {
			return errors.New("bit field must be an integer type")
		}
		if bitLen <= 0 {
			return errors.New("bit length must be greater than 0")
		}
		if bitLen > field.Type.Bits() {
			return errors.New("bit length must be less than or equal to the type size")
		}
	}
	return nil
}

func validateUnmarshalType(v any) error {
	if !isNonNilPointerToStruct(v) {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	return validateStruct(v)
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
