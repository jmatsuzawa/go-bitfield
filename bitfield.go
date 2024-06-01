package bitfield

import (
	"reflect"
	"strconv"
)

// Unmarshal parses a byte slice and stores the result in a struct with bit-fields pointed by out.
//
// The following is a simple example:
//
//	// Bit-fields definitions
//	type bitFields struct {
//	    // List fields from least significant bit
//	    A uint8 `bit:"1"`
//	    B uint8 `bit:"2"`
//	    _ uint8 `bit:"1"` // For place holder
//	    C uint8 `bit:"4"`
//	}
//
//	// Byte slice to parse and decode
//	input := []byte{0b1010_0_10_1} // 0xA5
//
//	// Variable of the bit-fields to store the result of decoding
//	var out bitFields
//
//	_ = bitfield.Unmarshal(input, &out)
//
//	fmt.Printf("A=%#b, B=%#b, C=%#b\n", out.A, out.B, out.C)
//	// Output: A=0b1, B=0b10, C=0b1010
//
// Bit-fields can be declared with a struct tag "bit" as in the above example. The constant number following `bit:` represents the bit size of the field. The bit size must be within the range of 1 to the size of the underlying integer type. For example, uint8 A `bit:"9"` is not acceptable, which causes [FieldError] to be returned. Fields must be listed in order, starting from the least significant bit.
//
// This library borrows the idea of bit-fields from the C language. The function [Unmarshal] is aimed to make it easy to create an instance of a struct with bit-fields from a byte slice in a declarative way, just like type casting of a byte array into a struct pointer in C. However, there are some differences between this package and C language:
//
//   - Bit-fields are not C-like packed bit-fields. Their actual size is the same as the size of their underlying type. For example, Field uint8 `bit:"4"` occupies 8 bits (not 4 bits) in a struct. The bit tag is used to specify the bit position to parse in the provided byte slice. Bit-fields does not reduce memory usage. Data packing is not the purpose of this package.
//
// [Unmarshal] parses multi-byte data in LittleEndian by default. You can change the byte order by specifying [WithByteOrder] option. Example:
//
//	var out struct {
//		A uint8  `bit:"4"`
//		B uint8  `bit:"8"`
//		C uint32 `bit:"20"`
//	}
//	data := []byte{0x06, 0x90, 0x95, 0xC4}
//	_ = bitfield.Unmarshal(data, &out, bitfield.WithByteOrder(bitfield.BigEndian))
//	fmt.Printf("A=%d, B=%d, C=%#x\n", out.A, out.B, out.C)
//	// Output: "A=6, B=0, C=0x995c4"
//
// The provided struct can also have plain integer fields without a bit tag. If an integer field does not have a bit tag, the bit size of the field will be the size of the type. The difference between bit-fields and plain integer fields is that bit-fields parse from the bit following the last parsed bit, while plain integer fields always parse from the LSB of the next byte. The following example code demonstrates the difference:
//
//	type withBitTag struct {
//		A uint8  `bit:"4"`
//		B uint8  `bit:"8"`
//	}
//	var out withBitTag
//	data := []byte{0x55, 0xaa}
//	_ = bitfield.Unmarshal(data, &out)
//	fmt.Printf("A=%#x, B=%#x\n", out.A, out.B)
//	// Output: "A=0x5, B=0xa5"
//
//	type withoutBitTag struct {
//		A uint8  `bit:"4"`
//		B uint8
//	}
//	var out withoutBitTag
//	data := []byte{0x55, 0xaa}
//	_ = bitfield.Unmarshal(data, &out)
//	fmt.Printf("A=%#x, B=%#x\n", out.A, out.B)
//	// Output: "A=0x5, B=0xaa"
//
// If out is not a non-nil pointer to a struct, Unmarshal returns [TypeError].
//
// opts is a variadic parameter to specify how to parse the byte slice. Currently, only [WithByteOrder] option is available to specify the byte order for multi-byte fields.
//
// Paramters:
//
//   - data: A byte slice to parse
//   - out: A non-nil pointer to a struct with bit-fields which stores the result of parsing
//   - opts: Options to specify how to parse the byte slice
//
// Returns:
//
//   - nil if the byte slice is successfully parsed and stored in the struct
//   - [FieldError] if the struct pointed by out has an invalid bit-field
//   - [TypeError] if out is not a non-nil pointer to a struct
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
