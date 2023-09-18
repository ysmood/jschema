// Package jschema ...
package jschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/naturalselectionlabs/vary"
)

type Schemas struct {
	refPrefix  string
	types      Types
	handlers   map[Ref]Handler
	names      map[string]map[string]int
	interfaces vary.Interfaces
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
		refPrefix:  refPrefix,
		types:      Types{},
		handlers:   map[Ref]Handler{},
		names:      map[string]map[string]int{},
		interfaces: vary.Default,
	}
}

func NewWithInterfaces(refPrefix string, interfaces vary.Interfaces) Schemas {
	s := New(refPrefix)
	s.interfaces = interfaces
	return s
}

// Schema is designed for typescript conversion.
// Its fields is a strict subset of json schema fields.
type Schema struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Default     JVal   `json:"default,omitempty"`
	Example     JVal   `json:"example,omitempty"`

	// Any type validation
	AnyOf             []*Schema  `json:"anyOf,omitempty"`
	Ref               *Ref       `json:"$ref,omitempty"`
	Type              SchemaType `json:"type,omitempty"` // string, number, boolean, null, array, object
	Enum              []JVal     `json:"enum,omitempty"`
	Properties        Properties `json:"properties,omitempty"`
	PatternProperties Properties `json:"patternProperties,omitempty"`
	Format            string     `json:"format,omitempty"`

	// Number validation
	Max *float64 `json:"maximum,omitempty"`
	Min *float64 `json:"minimum,omitempty"`

	// String validation
	MaxLen  *float64 `json:"maxLength,omitempty"`
	MinLen  *float64 `json:"minLength,omitempty"`
	Pattern string   `json:"pattern,omitempty"`

	// Array validation
	Items    *Schema `json:"items,omitempty"`
	MinItems *int    `json:"minItems,omitempty"`
	MaxItems *int    `json:"maxItems,omitempty"`

	// Object validation
	Required             Required `json:"required,omitempty"`
	AdditionalProperties *bool    `json:"additionalProperties,omitempty"`

	Defs Types `json:"$defs,omitempty"`
}

type Required []string

type SchemaType string

const (
	TypeString  SchemaType = "string"
	TypeNumber  SchemaType = "number"
	TypeInteger SchemaType = "integer"
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

	if implements(t, tEnumString) {
		scm.Enum = ToJValList(reflect.New(t).Interface().(EnumString).Values()...) //nolint: forcetypeassert
		return &Schema{Ref: &r}
	}

	if implements(t, tEnum) {
		scm.Enum = ToJValList(reflect.New(t).Interface().(Enum).Values()...) //nolint: forcetypeassert
		return &Schema{Ref: &r}
	}

	if iter := s.interfaces[vary.ID(t)]; iter != nil {
		return s.defineInstances(iter)
	}

	//nolint: exhaustive
	switch t.Kind() {
	case reflect.Interface:
		scm.Type = ""

	case reflect.Bool:
		scm.Type = TypeBool

	case reflect.String:
		scm.Type = TypeString

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		scm.Type = TypeInteger

	case reflect.Float32, reflect.Float64,
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
		scm.AdditionalProperties = new(bool)
		for i := 0; i < t.NumField(); i++ {
			scm.mergeProps(s.DefineFieldT(t.Field(i)))
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

func (s Schemas) DefineFieldT(f reflect.StructField) *Schema { //nolint: cyclop
	scm := &Schema{
		Properties: Properties{},
	}

	if !f.IsExported() {
		return nil
	}

	tag := ParseJSONTag(f.Tag)

	if tag != nil && tag.Ignore {
		return nil
	}

	if f.Anonymous && (tag == nil || tag.Name == "") && indirectType(f.Type).Kind() == reflect.Struct {
		for i := 0; i < f.Type.NumField(); i++ {
			scm.mergeProps(s.DefineFieldT(f.Type.Field(i)))
		}
		return scm
	}

	p := s.DefineT(f.Type)

	err := p.loadTags(f)
	if err != nil {
		panic(fmt.Errorf("fail to load tag on field %s: %w", f.Name, err))
	}

	n := f.Name

	if tag != nil {
		if tag.Name != "" {
			n = tag.Name
		}
		if tag.String {
			p.Type = TypeString
		}
	}

	scm.Properties[n] = p

	if tag == nil || !tag.Omitempty {
		scm.Required.Add(n)
	}

	return scm
}

func (s Schemas) defineInstances(i *vary.Interface) *Schema {
	scm := s.DefineT(i.Self)
	is := s.PeakSchema(scm)
	is.Type = ""

	imps := []struct {
		key vary.TypeID
		typ reflect.Type
	}{}

	for id, p := range i.Implementations {
		imps = append(imps, struct {
			key vary.TypeID
			typ reflect.Type
		}{id, p})
	}

	sort.Slice(imps, func(i, j int) bool {
		return imps[i].key < imps[j].key
	})

	for _, p := range imps {
		ps := s.DefineT(p.typ)
		is.AnyOf = append(is.AnyOf, ps)
	}

	return scm
}

func jsonValTag(f reflect.StructField, tagName string) (JVal, error) { //nolint: ireturn
	tag := f.Tag.Get(tagName)
	if tag == "" {
		return nil, nil //nolint: nilnil
	}

	d := reflect.New(f.Type).Interface()

	err := json.Unmarshal([]byte(tag), d)
	if err == nil {
		return reflect.ValueOf(d).Elem().Interface(), nil
	}

	// Try to quote the string and parse it again
	b, _ := json.Marshal(tag) //nolint: errchkjson
	e := json.Unmarshal(b, &d)
	if e == nil {
		return reflect.ValueOf(d).Elem().Interface(), nil
	}

	return nil, fmt.Errorf("value of %s tag is invalid json string: %w", tagName, err)
}

func toNum(v string) *float64 {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return &f
}

func toInt(v string) *int {
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil
	}
	ii := int(i)
	return &ii
}

func (s *Schema) loadTags(f reflect.StructField) error {
	t := f.Tag
	s.Description = t.Get("description")
	s.Format = t.Get("format")

	val, err := jsonValTag(f, "default")
	if err != nil {
		return err
	}
	s.Default = val

	val, err = jsonValTag(f, "example")
	if err != nil {
		return err
	}
	s.Example = val

	if s.Type == TypeString {
		s.Pattern = t.Get("pattern")
		s.MinLen = toNum(t.Get("min"))
		s.MaxLen = toNum(t.Get("max"))
	}

	if s.Type == TypeNumber || s.Type == TypeInteger {
		s.Min = toNum(t.Get("min"))
		s.Max = toNum(t.Get("max"))
	}

	if s.Type == TypeArray {
		if s.MinItems == nil {
			s.MinItems = toInt(t.Get("min"))
		}
		if s.MaxItems == nil {
			s.MaxItems = toInt(t.Get("max"))
		}
	}

	return nil
}

func (s *Schema) mergeProps(target *Schema) {
	if s.Properties == nil {
		s.Properties = Properties{}
	}

	if target == nil {
		return
	}

	for k, v := range target.Properties {
		if _, has := s.Properties[k]; !has {
			s.Properties[k] = v
		}
	}
	s.Required.Add(target.Required...)
}
