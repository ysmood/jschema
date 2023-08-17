package jschema

import (
	"bytes"
	"encoding/gob"
	"reflect"
)

// Description set the description for current type.
func (s Schemas) Description(v interface{}, desc string) {
	scm := s.PeakSchema(v)
	scm.Description = desc
}

// Define is a shortcut for [Schemas.DefineT].
func (s Schemas) Define(v interface{}) *Schema {
	return s.DefineT(reflect.TypeOf(v))
}

// PeakSchema returns the schema for the given target it won't modify the schema list.
// If the target is a schema it will auto expand the ref and return the schema itself.
func (s *Schemas) PeakSchema(v interface{}) *Schema {
	r := s.Ref(v)

	if scm, ok := v.(*Schema); ok {
		if scm.Ref == nil {
			return scm
		} else {
			r = *scm.Ref
		}
	}

	return s.types[r.ID]
}

// SetSchema sets the schema for the given target. It will keep the title and description.
func (s *Schemas) SetSchema(target interface{}, v *Schema) {
	s.Define(target)
	ss := s.PeakSchema(target)
	title := ss.Title
	desc := ss.Description
	*ss = *v
	ss.Title = title
	ss.Description = desc
}

func (s *Schema) Clone() *Schema {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(s)
	if err != nil {
		panic(err)
	}

	n := &Schema{}
	err = gob.NewDecoder(buf).Decode(n)
	if err != nil {
		panic(err)
	}

	return n
}

func (s *Schemas) OneOf(list ...interface{}) *Schema {
	ss := []*Schema{}

	for _, v := range list {
		ss = append(ss, s.Define(v))
	}

	return &Schema{
		OneOf: ss,
	}
}

func (s *Schemas) Const(v JVal) *Schema {
	ss := s.Define(v)
	ss.Enum = []JVal{v}
	return ss
}

func ToJValList[T any](list ...T) []JVal {
	to := []JVal{}

	for _, v := range list {
		to = append(to, v)
	}

	return to
}
