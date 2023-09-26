package jschema

import (
	"encoding/json"
	"math/big"
	"reflect"
	"time"
)

// Hijack is the custom handler for a special type.
// scm is the parsed schema without this handler, modify it to produce the final schema.
type Hijack func(scm *Schema)

func (s Schemas) Hijack(v interface{}, h Hijack) {
	r := s.RefT(reflect.TypeOf(v))
	s.handlers[r] = h
}

func (s Schemas) getHijack(r Ref) Hijack {
	if h, has := s.handlers[r]; has {
		return h
	}
	return nil
}

func (s Schemas) HijackTime() {
	s.Hijack(time.Time{}, func(scm *Schema) {
		scm.Type = TypeString
		scm.AdditionalProperties = nil
	})
}

func (s Schemas) HijackBigInt() {
	s.Hijack(big.Int{}, func(scm *Schema) {
		scm.Type = TypeNumber
		scm.AdditionalProperties = nil
	})
}

func (s Schemas) HijackJSONRawMessage() {
	s.Hijack(json.RawMessage{}, func(scm *Schema) {
		scm.Type = ""
		scm.Items = nil
	})
}
