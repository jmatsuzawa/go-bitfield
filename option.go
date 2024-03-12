package bitfield

type ByteOrder int

const (
	LittleEndian ByteOrder = iota
	BigEndian
)

type options struct {
	byteOrder ByteOrder
}

type Option func(*options) error

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
