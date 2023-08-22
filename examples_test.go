package jschema_test

import (
	"fmt"
	"time"

	"github.com/NaturalSelectionLabs/jschema"
)

func ExampleNew() {
	type Node struct {
		// The default tag only accepts json string.
		// So if you want to set a string value "jack",
		// you should use "\"jack\"" instead of "jack" for the field tag
		ID int `json:"id" default:"1"`

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
	//           "nullable": true,
	//           "oneOf": [
	//             {
	//               "$ref": "#/components/schemas/Node"
	//             }
	//           ]
	//         }
	//       },
	//       "id": {
	//         "type": "number",
	//         "default": 1
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
	IMetadata := jschema.DefineI(schemas, new(Metadata))

	type A string

	IMetadata.Define(A(""))

	type B int

	IMetadata.Define(B(0))

	type Node struct {
		Name     int      `json:"name"`
		Metadata Metadata `json:"metadata,omitempty"` // omitempty make this field optional
		Version  string   `json:"version"`
		Options  []string `json:"options"`
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

	// Define enum
	{
		node.Properties["options"].Enum = jschema.ToJValList(1, 2, 3)
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
	//     "type": "number",
	//     "title": "B",
	//     "description": "github.com/NaturalSelectionLabs/jschema_test.B"
	//   },
	//   "Metadata": {
	//     "title": "Metadata",
	//     "description": "github.com/NaturalSelectionLabs/jschema_test.Metadata",
	//     "oneOf": [
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
	//       "metadata": {
	//         "$ref": "#/components/schemas/Metadata"
	//       },
	//       "name": {
	//         "type": "number",
	//         "default": "jack"
	//       },
	//       "options": {
	//         "type": "array",
	//         "enum": [
	//           1,
	//           2,
	//           3
	//         ],
	//         "items": {
	//           "type": "string"
	//         }
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
	//       "options"
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
