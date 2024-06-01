package bitfield

type ByteOrder int

// ByteOrder is an enumeration type that represents the byte order of binary data.
// LittleEndian and BigEndian are the two possible values, representing little-endian and big-endian byte order respectively.
const (
	LittleEndian ByteOrder = iota
	BigEndian
)

type options struct {
	byteOrder ByteOrder
}

type Option func(*options) error

// WithByteOrder specifies the byte order in which Unmarshal parses multi-byte data in a byte slice
//
// Examples of usage:
//
//	// For little-endian:
//	Unmarshal(data, out, WithByteOrder(LittleEndian))
//	// Little-endian is the default byte order. It can be omitted. The following is equivalent to the above.
//	Unmarshal(data, out)
//
//	// For big-endian:
//	Unmarshal(data, out, WithByteOrder(BigEndian))
func WithByteOrder(order ByteOrder) Option {
	return func(o *options) error {
		o.byteOrder = order
		return nil
	}
}

func collectOptions(opts []Option) (options, error) {
	var options options
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return options, err
		}
	}
	return options, nil
}
