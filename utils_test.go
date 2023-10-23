package jschema_test

import (
	"testing"

	"github.com/ysmood/got"
	"github.com/ysmood/jschema"
)

func TestChangeDefs(t *testing.T) {
	g := got.T(t)

	type A struct {
		ID int
	}

	type B struct {
		A A
	}

	s := jschema.New("")

	scm := s.Define(B{})

	old := g.JSON(s.String())

	g.Eq(g.JSON(g.ToJSON(s.ToStandAlone(scm))), map[string]interface{} /* len=2 */ {
		"$defs": map[string]interface{} /* len=2 */ {
			"A": map[string]interface{} /* len=6 */ {
				`additionalProperties` /* len=20 */ : false,
				"description":                        `github.com/ysmood/jschema_test.A`, /* len=46 */
				"properties": map[string]interface{}{
					"ID": map[string]interface{}{
						"type": "integer",
					},
				},
				"required": []interface{} /* len=1 cap=1 */ {
					"ID",
				},
				"title": "A",
				"type":  "object",
			},
			"B": map[string]interface{} /* len=6 */ {
				`additionalProperties` /* len=20 */ : false,
				"description":                        `github.com/ysmood/jschema_test.B`, /* len=46 */
				"properties": map[string]interface{}{
					"A": map[string]interface{}{
						"$ref": "#/$defs/A",
					},
				},
				"required": []interface{} /* len=1 cap=1 */ {
					"A",
				},
				"title": "B",
				"type":  "object",
			},
		},
		"$ref": "#/$defs/B",
	})

	g.Eq(old, g.JSON(s.String()))
}
