package jschema_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/xeipuuv/gojsonschema"
	"github.com/ysmood/got"
	"github.com/ysmood/jschema"
	"github.com/ysmood/jschema/lib/test"
	"github.com/ysmood/vary"
)

func TestTypeName(t *testing.T) {
	g := got.T(t)

	g.Eq(reflect.TypeOf(1).PkgPath(), "")
}

func TestNil(t *testing.T) {
	g := got.T(t)

	type A struct {
		A *A
	}

	c := jschema.New("")

	out := c.Define(A{})

	g.Eq(g.JSON(g.ToJSONString(out)), map[string]interface{}{
		"$ref": `#/$defs/A`, /* len=42 */
	})
}

func TestCommonSchema(t *testing.T) {
	g := got.T(t)

	c := jschema.New("")

	type Node2 struct {
		Map map[string]float64
		Any interface{}
	}

	type Node1 struct {
		Str     string `format:"email" pattern:"." minLen:"1" maxLen:"10"`
		Num     int    `json:"num,omitempty"`
		Bool    bool   `json:"bool"`
		Ignore  string `json:"-"`
		Slice   []Node1
		Arr     [2]float64 `item-min:"0"`
		Obj     *Node2
		Enum    test.Enum
		EnumPtr *test.Enum
		private int //nolint: unused
	}

	c.Define(Node1{})
	c.Define(Node2{})

	g.Eq(g.JSON(g.ToJSONString(c.Define(Node1{}))), map[string]interface{}{
		"$ref": "#/$defs/Node1",
	})

	g.Snapshot("common schema", c.JSON())
}

func TestHandler(t *testing.T) {
	g := got.T(t)

	c := jschema.New("")

	type A struct {
		Str string
	}
	type B struct {
		A A
	}

	c.Hijack(A{}, func(scm *jschema.Schema) {
		*scm = jschema.Schema{
			Description: "type A",
			Title:       "AA",
			Type:        "number",
		}
	})

	c.Define(B{})

	g.Snapshot("handler", c.JSON())
}

type Enum int

func (Enum) Values() []json.RawMessage {
	list := []json.RawMessage{}

	for _, v := range []Enum{One, Two} {
		b, _ := v.MarshalJSON()
		list = append(list, json.RawMessage(b))
	}

	return list
}

func (e Enum) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%d"`, e)), nil
}

func (e *Enum) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}

	*e = Enum(i)
	return nil
}

const (
	One Enum = 1
	Two Enum = 2
)

func TestEnum(t *testing.T) {
	g := got.T(t)

	c := jschema.New("")

	c.Define(Enum(0))

	g.Eq(g.JSON(c.String()), map[string]interface{}{
		"Enum": map[string]interface{} /* len=3 */ {
			"description": `github.com/ysmood/jschema_test.Enum`, /* len=49 */
			"enum": []interface{} /* len=2 cap=2 */ {
				"1",
				"2",
			},
			"title": "Enum",
		},
	})
}

func TestTime(t *testing.T) {
	g := got.T(t)

	c := jschema.New("")
	c.HijackTime()
	c.Define(time.Now())

	g.Eq(g.JSON(c.String()), map[string]interface{}{
		`Time` /* len=37 */ : map[string]interface{} /* len=3 */ {
			"description": "time.Time",
			"title":       "Time",
			"type":        "string",
		},
	})
}

func TestBigInt(t *testing.T) {
	g := got.T(t)

	c := jschema.New("")
	c.HijackBigInt()
	c.Define(big.Int{})

	g.Eq(g.JSON(c.String()), map[string]interface{}{
		`Int` /* len=36 */ : map[string]interface{} /* len=3 */ {
			"description": "math/big.Int",
			"title":       "Int",
			"type":        "number",
		},
	})
}

func TestNameConflict(t *testing.T) {
	g := got.T(t)

	c := jschema.New("")

	type Time struct {
		Name string
	}

	c.Define(time.Time{})
	c.Define(Time{})

	g.Snapshot("conflict", c.JSON())
}

func TestRawMessage(t *testing.T) {
	g := got.T(t)

	c := jschema.New("")
	c.HijackJSONRawMessage()

	type A struct {
		A json.RawMessage
	}

	c.Define(A{})

	g.Eq(g.JSON(c.String()), map[string]interface{}{
		"A": map[string]interface{}{
			"additionalProperties": false,
			"description":          "github.com/ysmood/jschema_test.A",
			"properties": map[string]interface{}{
				"A": map[string]interface{}{
					"$ref": "#/$defs/RawMessage",
				},
			},
			"required": []interface{}{
				"A",
			},
			"title": "A",
			"type":  "object",
		},
		"RawMessage": map[string]interface{}{
			"description": "encoding/json.RawMessage",
			"title":       "RawMessage",
		},
	})
}

func TestRef(t *testing.T) {
	g := got.T(t)

	c := jschema.New("")

	type A struct{}

	type C[T any] struct{}

	type B struct {
		A  A
		C  C[string]
		C2 C[int]
	}

	c.Define(B{})

	g.Eq(c.PeakSchema(A{}).Title, "A")

	g.Snapshot("ref", c.JSON())
}

