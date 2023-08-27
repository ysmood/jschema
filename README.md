# Overview

A lib to convert existing golang struct into json schema list.

Features:

- No need to modify the existing struct
- Support `anyOf` for interface typing
- Support custom type handler
- Support easy modification of the generated schema
- Support enum [](https://github.com/ent/ent/blob/a792f429a659bf74debdabea1b27856daeb47d22/schema/field/field.go#L920-L923) type

## Usage

Read [example](examples_test.go) for more details.
