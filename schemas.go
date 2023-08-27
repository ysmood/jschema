// Package jschema ...
package jschema

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Schemas struct {
	refPrefix string
	types     Types
	handlers  map[Ref]Handler
	names     map[string]map[string]int
}

type Types map[string]*Schema

// New Schemas instance. The defs is the prefix used for each $ref path.
// Such as if you set defs to "#/components/schemas",
// then a $ref may looks like "#/components/schemas/Node".
func New(refPrefix string) Schemas {
	if refPrefix == "" {
		refPrefix = "#/$defs"
	}

	return Schemas{
		refPrefix: refPrefix,
		types:     Types{},
		handlers:  map[Ref]Handler{},
		names:     map[string]map[string]int{},
	}
}

// Schema is designed for typescript conversion.
// Its fields is a strict subset of json schema fields.
type Schema struct {
	// string, number, boolean, null, array, object
	Type SchemaType `json:"type,omitempty"`

	Title string `json:"title,omitempty"`

	Description string `json:"description,omitempty"`

	Ref *Ref `json:"$ref,omitempty"`

	AnyOf []*Schema `json:"anyOf,omitempty"`
	Enum  []JVal    `json:"enum,omitempty"`

	Properties           Properties `json:"properties,omitempty"`
	PatternProperties    Properties `json:"patternProperties,omitempty"`
	Required             []string   `json:"required,omitempty"`
	AdditionalProperties *bool      `json:"additionalProperties,omitempty"`

	Items    *Schema `json:"items,omitempty"`
	MinItems *int    `json:"minItems,omitempty"`
	MaxItems *int    `json:"maxItems,omitempty"`
	Default  JVal    `json:"default,omitempty"`
}

type SchemaType string

const (
	TypeString  SchemaType = "string"
	TypeNumber  SchemaType = "number"
	TypeObject  SchemaType = "object"
	TypeArray   SchemaType = "array"
	TypeBool    SchemaType = "boolean"
	TypeNull    SchemaType = "null"
	TypeUnknown SchemaType = "unknown"
)

type Properties map[string]*Schema

// JVal can be any valid json value, e.g. string, number, bool, null, []interface{}, map[string]interface{}.
type JVal interface{}

func (s Schemas) has(r Ref) bool {
	_, has := s.types[r.ID]
	return has
}

func (s Schemas) add(r Ref, scm *Schema) {
	if r.Unique() {
		s.types[r.ID] = scm
	}
}

// DefineT converts the t to Schema recursively and append newly meet schemas to the schema list s.
func (s Schemas) DefineT(t reflect.Type) *Schema { //nolint: cyclop
	r := s.RefT(t)
	if s.has(r) {
		return &Schema{Ref: &r}
	}

	scm := &Schema{}
	s.add(r, scm)

	if r.Package != "" {
		scm.Title = r.Name
		scm.Description = fmt.Sprintf("%s.%s", r.Package, r.Name)
	}

	if h := s.getHandler(r); h != nil {
		*scm = *h()
		return scm
	}

	if t.Implements(tEnumValues) {
		scm.Type = TypeString
		scm.Enum = ToJValList(reflect.New(t).Interface().(EnumValues).Values()...) //nolint: forcetypeassert
		return &Schema{Ref: &r}
	}

	//nolint: exhaustive
	switch t.Kind() {
	case reflect.Interface:
		scm.Type = TypeObject

	case reflect.Bool:
		scm.Type = TypeBool

	case reflect.String:
		scm.Type = TypeString

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Uintptr, reflect.Complex64, reflect.Complex128:
		scm.Type = TypeNumber

	case reflect.Slice:
		el := s.DefineT(t.Elem())
		scm.Type = TypeArray
		scm.Items = el

	case reflect.Array:
		el := s.DefineT(t.Elem())
		l := t.Len()
		*scm = Schema{
			Type:     TypeArray,
			Items:    el,
			MinItems: &l,
			MaxItems: &l,
		}

	case reflect.Map:
		*scm = Schema{
			Type: TypeObject,
			PatternProperties: map[string]*Schema{
				"": s.DefineT(t.Elem()),
			},
		}

	case reflect.Struct:
		scm.Type = TypeObject
		scm.Properties = Properties{}
		scm.AdditionalProperties = new(bool)
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			if !f.IsExported() {
				continue
			}

			tag := ParseJSONTag(f.Tag)

			if tag != nil && tag.Ignore {
				continue
			}

			n := f.Name

			p := s.DefineT(f.Type)

			desc := f.Tag.Get("description")
			if desc != "" {
				p.Description = desc
			}

			defaultVal := f.Tag.Get("default")
			if defaultVal != "" {
				var d JVal
				err := json.Unmarshal([]byte(defaultVal), &d)
				if err != nil {
					panic(fmt.Errorf("default value is invalid json string for %s: %w", n, err))
				}
				p.Default = d
			}

			if tag != nil {
				if tag.Name != "" {
					n = tag.Name
				}
				if tag.String {
					p.Type = TypeString
				}
			}
			if tag == nil || !tag.Omitempty {
				scm.Required = append(scm.Required, n)
			}

			scm.Properties[n] = p
		}

	case reflect.Ptr:
		*scm = *s.DefineT(t.Elem())

		if scm.Ref == nil {
			n := *scm
			scm.Type = ""
			scm.AnyOf = []*Schema{&n, {Type: TypeNull}}
		} else {
			scm.AnyOf = []*Schema{{Ref: scm.Ref}, {Type: TypeNull}}
			scm.Ref = nil
		}

	default:
		scm.Type = TypeUnknown
	}

	if r.Unique() {
		scm = &Schema{
			Ref: &r,
		}
	}

	return scm
}

type EnumValues interface {
	Values() []string
}

var tEnumValues = reflect.TypeOf((*EnumValues)(nil)).Elem()