func TestEmbeddedStruct(t *testing.T) {
	g := got.T(t)

	type A struct {
		Val float64
	}

	type B struct {
		A
	}

	c := jschema.New("")

	c.Define(B{})

	g.Eq(g.JSON(c.String()), map[string]interface{} /* len=2 */ {
		"B": map[string]interface{} /* len=6 */ {
			`additionalProperties` /* len=20 */ : false,
			"description":                        `github.com/ysmood/jschema_test.B`, /* len=46 */
			"properties": map[string]interface{}{
				"Val": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []interface{} /* len=1 cap=1 */ {
				"Val",
			},
			"title": "B",
			"type":  "object",
		},
	})
}

type Shape interface {
	Area() float64
}

var IShape = vary.New(new(Shape))

type Rectangle struct {
	Width  int
	Height int
}

var _ = IShape.Add(Rectangle{})

func (r Rectangle) Area() float64 {
	return float64(r.Width * r.Height)
}

type Circle struct {
	Radius float64
}

var _ = IShape.Add(Circle{})

func (r Circle) Area() float64 {
	return 2 * math.Pi * r.Radius
}

type Data struct {
	Shape Shape `json:"shape"`
}

func TestAnyOf(t *testing.T) {
	g := got.T(t)

	s := jschema.New("")

	s.Define(Data{})

	g.Snapshot("anyOf", s.JSON())

	js := s.JSON()

	schema := gojsonschema.NewGoLoader(map[string]interface{}{
		"$ref":  "#/$defs/Shape",
		"$defs": js,
	})

	{
		result, err := gojsonschema.Validate(
			schema,
			gojsonschema.NewGoLoader(map[string]interface{}{"Width": 1, "Height": 2}),
		)
		g.E(err)

		g.Desc("%v", result.Errors()).True(result.Valid())
	}

	{
		result, err := gojsonschema.Validate(
			schema,
			gojsonschema.NewGoLoader(map[string]interface{}{"Width": 1, "Height": 2, "Radius": 3}),
		)
		g.E(err)

		g.Eq(result.Errors()[1].String(), "(root): Additional property Radius is not allowed")
	}
}

func TestSchemaT(t *testing.T) {
	g := got.T(t)
	{
		s := jschema.New("")
		c := s.SchemaT(reflect.TypeOf([]int{}))

		g.Eq(g.JSON(c.String()), map[string]interface{} /* len=2 */ {
			"items": map[string]interface{}{
				"type": "integer",
			},
			"type": "array",
		})
	}

	{
		type A struct{}
		type B struct {
			A A
		}

		s := jschema.New("")
		c := s.SchemaT(reflect.TypeOf(B{}))

		g.Eq(g.JSON(c.String()), map[string]interface{} /* len=2 */ {
			"$defs": map[string]interface{} /* len=2 */ {
				"A": map[string]interface{} /* len=4 */ {
					`additionalProperties` /* len=20 */ : false,
					"description":                        `github.com/ysmood/jschema_test.A`, /* len=46 */
					"title":                              "A",
					"type":                               "object",
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
	}
}

func TestDefaultTag(t *testing.T) {
	type X struct {
		A int    `default:"1"`
		B *int   `default:"1"`
		C uint   `default:"1"`
		D *uint  `default:"1"`
		E []uint `default:"[1,2,3]"`
		F string `default:""`
	}

	s := jschema.New("")

	s.Define(X{})

	x := reflect.ValueOf(&X{}).Elem()
	for k, p := range s.JSON()["X"].Properties {
		x.FieldByName(k).Set(reflect.ValueOf(p.Default))
	}
}

func TestOverrideRef(t *testing.T) {
	g := got.T(t)

	type A struct {
		A int
	}

	type B struct {
		A A `description:"B" max:"10"`
	}

	s := jschema.New("")

	s.Define(B{})

	g.Eq(g.JSON(s.String()), map[string]interface{}{
		"A": map[string]interface{}{
			"additionalProperties": false,
			"description":          "github.com/ysmood/jschema_test.A",
			"properties": map[string]interface{}{
				"A": map[string]interface{}{
					"type": "integer",
				},
			},
			"required": []interface{}{
				"A",
			},
			"title": "A",
			"type":  "object",
		},
		"B": map[string]interface{}{
			"additionalProperties": false,
			"description":          "github.com/ysmood/jschema_test.B",
			"properties": map[string]interface{}{
				"A": map[string]interface{}{
					"description": "B",
					"maximum":     10.0,
					"$ref":        "#/$defs/A",
				},
			},
			"required": []interface{}{
				"A",
			},
			"title": "B",
			"type":  "object",
		},
	})
}

func TestRefInterface(t *testing.T) {
	g := got.T(t)

	s := jschema.New("")
	ref := s.RefI(new(Shape))

	g.Eq(ref, jschema.Ref{
		Defs:    "#/$defs",
		Package: "github.com/ysmood/jschema_test",
		Name:    "Shape",
		Hash:    "ae1cec3fbb3190bd993c3e5a681546a3",
		ID:      "Shape",
	})
}

func TestAnyOfInterface(t *testing.T) {
	g := got.T(t)

	type A struct{}

	type B struct{}

	type C interface{}

	_ = vary.New(new(C), A{}, B{})

	s := jschema.New("")

	s.Define(new(C))

	g.Snapshot("any", s.JSON())
}
