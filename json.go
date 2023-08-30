package jschema

import (
	"encoding/json"
	"reflect"
)

// JSON returns a JSON representation of the schemas.
// Usually you [json.Marshal] it to generate a json file.
func (s Schemas) JSON() map[string]*Schema {
	m := map[string]*Schema{}

	for id, v := range s.types {
		m[id] = v
	}

	return m
}

// String returns a json string representation of the schemas.
func (s Schemas) String() string {
	b, err := json.MarshalIndent(s.JSON(), "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

type Tag struct {
	Name      string
	Ignore    bool
	Omitempty bool
	String    bool
}

func ParseJSONTag(st reflect.StructTag) *Tag {
	v := st.Get("json")

	if v == "" {
		return nil
	}

	if v == "-" {
		return &Tag{
			Ignore: true,
		}
	}

	name, t := parseTag(v)

	return &Tag{
		Name:      name,
		Omitempty: t.Contains("omitempty"),
		String:    t.Contains("string"),
	}
}
