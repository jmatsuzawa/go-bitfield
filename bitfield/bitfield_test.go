package bitfield

import (
	"testing"
)

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
