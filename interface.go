package jschema

import (
	"reflect"
)

type Interface[T any] struct {
	s   Schemas
	ref Ref
}

func DefineI[T any](s Schemas, i *T) Interface[T] {
	typ := reflect.TypeOf(i).Elem()
	s.DefineT(typ)
	return Interface[T]{s, s.RefT(typ)}
}

// Define implementation.
func (i Interface[T]) Define(target T) *Schema {
	scm := i.s.types[i.ref.ID]
	scm.Type = ""
	s := i.s.Define(target)
	scm.AnyOf = append(scm.AnyOf, s)
	return s
}
