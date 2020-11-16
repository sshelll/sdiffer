package sdiffer

import (
	"fmt"
	"strings"
)

const diffTmpl = `Field: "%s", A: %v, B: %v`

type diff struct {
	name string
	va   interface{}
	vb   interface{}
}

func newDiff(name string, a, b interface{}) *diff {
	return &diff{
		name: name,
		va:   a,
		vb:   b,
	}
}

// Tag generate a short tag of the diff name.
// for example:
// Person.Schools[0].Buildings[2].Name => Person.Schools.Buildings.Name
func (d *diff) Tag() (tag string) {
	cut := func(str string) string {
		idx := strings.Index(str, "[")
		if idx > 0 {
			return str[:idx]
		}
		return str
	}
	words := strings.Split(d.name, ".")
	for _, word := range words {
		if strings.HasSuffix(word, "]") {
			word = cut(word)
		}
		tag = iF(isStringBlank(tag), word, concat(tag, ".", word)).(string)
	}
	return
}

func (d *diff) String() string {
	return fmt.Sprintf(diffTmpl, d.name, d.va, d.vb)
}
