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

func TestUnmarshalError(t *testing.T) {
	var integer int
	testCases := map[string]struct {
		argData []byte
		argV    any
	}{
		"Nil provided":                   {[]byte{0x00}, nil},
		"Non-pointer provided":           {[]byte{0x00}, integer},
		"Pointer to non-struct provided": {[]byte{0x00}, &integer},
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
