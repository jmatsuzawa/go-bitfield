package bitfield_test

import (
	"fmt"

	"github.com/jmatsuzawa/go-bitfield"
)

func ExampleUnmarshal() {
	var out struct {
		A uint8  `bit:"4"`
		B uint8  `bit:"8"`
		C uint32 `bit:"20"`
	}
	input := []byte{0x12, 0x34, 0x56, 0x78}

	_ = bitfield.Unmarshal(input, &out)
	fmt.Printf("A=%#x, B=%#x, C=%#x\n", out.A, out.B, out.C)
	// Output: A=0x2, B=0x41, C=0x78563
}

// func ExampleUnmarshal() {
// 	var out struct {
// 		A uint8 `bit:"4"`
// 		B uint8 `bit:"4"`
// 		C uint16
// 		D uint8 `bit:"5"`
// 		E uint8 `bit:"2"`
// 		F uint8 `bit:"1"`
// 		G int16 `bit:"10"`
// 		H int8  `bit:"6"`
// 	}
// 	input := []byte{
// 		0b0101_0100, // A=0b0100 B=0b0101 (Least significant bit first )
// 		0x01, 0x23,  // C=0x2301 (Little endian by default for plain integer fields)
// 		0b1_10_11010,            // D=0b11010 E=0b10 F=0b1
// 		0b00000000, 0b100000_10, // G=0b1000000000 (-512) H=0b100000 (-32)
// 	}

// 	_ = bitfield.Unmarshal(input, &out)
// 	fmt.Printf("A=%#04b, B=%#04b, C=%#x, D=%#05b, E=%#02b, F=%#b, G=%d, H=%d\n",
// 		out.A, out.B, out.C, out.D, out.E, out.F, out.G, out.H)
// 	// Output: A=0b0100, B=0b0101, C=0x2301, D=0b11010, E=0b10, F=0b1, G=-512, H=-32
// }
