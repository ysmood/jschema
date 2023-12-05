package jschema

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
)

type Ref struct {
	Defs    string
	Package string
	Name    string
	Hash    string
	ID      string
}

func (s *Schemas) Ref(v interface{}) Ref {
	return s.RefT(reflect.TypeOf(v))
}

// RefI is a shortcut for [Schemas.Ref] to get the Ref from an interface pointer, such as:
//
//	var Animal interface{}
//	s.RefI(new(Animal))
//
// Because there's no other way to get the [reflect.Type] of an interface in golang.
func (s *Schemas) RefI(v interface{}) Ref {
	return s.RefT(reflect.TypeOf(v).Elem())
}

var regTrimGeneric = regexp.MustCompile(`\[.+\]$`)

func (s *Schemas) RefT(t reflect.Type) Ref {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(t.PkgPath()+t.Name())))

	id := regTrimGeneric.ReplaceAllString(t.Name(), "")

	list, ok := s.names[id]
	if !ok {
		list = map[string]int{}
		s.names[id] = list
	}

	i := 0
	if _, has := list[hash]; !has {
		i = len(list)
		list[hash] = i
	}
	if i != 0 {
		id = fmt.Sprintf("%s%d", id, i)
	}

	return Ref{s.refPrefix, t.PkgPath(), t.Name(), hash, id}
}

func (r Ref) String() string {
	return r.Package + "." + r.Name
}

func (r Ref) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%s/%s", r.Defs, r.ID))
}

func (r Ref) Unique() bool {
	return r.Package != "" && r.Name != ""
}
