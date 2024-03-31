package bitfield

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal_CompositeOfBitFieldsAndNonNormalInteger(t *testing.T) {
	// Setup
	type compositeFields struct {
		A_u6bits  uint8 `bit:"6"`
		B_u2bits  uint8 `bit:"2"`
		C_Int8    int8
		D_i10bits int16 `bit:"10"`
		E_i6bits  int8  `bit:"6"`
		F_Uint32  uint32
		G_Uint8   uint8
		H_u5bits  uint8 `bit:"5"`
		I_u3bits  uint8 `bit:"3"`
		J_u3bits  uint8 `bit:"3"`
		K_Uint16  uint16
	}
	input := []byte{0b10100101, 0x5A, 0b10110110, 0b01101011, 0x5A, 0xA5, 0x55, 0xAA, 0xF0, 0b10101010, 0xA5, 0x6B, 0xB6}
	want := compositeFields{
		A_u6bits:  0b100101,
		B_u2bits:  0b10,
		C_Int8:    0x5A,
		D_i10bits: -74, // 0b1110110110 (signed)
		E_i6bits:  0b011010,
		F_Uint32:  0xAA55A55A,
		G_Uint8:   0xF0,
		H_u5bits:  0b01010,
		I_u3bits:  0b101,
		J_u3bits:  0b101,
		K_Uint16:  0xB66B,
	}

	// Exercise
	var got compositeFields
	err := Unmarshal(input, &got)

	// Verify
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func TestUnmarshal_PlainIntFields(t *testing.T) {
	// Setup
	type plainFields struct {
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		Uint64 uint64
		Int8   int8
		Int16  int16
		Int32  int32
		Int64  int64
	}

	inputData := []byte{
		0x01,
		0x23, 0x45,
		0x67, 0x89, 0xAB, 0xCD,
		0xEF, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD,
		0x01,
		0x23, 0x45,
		0x67, 0x89, 0xAB, 0xCD,
		0xEF, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD,
	}
	want := plainFields{
		Uint8:  0x01,
		Uint16: 0x4523,
		Uint32: 0xCDAB8967,
		Uint64: 0xCDAB8967452301EF,
		Int8:   0x01,
		Int16:  0x4523,
		Int32:  -844_396_185,               // 0xCDAB8967 (signed)
		Int64:  -3_626_653_998_282_243_601, // 0xCDAB8967452301EF (signed)
	}

	// Exercise
	var got plainFields
	err := Unmarshal(inputData, &got)

	// Verify
	assert.Nil(t, err)
	assert.Equal(t, want, got)
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
	var over64Bits struct {
		A uint64 `bit:"65"`
	}
	var sizeNonNumber struct {
		A uint8 `bit:"x"`
	}
	var nonIntField struct {
		A struct{} `bit:"1"`
	}

	testCases := map[string]struct {
		in  []byte
		out any
	}{
		"Size empty":          {[]byte{0x00}, &sizeEmpty},
		"Size zero":           {[]byte{0x00}, &sizeZero},
		"Size less than zero": {[]byte{0x00}, &sizeLessThanZero},
		"Over type size":      {[]byte{0x00}, &overTypeSize},
		"Over 64 bits":        {[]byte{0x00}, &over64Bits},
		"Size non-number":     {[]byte{0x00}, &sizeNonNumber},
		"Non-int field":       {[]byte{0x00}, &nonIntField},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Exercise
			err := Unmarshal(tc.in, tc.out)

			// Verify
			var fieldError *FieldError
			assert.ErrorAs(t, err, &fieldError)
		})
	}
}

func TestUnmarshalError(t *testing.T) {
	// Setup
	var integer int
	var nilPointer *struct{} = nil
	testCases := map[string]struct {
		in  []byte
		out any
	}{
		"Nil provided":                   {[]byte{0x00}, nil},
		"Non-pointer provided":           {[]byte{0x00}, integer},
		"Pointer to non-struct provided": {[]byte{0x00}, &integer},
		"Nil pointer provided":           {[]byte{0x00}, nilPointer},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Exercise
			err := Unmarshal(tc.in, tc.out)

			// Verify
			var typeError *TypeError
			assert.ErrorAs(t, err, &typeError)
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
			assert.Nil(t, err)
			assert.Equal(t, tc.want, tc.argV.A)
		})
	}
}
