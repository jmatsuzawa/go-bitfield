# go-bitfield: Declarative bit-fields decoder for Go

## Description

Go library for simple declarative decoding of bit-fields.

## Usage

The following is a simple example:

```go
// Bit-fields definitions
type bitFields struct {
    // List fields in order from least significant bit
    A uint8 `bit:"1"`
    B uint8 `bit:"2"`
    _ uint8 `bit:"1"` // For place holder
    C uint8 `bit:"4"`
}

// Byte slice to parse and decode
input := []byte{0b1010_0_10_1} // 0xA5

// Variable of the bit-fields to store the result of decoding
var out bitFields

_ = bitfield.Unmarshal(input, &out)

fmt.Printf("A=%#b, B=%#b, C=%#b\n", out.A, out.B, out.C)
// Output: A=0b1, B=0b10, C=0b1010
```

Bit-fields and their bit sizes are specified by a `bit` struct tag. `bit:"N"` means that the field will parse N bits from the input byte slice.

For more details, refer to the [the documents in pkg.go.dev](https://pkg.go.dev/github.com/jmatsuzawa/go-bitfield).

## Installation

```console
go get github.com/jmatsuzawa/go-bitfield
```

## TODO

The following is a part of the TODO list:

* Marshaling (Encoding) bit-fields to a byte slice
* Streaming Encoders and Decoders

## Licensing

MIT License.

## Authors

* Jiro Matsuzawa

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
