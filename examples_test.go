package jschema_test

import (
	"encoding/json"
	"fmt"

	"github.com/NaturalSelectionLabs/jschema"
)

func ExampleNew() {
	type Node struct {
		ID       int     `json:"id"`
		Children []*Node `json:"children"`
	}

	// Create a schema list instance
	schemas := jschema.New("#/components/schemas")

	// Define a type within the schema
	schemas.Define(Node{})

	// Marshal the schema list to json string
	out, err := json.MarshalIndent(schemas.JSON(), "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

	// Output:
	// {
	//   "Node": {
	//     "type": "object",
	//     "title": "Node",
	//     "description": "github.com/NaturalSelectionLabs/jschema_test.Node",
	//     "properties": {
	//       "children": {
	//         "type": "array",
	//         "items": {
	//           "nullable": true,
	//           "anyOf": [
	//             {
	//               "$ref": "#/components/schemas/Node"
	//             }
	//           ]
	//         }
	//       },
	//       "id": {
	//         "type": "number"
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

	type A string
	type B int

	type Node struct {
		Name     int         `json:"name"`
		Metadata interface{} `json:"metadata,omitempty"` // omitempty make this field optional
		Version  string      `json:"version"`
		Options  []string    `json:"options"`
	}

	schemas.Define(Node{})
	node := schemas.GetSchema(Node{})

	// Make the metadata field accept either A or B
	{
		node.Properties["metadata"] = schemas.AnyOf(A(""), B(0))
	}

	// Define constants
	{
		node.Properties["version"] = schemas.Const("v1")
	}

	// Define enum
	{
		node.Properties["options"].Enum = jschema.ToJValList(1, 2, 3)
	}

	// Marshal the schema list to json string
	out, err := json.MarshalIndent(schemas.JSON(), "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

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
	//   "Node": {
	//     "type": "object",
	//     "title": "Node",
	//     "description": "github.com/NaturalSelectionLabs/jschema_test.Node",
	//     "properties": {
	//       "metadata": {
	//         "anyOf": [
	//           {
	//             "$ref": "#/components/schemas/A"
	//           },
	//           {
	//             "$ref": "#/components/schemas/B"
	//           }
	//         ]
	//       },
	//       "name": {
	//         "type": "number"
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
