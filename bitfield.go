package bitfield

import (
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
	byteOrder := options.byteOrder
	for iField := 0; iField < rt.NumField(); iField++ {
		vf := reflect.ValueOf(out).Elem().Field(iField)
		var bitSize int
		if tag, ok := rt.Field(iField).Tag.Lookup("bit"); ok {
			// Already checked error
			bitSize, _ = strconv.Atoi(tag)
		} else if isFixedInteger(rt.Field(iField).Type.Kind()) {
			bitSize = rt.Field(iField).Type.Bits()
			// If the previous field is not fully read, the next plain integer field should be read from the next byte
			if iBitInData > 0 {
				iData++
				iBitInData = 0
			}
		} else {
			// Ignore non-integer fields
			continue
		}
		var val uint64
		val, iData, iBitInData = parseValue(data, bitSize, iData, iBitInData, byteOrder)

		if rt.Field(iField).IsExported() {
			if vf.CanUint() {
				vf.SetUint(val)
			} else if vf.CanInt() {
				vf.SetInt(signed(val, bitSize))
			}
		}
	}
}

func parseValue(
	data []byte,
	bitSize, iData, iBitInData int,
	byteOrder ByteOrder,
) (val uint64, nextIData, nextIBitInData int) {
	if byteOrder == LittleEndian {
		return parseValueLittleEndian(data, bitSize, iData, iBitInData)
	} else {
		return parseValueBigEndian(data, bitSize, iData, iBitInData)
	}
}

func parseValueLittleEndian(
	data []byte,
	bitSize, iData, iBitInData int,
) (val uint64, nextIData, nextIBitInData int) {
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
	nextIData = iData
	nextIBitInData = iBitInData
	return val, nextIData, nextIBitInData
}

func parseValueBigEndian(
	data []byte,
	bitSize, iData, iBitInData int,
) (val uint64, nextIData, nextIBitInData int) {
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
	nextIData = iData
	nextIBitInData = iBitInData
	return val, nextIData, nextIBitInData
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
