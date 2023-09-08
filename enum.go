package jschema

import (
	"encoding/json"
	"reflect"
)

type Enum interface {
	json.Marshaler
	json.Unmarshaler
	// Values returns all the possible raw json strings of the enum.
	Values() []json.RawMessage
}

var tEnum = reflect.TypeOf((*Enum)(nil)).Elem()

type EnumString interface {
	json.Marshaler
	json.Unmarshaler
	Values() []string
}

var tEnumString = reflect.TypeOf((*EnumString)(nil)).Elem()
