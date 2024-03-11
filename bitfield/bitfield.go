package bitfield

import (
	"encoding/binary"
	"reflect"
	"strconv"
)

// An InvalidTypeError describes an invalid type passed to [Unmarshal].
// (The argument to [Unmarshal] must be a non-nil pointer to a struct.)
type InvalidTypeError struct {
	Type reflect.Type
}

func (e *InvalidTypeError) Error() string {
	if e.Type == nil {
		return "bitfield: invalid type passed (nil)"
	}
	if e.Type.Kind() != reflect.Pointer {
		return "bitfield: invalid type passed (non-pointer " + e.Type.String() + ")"
	}
	if e.Type.Elem().Kind() != reflect.Struct {
		return "bitfield: invalid type passed (pointer to non-struct " + e.Type.String() + ")"
	}
	return "bitfield: invalid type passed (nil " + e.Type.String() + ")"
}

type InvalidFieldError struct {
	Field   reflect.StructField
	problem string
	Err     error
}

func (e *InvalidFieldError) Error() string {
	return "bitfield: " + e.problem + " (" + e.Field.Name + " " + e.Field.Type.String() + " `" + string(e.Field.Tag) + "`)"
}

func (e *InvalidFieldError) Unwrap() error {
	return e.Err
}

func Unmarshal(data []byte, v any, opts ...Option) error {
	if err := validateUnmarshalType(v); err != nil {
		return err
	}
	options, err := collectOptions(opts)
	if err != nil {
		return err
	}
	unmarshal(data, v, options)
	return nil
}

func unmarshal(data []byte, v any, options options) {
	iData := 0
	iBitInData := 0
	rt := reflect.TypeOf(v).Elem()
	for iField := 0; iField < rt.NumField(); iField++ {
		vf := reflect.ValueOf(v).Elem().Field(iField)
		if tag, ok := rt.Field(iField).Tag.Lookup("bit"); ok {
			// Already checked error
			bitSize, _ := strconv.Atoi(tag)
			iData, iBitInData = setValueToBitField(&vf, data, bitSize, iData, iBitInData)
		} else if isInteger(vf.Kind()) {
			setValueToIntegerField(&vf, data[iData:], options)
			iData += int(vf.Type().Size())
		}
	}
}

func setValueToBitField(vf *reflect.Value, data []byte, bitSize, iData, iBitInData int) (int, int) {
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
	return iData, iBitInData
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

func setValueToIntegerField(vf *reflect.Value, data []byte, options options) {
	var byteOrder binary.ByteOrder = binary.LittleEndian
	if options.ByteOrder == BigEndian {
		byteOrder = binary.BigEndian
	}
	switch vf.Kind() {
	case reflect.Uint8:
		vf.SetUint(uint64(data[0]))
	case reflect.Uint16:
		vf.SetUint(uint64(byteOrder.Uint16(data)))
	case reflect.Uint32:
		vf.SetUint(uint64(byteOrder.Uint32(data)))
	case reflect.Uint64:
		vf.SetUint(byteOrder.Uint64(data))
	case reflect.Int8:
		vf.SetInt(int64(int8(data[0])))
	case reflect.Int16:
		vf.SetInt(int64(int16(byteOrder.Uint16(data))))
	case reflect.Int32:
		vf.SetInt(int64(int32(byteOrder.Uint32(data))))
	case reflect.Int64:
		vf.SetInt(int64(byteOrder.Uint64(data)))
	}
}

func isInteger(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func isNonNilPointerToStruct(v any) bool {
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Pointer && !rv.IsNil() && rv.Elem().Kind() == reflect.Struct
}

func validateStruct(v any) error {
	rt := reflect.TypeOf(v).Elem()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if err := validateField(field); err != nil {
			return err
		}
	}
	return nil
}

func validateField(field reflect.StructField) error {
	tag, ok := field.Tag.Lookup("bit")
	if !ok {
		return nil
	}

	bitSize, err := strconv.Atoi(tag)
	if err != nil {
		return &InvalidFieldError{
			Field:   field,
			problem: "bit size must be integer",
			Err:     err,
		}
	}
	if !isInteger(field.Type.Kind()) {
		return &InvalidFieldError{
			Field:   field,
			problem: "bit field must be fixed-size integer type",
		}
	}
	if !(1 <= bitSize && bitSize <= field.Type.Bits()) {
		return &InvalidFieldError{
			Field:   field,
			problem: "bit size must be within range 1 to its type size",
		}
	}
	return nil
}

func validateUnmarshalType(v any) error {
	if !isNonNilPointerToStruct(v) {
		return &InvalidTypeError{reflect.TypeOf(v)}
	}
	return validateStruct(v)
}
