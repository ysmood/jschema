package jschema

import (
	"bytes"
	"encoding/gob"
	"reflect"

	"github.com/huandu/go-clone"
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

func (s *Schema) ChangeDefs(to string) {
	if s == nil {
		return
	}

	if s.Ref != nil {
		s.Ref.Defs = to
	}

	for _, ss := range s.AnyOf {
		ss.ChangeDefs(to)
	}

	for _, p := range s.Properties {
		p.ChangeDefs(to)
	}

	for _, p := range s.PatternProperties {
		p.ChangeDefs(to)
	}

	s.Items.ChangeDefs(to)

	for _, p := range s.Defs {
		p.ChangeDefs(to)
	}
}

func (s *Schemas) AnyOf(list ...interface{}) *Schema {
	ss := []*Schema{}

	for _, v := range list {
		ss = append(ss, s.Define(v))
	}

	return &Schema{
		AnyOf: ss,
	}
}

func (s *Schemas) Const(v JVal) *Schema {
	ss := s.Define(v)
	ss.Enum = []JVal{v}
	return ss
}

func (s *Schemas) ToStandAlone(scm *Schema) *Schema {
	scm = clone.Clone(scm).(*Schema)      //nolint: forcetypeassert
	types := clone.Clone(s.types).(Types) //nolint: forcetypeassert

	scm.Defs = types

	scm.ChangeDefs("#/$defs")

	return scm
}

// SchemaT returns a standalone schema for the given type.
func (s *Schemas) SchemaT(t reflect.Type) *Schema {
	return s.ToStandAlone(s.DefineT(t))
}

func ToJValList[T any](list ...T) []JVal {
	to := []JVal{}

	for _, v := range list {
		to = append(to, v)
	}

	return to
}

func indirectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// Check if x or x's pointer implements y.
// x can be a direct value or a pointer to the a direct value.
func implements(x, y reflect.Type) bool {
	return x.Implements(y) || reflect.New(x).Type().Implements(y)
}

// Add names to the list if they are not in the list.
func (r *Required) Add(names ...string) {
	for _, name := range names {
		if !r.Has(name) {
			*r = append(*r, name)
		}
	}
}

func (r *Required) Has(name string) bool {
	for _, n := range *r {
		if n == name {
			return true
		}
	}
	return false
}
