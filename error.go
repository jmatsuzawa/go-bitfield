package bitfield

import (
	"reflect"
)

// TypeError describes an invalid type passed to [Unmarshal].
// (The argument to [Unmarshal] must be a non-nil pointer to a struct.)
type TypeError struct {
	Type    reflect.Type
	problem string
}

func (e *TypeError) Error() string {
	return "bitfield: " + e.problem
}

// FieldError describes an invalid bit-field in a struct passed to [Unmarshal].
type FieldError struct {
	Field   reflect.StructField
	problem string
}

func (e *FieldError) Error() string {
	return "bitfield: " + e.problem + " (" + e.Field.Name + " " + e.Field.Type.String() + " `" + string(e.Field.Tag) + "`)"
}
