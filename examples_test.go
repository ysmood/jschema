package jschema_test

import (
	"fmt"
	"time"

	"github.com/NaturalSelectionLabs/jschema"
	"github.com/NaturalSelectionLabs/jschema/lib/test"
	"github.com/ysmood/vary"
)

func ExampleNew() {
	type Node struct {
		// The default tag only accepts json string.
		// So if you want to set a string value "jack",
		// you should use "\"jack\"" instead of "jack" for the field tag
		ID int `json:"id" default:"1" example:"2"`

		// Use the description tag to set the description of the field
		Children []*Node `json:"children" description:"The children of the node"`
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
	//     "type": "object",
	//     "title": "Node",
	//     "description": "A node in the tree",
	//     "properties": {
	//       "children": {
	//         "type": "array",
	//         "description": "The children of the node",
	//         "items": {
	//           "anyOf": [
	//             {
	//               "$ref": "#/components/schemas/Node"
	//             },
	//             {
	//               "type": "null"
	//             }
	//           ]
	//         }
	//       },
	//       "id": {
	//         "type": "integer",
	//         "default": 1,
	//         "example": 2
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
		Name     string    `json:"name"`
		Metadata Metadata  `json:"metadata,omitempty"` // omitempty make this field optional
		Version  string    `json:"version"`
		Enum     test.Enum `json:"enum"`
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
	//     "type": "string",
	//     "title": "A",
	//     "description": "github.com/NaturalSelectionLabs/jschema_test.A"
	//   },
	//   "B": {
	//     "type": "integer",
	//     "title": "B",
	//     "description": "github.com/NaturalSelectionLabs/jschema_test.B"
	//   },
	//   "Enum": {
	//     "type": "string",
	//     "title": "Enum",
	//     "description": "github.com/NaturalSelectionLabs/jschema/lib/test.Enum",
	//     "enum": [
	//       "one",
	//       "two",
	//       "three"
	//     ]
	//   },
	//   "Metadata": {
	//     "title": "Metadata",
	//     "description": "github.com/NaturalSelectionLabs/jschema_test.Metadata",
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
	//     "type": "object",
	//     "title": "Node",
	//     "description": "github.com/NaturalSelectionLabs/jschema_test.Node",
	//     "properties": {
	//       "enum": {
	//         "$ref": "#/components/schemas/Enum"
	//       },
	//       "metadata": {
	//         "$ref": "#/components/schemas/Metadata"
	//       },
	//       "name": {
	//         "type": "string",
	//         "default": "jack"
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

	s.AddHandler(time.Time{}, func() *jschema.Schema {
		return &jschema.Schema{
			Description: "time.Time",
			Title:       "Time",
			Type:        jschema.TypeNumber,
		}
	})

	type Data struct {
		Time time.Time `json:"time"`
	}

	s.Define(Data{})

	fmt.Println(s.String())

	// Output:
	// {
	//   "Data": {
	//     "type": "object",
	//     "title": "Data",
	//     "description": "github.com/NaturalSelectionLabs/jschema_test.Data",
	//     "properties": {
	//       "time": {
	//         "type": "number",
	//         "title": "Time",
	//         "description": "time.Time"
	//       }
	//     },
	//     "required": [
	//       "time"
	//     ],
	//     "additionalProperties": false
	//   },
	//   "Time": {
	//     "type": "number",
	//     "title": "Time",
	//     "description": "time.Time"
	//   }
	// }
}
