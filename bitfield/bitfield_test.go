package bitfield

import (
	"testing"
)

func TestUnmarshalPlainStruct(t *testing.T) {
	// Setup
	var v struct {
		N uint8
	}

	// Exercise
	err := Unmarshal([]byte{0xA5}, &v)

	// Verify
	if err != nil {
		t.Errorf("Unmarshal() = %v; want nil", err)
	}

	if v.N != 0xA5 {
		t.Errorf("Unmarshal() = %v; want 0xA5", v.N)
	}
}

func TestUnmarshalPlainStructUint32(t *testing.T) {
	// Setup
	var v struct {
		N uint32
	}

	// Exercise
	err := Unmarshal([]byte{0x01, 0x23, 0x45, 0x67}, &v)

	// Verify
	if err != nil {
		t.Errorf("Unmarshal() = %v; want nil", err)
	}

	if v.N != 0x67452301 {
		t.Errorf("Unmarshal() = %v; want 0x67452301", v.N)
	}
}

func TestUnmarshalSimpleBitsSingleByte(t *testing.T) {
	// Setup
	var v struct {
		OneBit   uint8 `bit:"1"`
		TwoBits  uint8 `bit:"2"`
		FiveBits uint8 `bit:"5"`
	}

	// Exercise
	err := Unmarshal([]byte{0b11101101}, &v)

	// Verify
	if err != nil {
		t.Fatalf("Unmarshal() = %v; want nil", err)
	}
	if v.OneBit != 0b1 {
		t.Fatalf("Unmarshal() -> v.OneBit = %+b; want 1", v.OneBit)
	}
	if v.TwoBits != 0b10 {
		t.Fatalf("Unmarshal() -> v.TwoBits = %+b; want 0b10", v.TwoBits)
	}
	if v.FiveBits != 0b11101 {
		t.Fatalf("Unmarshal() -> v.FiveBits = %+b; want 0b11101", v.FiveBits)
	}
}

func TestUnmarshal9bits(t *testing.T) {
	// Setup
	var v struct {
		NineBits uint16 `bit:"9"`
	}

	// Exercise
	err := Unmarshal([]byte{0xa5, 0x01}, &v)

	// Verify
	if err != nil {
		t.Fatalf("Unmarshal() = %v; want nil", err)
	}
	if v.NineBits != 0x1a5 {
		t.Fatalf("Unmarshal() -> v.NineBits = 0x%x; want 0x1a5", v.NineBits)
	}
}

func TestUnmarshalSignedInt(t *testing.T) {
	// Setup
	var v struct {
		OneBit   uint8 `bit:"1"`
		TwoBits  uint8 `bit:"2"`
		FiveBits int8  `bit:"5"`
	}

	// Exercise
	err := Unmarshal([]byte{0b11101101}, &v)

	// Verify
	if err != nil {
		t.Fatalf("Unmarshal() = %v; want nil", err)
	}
	if v.OneBit != 0b1 {
		t.Fatalf("Unmarshal() -> v.OneBit = %+b; want 1", v.OneBit)
	}
	if v.TwoBits != 0b10 {
		t.Fatalf("Unmarshal() -> v.TwoBits = %+b; want 0b10", v.TwoBits)
	}
	if v.FiveBits != -3 {
		t.Fatalf("Unmarshal() -> v.FiveBits = %d; want -3", v.FiveBits)
	}
}

func TestUnmarshalPlainInteger(t *testing.T) {
	// Setup
	var v struct {
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		Uint64 uint64
		Int8   int8
		Int16  int16
		Int32  int32
		Int64  int64
	}

	// Exercise
	err := Unmarshal([]byte{
		0x01,
		0x23, 0x45,
		0x67, 0x89, 0xAB, 0xCD,
		0xEF, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD,
		0x01,
		0x23, 0x45,
		0x67, 0x89, 0xAB, 0xCD,
		0xEF, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD,
	}, &v)

	if err != nil {
		t.Fatalf("Unmarshal() = %v; want nil", err)
	}

	if v.Uint8 != 0x01 {
		t.Errorf("Unmarshal() -> v.Uint8 = %#x; want 0x01", v.Uint8)
	}
	if v.Uint16 != 0x4523 {
		t.Errorf("Unmarshal() -> v.Uint16 = %#x; want 0x4523", v.Uint16)
	}
	if v.Uint32 != 0xCDAB8967 {
		t.Errorf("Unmarshal() -> v.Uint32 = %#x; want 0xCDAB8967", v.Uint32)
	}
	if v.Uint64 != 0xCDAB8967452301EF {
		t.Errorf("Unmarshal() -> v.Uint64 = %#x; want 0xCDAB8967452301EF", v.Uint64)
	}
	if v.Int8 != int8(0x01) {
		t.Errorf("Unmarshal() -> uint8(v.Int8) = %#x; want 0x01", v.Int8)
	}
	if v.Int16 != int16(0x4523) {
		t.Errorf("Unmarshal() -> uint16(v.Int16) = %#x; want 0x4523", v.Int16)
	}
	// if uint32(v.Int32) != 0xCDAB8967 {
	if v.Int32 != -844_396_185 { // 0xCDAB8967
		t.Errorf("Unmarshal() -> uint32(v.Int32) = %#x; want 0xCDAB8967", v.Int32)
	}
	// if uint64(v.Int64) != 0xCDAB8967452301EF {
	if v.Int64 != -3_626_653_998_282_243_601 { // 0xCDAB8967452301EF
		t.Errorf("Unmarshal() -> uint64(v.Int64) = %#x; want 0xCDAB8967452301EF", v.Int64)
	}
}

