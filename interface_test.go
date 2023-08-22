package jschema_test

import (
	"math"
	"testing"

	"github.com/NaturalSelectionLabs/jschema"
	"github.com/ysmood/got"
)

var s = jschema.New("")

type Shape interface {
	Area() float64
}

var IShape = jschema.DefineI(s, new(Shape))

type Rectangle struct {
	Width  int
	Height int
}

var _ = IShape.Define(Rectangle{})

func (r Rectangle) Area() float64 {
	return float64(r.Width * r.Height)
}

type Circle struct {
	Radius float64
}

var _ = IShape.Define(Circle{})

func (r Circle) Area() float64 {
	return 2 * math.Pi * r.Radius
}

type Data struct {
	Shape Shape `json:"shape"`
}

func TestInterface(t *testing.T) {
	g := got.T(t)
	s.Define(Data{})

	g.Eq(g.JSON(g.ToJSONString(s.JSON())), map[string]interface{} /* len=4 */ {
		"Circle": map[string]interface{} /* len=6 */ {
			`additionalProperties` /* len=20 */ : false,
			"description":                        `github.com/NaturalSelectionLabs/jschema_test.Circle`, /* len=51 */
			"properties": map[string]interface{}{
				"Radius": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []interface{} /* len=1 cap=1 */ {
				"Radius",
			},
			"title": "Circle",
			"type":  "object",
		},
		"Data": map[string]interface{} /* len=6 */ {
			`additionalProperties` /* len=20 */ : false,
			"description":                        `github.com/NaturalSelectionLabs/jschema_test.Data`, /* len=49 */
			"properties": map[string]interface{}{
				"shape": map[string]interface{}{
					"$ref": "#/$defs/Shape",
				},
			},
			"required": []interface{} /* len=1 cap=1 */ {
				"shape",
			},
			"title": "Data",
			"type":  "object",
		},
		"Rectangle": map[string]interface{} /* len=6 */ {
			`additionalProperties` /* len=20 */ : false,
			"description":                        `github.com/NaturalSelectionLabs/jschema_test.Rectangle`, /* len=54 */
			"properties": map[string]interface{} /* len=2 */ {
				"Height": map[string]interface{}{
					"type": "number",
				},
				"Width": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []interface{} /* len=2 cap=2 */ {
				"Width",
				"Height",
			},
			"title": "Rectangle",
			"type":  "object",
		},
		"Shape": map[string]interface{} /* len=3 */ {
			"description": `github.com/NaturalSelectionLabs/jschema_test.Shape`, /* len=50 */
			"anyOf": []interface{} /* len=2 cap=2 */ {
				map[string]interface{}{
					"$ref": `#/$defs/Rectangle`, /* len=17 */
				},
				map[string]interface{}{
					"$ref": "#/$defs/Circle",
				},
			},
			"title": "Shape",
		},
	})
}
