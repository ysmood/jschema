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

func (s *Schema) ChangeDefs(to string) *Schema { //nolint: cyclop
	if s == nil {
		return s
	}

	n := *s

	if n.Ref != nil {
		n.Ref = &Ref{}
		*n.Ref = *s.Ref
		n.Ref.Defs = to
	}

	if n.AnyOf != nil {
		n.AnyOf = make([]*Schema, len(s.AnyOf))
		for i, ss := range s.AnyOf {
			n.AnyOf[i] = ss.ChangeDefs(to)
		}
	}

	if n.Enum != nil {
		n.Enum = make([]JVal, len(s.Enum))
		copy(n.Enum, s.Enum)
	}

	if n.Properties != nil {
		n.Properties = make(Properties, len(s.Properties))
		for k, p := range s.Properties {
			n.Properties[k] = p.ChangeDefs(to)
		}
	}

	if n.PatternProperties != nil {
		n.PatternProperties = make(Properties, len(s.Properties))
		for k, p := range s.PatternProperties {
			n.PatternProperties[k] = p.ChangeDefs(to)
		}
	}

	if n.Maximum != nil {
		n.Maximum = new(float64)
		*n.Maximum = *s.Maximum
	}
	if n.Minimum != nil {
		n.Minimum = new(float64)
		*n.Minimum = *s.Minimum
	}

	n.Items = s.Items.ChangeDefs(to)
	if n.Maximum != nil {
		n.MaxItems = new(int)
		*n.MaxItems = *s.MaxItems
	}
	if n.Minimum != nil {
		n.MinItems = new(int)
		*n.MinItems = *s.MinItems
	}

	if n.Required != nil {
		n.Required = make(Required, len(s.Required))
		copy(n.Required, s.Required)
	}

	if n.AdditionalProperties != nil {
		n.AdditionalProperties = new(bool)
		*n.AdditionalProperties = *s.AdditionalProperties
	}

	if n.Defs != nil {
		n.Defs = make(Types)
		for k, p := range s.Defs {
			n.Defs[k] = p.ChangeDefs(to)
		}
	}

	return &n
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
	if scm.Ref != nil {
		return &Schema{
			Ref:  scm.Ref,
			Defs: s.types,
		}
	}

	scm.Defs = s.types

	return scm.ChangeDefs("#/$defs")
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