func TestUnmarshalCompositeOfBitFieldsAndNonNormalInteger(t *testing.T) {
	// Setup
	var v struct {
		BitA    uint8 `bit:"6"`
		BitB    uint8 `bit:"2"`
		Int8C   int8
		BitD    int16 `bit:"10"`
		BitE    int8  `bit:"6"`
		Uint32F uint32
		Uint8G  uint8
		BitH    uint8 `bit:"5"`
		BitI    uint8 `bit:"3"`
	}

	// Exercise
	err := Unmarshal([]byte{0b10100101, 0x5A, 0xB6, 0x6B, 0x5A, 0xA5, 0x55, 0xAA, 0xF0, 0b10101010}, &v)

	// Verify
	if err != nil {
		t.Fatalf("Unmarshal() = %v; want nil", err)
	}
	if v.BitA != 0b100101 {
		t.Fatalf("Unmarshal() -> v.BitA = %#b; want 0b100101", v.BitA)
	}
	if v.BitB != 0b10 {
		t.Fatalf("Unmarshal() -> v.BitB = %#b; want 0b10", v.BitB)
	}
	if v.Int8C != 0x5A {
		t.Fatalf("Unmarshal() -> v.Int8C = %#x; want -3", v.Int8C)
	}
	if v.BitD != -74 {
		t.Fatalf("Unmarshal() -> v.BitD = %d (%#x); want -74", v.BitD, v.BitD)
	}
	if v.BitE != 0b011010 {
		t.Fatalf("Unmarshal() -> v.BitE = %#b; want 0b011010", v.BitE)
	}
	if v.Uint32F != 0xAA55A55A {
		t.Fatalf("Unmarshal() -> v.Uint32F = %#x; want 0xaa55a55a", v.Uint32F)
	}
	if v.Uint8G != 0xf0 {
		t.Fatalf("Unmarshal() -> v.Uint8G = %#x; want 0xf0", v.Uint8G)
	}
	if v.BitH != 0b01010 {
		t.Fatalf("Unmarshal() -> v.BitH = %#b; want 0b10101", v.BitH)
	}
	if v.BitI != 0b101 {
		t.Fatalf("Unmarshal() -> v.BitI = %#b; want 0b101", v.BitI)
	}
}

func TestUnmarshalBitSizeLimitError(t *testing.T) {
	// Setup
	var sizeEmpty struct {
		A uint8 `bit:""`
	}
	var sizeZero struct {
		A uint8 `bit:"0"`
	}
	var sizeLessThanZero struct {
		A uint8 `bit:"-1"`
	}
	var overTypeSize struct {
		A uint8 `bit:"9"`
	}
	var sizeNonNumber struct {
		A uint8 `bit:"x"`
	}
	var nonIntField struct {
		A struct{} `bit:"1"`
	}

	testCases := map[string]struct {
		argData []byte
		argV    any
	}{
		"size empty":          {[]byte{0x00}, &sizeEmpty},
		"size zero":           {[]byte{0x00}, &sizeZero},
		"size less than zero": {[]byte{0x00}, &sizeLessThanZero},
		"over type size":      {[]byte{0x00}, &overTypeSize},
		"size non-number":     {[]byte{0x00}, &sizeNonNumber},
		"non-int field":       {[]byte{0x00}, &nonIntField},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Exercise
			err := Unmarshal(tc.argData, tc.argV)
			// Verify
			if err == nil {
				t.Errorf("Unmarshal() = %v; want error", err)
			}
		})
	}
}

func TestUnmarshalError(t *testing.T) {
	var integer int
	var nilPointer *struct{} = nil
	testCases := map[string]struct {
		argData []byte
		argV    any
	}{
		"Nil provided":                   {[]byte{0x00}, nil},
		"Non-pointer provided":           {[]byte{0x00}, integer},
		"Pointer to non-struct provided": {[]byte{0x00}, &integer},
		"Nil pointer provided":           {[]byte{0x00}, nilPointer},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := Unmarshal(tc.argData, tc.argV)
			if err == nil {
				t.Errorf("Unmarshal() = %v; want error", err)
			}
			switch err := err.(type) {
			case *InvalidUnmarshalError:
			default:
				t.Errorf("Unmarshal() = %s; want InvalidUnmarshalError", err)
			}
		})
	}
}

func TestUnmarshalByteOrder(t *testing.T) {
	// Setup
	type a struct{ A uint32 }
	testCases := map[string]struct {
		argData []byte
		argV    a
		argOpts []Option
		want    uint32
	}{
		"LittleEndian": {
			argData: []byte{0x01, 0x23, 0x45, 0x67},
			argV:    a{},
			argOpts: []Option{WithByteOrder(LittleEndian)},
			want:    0x67452301,
		},
		"BigEndian": {
			argData: []byte{0x01, 0x23, 0x45, 0x67},
			argV:    a{},
			argOpts: []Option{WithByteOrder(BigEndian)},
			want:    0x01234567,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Exercise
			err := Unmarshal(tc.argData, &tc.argV, tc.argOpts...)

			// Verify
			if err != nil {
				t.Fatalf("Unmarshal() = %v; want nil", err)
			}
			if tc.argV.A != tc.want {
				t.Errorf("Unmarshal() -> v.A = %#x; want %#x", tc.argV.A, tc.want)
			}
		})
	}
}
