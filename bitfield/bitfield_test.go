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
		})
	}
}
