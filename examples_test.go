package jschema_test

import (
	"fmt"
	"time"

	"github.com/ysmood/jschema"
	"github.com/ysmood/jschema/lib/test"
	"github.com/ysmood/vary"
)

func ExampleNew() {
	type Node struct {
		// Use tags like default, examples to set the json schema validation rules.
		ID int `json:"id" default:"1" examples:"[1,2,3]" min:"0" max:"100"`

		// Use the tags to set description, min, max, etc. All available tags are [jschema.JTag].
		// Use [jschema.JTagItemPrefix] to prefix [jschema.JTag] to set the array item.
		Children []*Node `json:"children" description:"The children of the node" minItems:"0" maxItems:"10"`
	}

	// Create a schema list instance
	schemas := jschema.New("#/components/schemas")

	// Define a type within the schema
	schemas.Define(Node{})
	schemas.Description(Node{}, "A node in the tree")

	fmt.Println(schemas.String())

	// Output:
	// {
	//   "Node": {
	//     "title": "Node",
	//     "description": "A node in the tree",
	//     "type": "object",
	//     "properties": {
	//       "children": {
	//         "description": "The children of the node",
	//         "type": "array",
	//         "items": {
	//           "anyOf": [
	//             {
	//               "$ref": "#/components/schemas/Node"
	//             },
	//             {
	//               "type": "null"
	//             }
	//           ]
	//         },
	//         "minItems": 0,
	//         "maxItems": 10
	//       },
	//       "id": {
	//         "default": 1,
	//         "examples": [
	//           1,
	//           2,
	//           3
	//         ],
	//         "type": "integer",
	//         "maximum": 100,
	//         "minimum": 0
	//       }
	//     },
	//     "required": [
	//       "id",
	//       "children"
	//     ],
	//     "additionalProperties": false
	//   }
	// }
}

func ExampleSchemas() {
	// Create a schema list instance
	schemas := jschema.New("#/components/schemas")

	type Metadata interface{}

	// Make the metadata field accept either A or B
	iMetadata := vary.New(new(Metadata))

	type A string

	iMetadata.Add(A(""))

	type B int

	iMetadata.Add(B(0))

	type Node struct {
		// Use the pattern or format tag to set the standard json schema validation rule
		Name     string   `json:"name" pattern:"^[a-z]+$" format:"name"`
		Metadata Metadata `json:"metadata,omitempty"` // omitempty make this field optional
		Version  string   `json:"version"`
		// jschema supports github.com/dmarkham/enumer generated enum
		// The enum type must implements [jschema.Enum] or [jschema.EnumString].
		// Otherwise, it will be treated as a normal type.
		Enum test.Enum `json:"enum"`
	}

	schemas.Define(Node{})
	node := schemas.PeakSchema(Node{})

	// Define default value
	{
		node.Properties["name"].Default = "jack"
	}

	// Define constants
	{
		node.Properties["version"] = schemas.Const("v1")
	}

	fmt.Println(schemas.String())

	// Output:
	// {
	//   "A": {
	//     "title": "A",
	//     "description": "github.com/ysmood/jschema_test.A",
	//     "type": "string"
	//   },
	//   "B": {
	//     "title": "B",
	//     "description": "github.com/ysmood/jschema_test.B",
	//     "type": "integer"
	//   },
	//   "Enum": {
	//     "title": "Enum",
	//     "description": "github.com/ysmood/jschema/lib/test.Enum",
	//     "enum": [
	//       "one",
	//       "three",
	//       "two"
	//     ]
	//   },
	//   "Metadata": {
	//     "title": "Metadata",
	//     "description": "github.com/ysmood/jschema_test.Metadata",
	//     "anyOf": [
	//       {
	//         "$ref": "#/components/schemas/A"
	//       },
	//       {
	//         "$ref": "#/components/schemas/B"
	//       }
	//     ]
	//   },
	//   "Node": {
	//     "title": "Node",
	//     "description": "github.com/ysmood/jschema_test.Node",
	//     "type": "object",
	//     "properties": {
	//       "enum": {
	//         "$ref": "#/components/schemas/Enum"
	//       },
	//       "metadata": {
	//         "$ref": "#/components/schemas/Metadata"
	//       },
	//       "name": {
	//         "default": "jack",
	//         "type": "string",
	//         "format": "name",
	//         "pattern": "^[a-z]+$"
	//       },
	//       "version": {
	//         "type": "string",
	//         "enum": [
	//           "v1"
	//         ]
	//       }
	//     },
	//     "required": [
	//       "name",
	//       "version",
	//       "enum"
	//     ],
	//     "additionalProperties": false
	//   }
	// }
}

func Example_custom_handler() {
	s := jschema.New("")

	s.Hijack(time.Time{}, func(scm *jschema.Schema) {
		// If we don't this the time will be a struct
		scm.Type = jschema.TypeString
		scm.AdditionalProperties = nil
	})

	type Data struct {
		Time time.Time `json:"time"`
	}

	s.Define(Data{})

	fmt.Println(s.String())

	// Output:
	// {
	//   "Data": {
	//     "title": "Data",
	//     "description": "github.com/ysmood/jschema_test.Data",
	//     "type": "object",
	//     "properties": {
	//       "time": {
	//         "$ref": "#/$defs/Time"
	//       }
	//     },
	//     "required": [
	//       "time"
	//     ],
	//     "additionalProperties": false
	//   },
	//   "Time": {
	//     "title": "Time",
	//     "description": "time.Time",
	//     "type": "string"
	//   }
	// }
}
