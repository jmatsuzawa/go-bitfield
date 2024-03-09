package bitfield

import (
	"encoding/binary"
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
			bitSize, _ := strconv.Atoi(tag)
			var val uint64

			i := 0
			for i < bitSize && iData < len(data) {
				d := uint64(data[iData])
				for ; iBitInData < 8 && i < bitSize; iBitInData, i = iBitInData+1, i+1 {
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
				vf.SetInt(signed(val, bitSize))
			}
		} else if isInteger(tf) {
			switch tf.Type.Kind() {
			case reflect.Uint8:
				vf.SetUint(uint64(data[iData]))
			case reflect.Uint16:
				vf.SetUint(uint64(binary.LittleEndian.Uint16(data[iData:])))
			case reflect.Uint32:
				vf.SetUint(uint64(binary.LittleEndian.Uint32(data[iData:])))
			case reflect.Uint64:
				vf.SetUint(binary.LittleEndian.Uint64(data[iData:]))
			case reflect.Int8:
				vf.SetInt(int64(int8(data[iData])))
			case reflect.Int16:
				vf.SetInt(int64(int16(binary.LittleEndian.Uint16(data[iData:]))))
			case reflect.Int32:
				vf.SetInt(int64(int32(binary.LittleEndian.Uint32(data[iData:]))))
			case reflect.Int64:
				vf.SetInt(int64(binary.LittleEndian.Uint64(data[iData:])))
			}
			iData += int(tf.Type.Size())
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
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func validateStruct(v any) error {
	rt := reflect.TypeOf(v).Elem()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag, ok := field.Tag.Lookup("bit")
		if !ok {
			continue
		}

		bitSize, err := strconv.Atoi(tag)
		if err != nil {
			return err
		}
		if !isInteger(field) {
			return errors.New("bit field must be an integer type")
		}
		if bitSize <= 0 {
			return errors.New("bit length must be greater than 0")
		}
		if bitSize > field.Type.Bits() {
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
 * For example, signed(val = 0b00101101, bitSize = 6) returns 0b11101101
 */
func signed(val uint64, bitSize int) int64 {
	msb := val >> (bitSize - 1)
	pattern := (0 - msb) << bitSize
	return int64(val | pattern)
}
