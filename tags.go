package jschema

import "strings"

type JTag string

const (
	JTagDescription JTag = "description"
	JTagFormat      JTag = "format"
	JTagDefault     JTag = "default"
	JTagExamples    JTag = "examples"
	JTagPattern     JTag = "pattern"
	JTagMin         JTag = "min"
	JTagMax         JTag = "max"
)

const JTagItemPrefix = "item-"

func (t JTag) String() string {
	return string(t)
}

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// parseTag splits a struct field's json tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	tag, opt, _ := strings.Cut(tag, ",")
	return tag, tagOptions(opt)
}

// Contains reports whether a comma-separated list of options
// contains a particular option.
func (o tagOptions) Contains(option string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var name string
		name, s, _ = strings.Cut(s, ",")
		if name == option {
			return true
		}
	}
	return false
}
