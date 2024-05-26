package bitfield

import (
	"encoding/binary"
	"reflect"
	"strconv"
)

// Unmarshal decodes the byte slice data and stores the result in a struct with bit-fields pointed by out.
//
// The bit size of bit-fields of a struct is specified by a struct tag "bit". See the following example code
// The bit size must be within the range 1 to the size of the field type. For example, uint8 A `bit:"9"` is not acceptable, which causes FieldError.
//
// You can define both plain integer fields without a bit tag and bit-fields with a bit tag in a struct. Fields of not-integer types are ignored by Unmarshal.
//
// If out is not a non-nil pointer to a struct, Unmarshal returns a TypeError.
func Unmarshal(data []byte, out any, opts ...Option) error {
	if err := validateUnmarshalType(out); err != nil {
		return err
	}
	options, err := collectOptions(opts)
	if err != nil {
		return err
	}
	unmarshal(data, out, options)
	return nil
}

func unmarshal(data []byte, out any, options options) {
	iData := 0
	iBitInData := 0
	rt := reflect.TypeOf(out).Elem()
	for iField := 0; iField < rt.NumField(); iField++ {
		vf := reflect.ValueOf(out).Elem().Field(iField)
		if tag, ok := rt.Field(iField).Tag.Lookup("bit"); ok {
			// Already checked error
			bitSize, _ := strconv.Atoi(tag)
			iData, iBitInData = setValueToBitField(&vf, data, bitSize, iData, iBitInData, rt.Field(iField).IsExported(), options)
		} else if isFixedInteger(vf.Kind()) {
			// If the previous field is not fully read, the next plain integer field should be read from the next byte
			if iBitInData > 0 {
				iData++
				iBitInData = 0
			}
			if rt.Field(iField).IsExported() {
				setValueToIntegerField(&vf, data[iData:], options)
			}
			iData += int(vf.Type().Size())
		}
	}
}

func setValueToBitFieldLittleEndian(vf *reflect.Value, data []byte, bitSize, iData, iBitInData int, isExported bool) (int, int) {
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
	if isExported {
		if vf.CanUint() {
			vf.SetUint(val)
		} else if vf.CanInt() {
			vf.SetInt(signed(val, bitSize))
		}
	}
	return iData, iBitInData
}

func setValueToBitFieldBigEndian(vf *reflect.Value, data []byte, bitSize, iData, iBitInData int, isExported bool) (int, int) {
	var val uint64

	for consumedBits := 0; consumedBits < bitSize && iData < len(data); {
		remainedBitInThisByte := 8 - iBitInData
		var wantBitInThisByte int
		if (bitSize - consumedBits) < remainedBitInThisByte {
			wantBitInThisByte = bitSize - consumedBits
		} else {
			wantBitInThisByte = remainedBitInThisByte
		}

		var mask byte = 0xff >> (8 - wantBitInThisByte)
		var b byte = data[iData] >> iBitInData
		consumedBits += wantBitInThisByte
		val = (val << wantBitInThisByte) | uint64(b&mask)
		iBitInData += wantBitInThisByte
		if iBitInData >= 8 {
			iData++
			iBitInData = 0
		}
	}
	if isExported {
		if vf.CanUint() {
			vf.SetUint(val)
		} else if vf.CanInt() {
			vf.SetInt(signed(val, bitSize))
		}
	}
	return iData, iBitInData
}

func setValueToBitField(vf *reflect.Value, data []byte, bitSize, iData, iBitInData int, isExported bool, options options) (int, int) {
	if options.byteOrder == LittleEndian {
		return setValueToBitFieldLittleEndian(vf, data, bitSize, iData, iBitInData, isExported)
	} else {
		return setValueToBitFieldBigEndian(vf, data, bitSize, iData, iBitInData, isExported)
	}
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
	if options.byteOrder == BigEndian {
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

func isFixedInteger(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func ensureNonNilPointerToStruct(v any) error {
	errMsg := "de/encoded object must be non-nil pointer to struct"
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Invalid {
		return &TypeError{
			Type:    reflect.TypeOf(v),
			problem: errMsg + " (nil passed)",
		}
	}
	if rv.Kind() != reflect.Pointer {
		return &TypeError{
			Type:    reflect.TypeOf(v),
			problem: errMsg + " (" + rv.Type().String() + " passed)",
		}
	}
	if rv.IsNil() {
		return &TypeError{
			Type:    reflect.TypeOf(v),
			problem: errMsg + " (nil " + rv.Type().String() + " passed)",
		}
	}
	if rv.Elem().Kind() != reflect.Struct {
		return &TypeError{
			Type:    reflect.TypeOf(v),
			problem: errMsg + " (" + rv.Type().String() + " passed)",
		}
	}
	return nil
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
		return &FieldError{
			Field:   field,
			problem: "bit size must be integer",
		}
	}
	if !isFixedInteger(field.Type.Kind()) {
		return &FieldError{
			Field:   field,
			problem: "bit field must be fixed-size integer type",
		}
	}
	if !(1 <= bitSize && bitSize <= field.Type.Bits()) {
		return &FieldError{
			Field:   field,
			problem: "bit size must be within range 1 to its type size",
		}
	}
	return nil
}

func validateUnmarshalType(v any) error {
	if err := ensureNonNilPointerToStruct(v); err != nil {
		return err
	}
	return validateStruct(v)
}
