package sdiffer

import (
	. "reflect"
	"regexp"
	"strconv"
	"strings"
)

type diffMode int

const (
	ignoreMode diffMode = iota
	includeMode
	allDiffMode
)

const (
	initTypeName      = "$"
	null              = "<nil>"
	notNull           = "<not nil>"
	defaultDepthLimit = 30
)

// Differ compares two interfaces with the same reflect.Type.
//
// For example:
// differ := NewDiffer().Ignore(`xxx`, `xxx`).Compare(a, b)
//
// Attention:
// Differ may cause panic when you call Compare.
type Differ struct {
	diffs       []*diff
	ignores     []*regexp.Regexp
	includes    []*regexp.Regexp
	trimSpaces  []*regexp.Regexp
	trimTags    []*trimTag
	comparators []Comparator
	maxDepth    int
	diffTmpl    string
	bff         *bufferF
}

func NewDiffer() *Differ {
	return &Differ{
		diffs:    make([]*diff, 0, 20),
		bff:      newBufferF(),
		maxDepth: defaultDepthLimit,
	}
}

func (d *Differ) String() string {
	for _, df := range d.diffs {
		d.bff.sprintf("%s\n", df.String(d.diffTmpl))
	}
	return d.bff.String()
}

func (d *Differ) Diffs() []*diff {
	return d.diffs
}

// WithMaxDepth set the max depth of Differ.
// Differ will panic if depth is over max depth when comparing.
func (d *Differ) WithMaxDepth(depth int) *Differ {
	d.maxDepth = depth
	return d
}

// WithTmpl set diff tmpl for Differ.
// Tmpl must contains exactly 3 placeholders, such as:
// `Field: "%s", A: %v, B: %v`
func (d *Differ) WithTmpl(tmpl string) *Differ {
	d.diffTmpl = tmpl
	return d
}

// Ignore set fields that do not need to be compared.
// Ignore will not work after Includes is called.
func (d *Differ) Ignore(regexps ...string) *Differ {
	if len(d.includes) > 0 {
		return d
	}
	d.ignores = make([]*regexp.Regexp, 0, len(regexps))
	for _, expr := range regexps {
		mustSuccess(func() error {
			if r, err := regexp.Compile(expr); err != nil {
				return err
			} else {
				d.ignores = append(d.ignores, r)
				return nil
			}
		})
	}
	return d
}

// Includes set fields that need to be compared.
// Ignore will not work after Includes is called.
func (d *Differ) Includes(regexps ...string) *Differ {
	d.includes = make([]*regexp.Regexp, 0, len(regexps))
	for _, expr := range regexps {
		mustSuccess(func() error {
			if r, err := regexp.Compile(expr); err != nil {
				return err
			} else {
				d.includes = append(d.includes, r)
				return nil
			}
		})
	}
	return d
}

// WithComparator specify some fields to use a customized Comparator.
func (d *Differ) WithComparator(c Comparator) *Differ {
	d.comparators = append(d.comparators, c)
	return d
}

// WithTrim trim string before comparison.
func (d *Differ) WithTrim(fieldPath string, cutset string) *Differ {
	d.trimTags = append(d.trimTags, newTrimTag(fieldPath, cutset))
	return d
}

func (d *Differ) WithTrimSpace(fieldPaths ...string) *Differ {
	for _, exp := range fieldPaths {
		d.trimSpaces = append(d.trimSpaces, regexp.MustCompile(exp))
	}
	return d
}

func (d *Differ) Reset() *Differ {
	d.includes = make([]*regexp.Regexp, 0, len(d.includes))
	d.ignores = make([]*regexp.Regexp, 0, len(d.ignores))
	d.trimSpaces = make([]*regexp.Regexp, 0, len(d.trimSpaces))
	d.trimTags = make([]*trimTag, 0, len(d.trimTags))
	d.comparators = make([]Comparator, 0, len(d.comparators))
	d.diffs = make([]*diff, 0)
	d.bff = newBufferF()
	return d
}

func (d *Differ) Compare(a, b interface{}) *Differ {
	va, vb := ValueOf(a), ValueOf(b)
	if va.Type() != vb.Type() {
		typeMismatchPanic(a, b)
	}
	tName := va.Type().Name()
	if va.Kind() == Ptr {
		tName = va.Elem().Type().Name()
	}
	d.doCompare(va, vb, iF(isStringBlank(tName), initTypeName, tName).(string), 0)
	return d
}

func (d *Differ) doCompare(a, b Value, fieldPath string, depth int) {
	if depth > d.maxDepth {
		panic("depth over limit")
	}

	if !a.IsValid() || !b.IsValid() {
		panic("value invalid: " + a.Type().String())
	}

	if a.Type() != b.Type() {
		typeMismatchPanic(a.Type(), b.Type())
	}

	for _, c := range d.comparators {
		if c.Match(fieldPath) {
			fieldPath = fieldPath + ".$[customized]"
			dt, va, vb := c.Equals(a.Interface(), b.Interface())
			switch dt {
			case LengthDiff:
				d.setLenDiff(fieldPath, a, b)
			case NilDiff:
				d.setNilDiff(fieldPath, a, b)
			case ElemDiff:
				d.setDiff(fieldPath, va, vb)
			case NoDiff:
				return
			default:
				panic("customized comparator returned an unexpected DiffType")
			}
			return
		}
	}

	switch a.Kind() {
	case Array:
		for i := 0; i < minInt(a.Len(), b.Len()); i++ {
			d.doCompare(a.Index(i), b.Index(i), a.Index(i).Type().Name(), depth)
		}
	case Slice:
		if a.IsNil() != b.IsNil() {
			d.setNilDiff(fieldPath, a, b)
			return
		}
		if a.Len() != b.Len() {
			d.setLenDiff(fieldPath, a, b)
		}
		if a.Pointer() == b.Pointer() {
			return
		}
		for i := 0; i < minInt(a.Len(), b.Len()); i++ {
			d.doCompare(a.Index(i), b.Index(i),
				concat(fieldPath, "[", strconv.Itoa(i), "]"), depth)
		}
	case Interface:
		if a.IsNil() != b.IsNil() {
			d.setNilDiff(fieldPath, a, b)
			return
		}
		d.doCompare(a, b, a.Type().Name(), depth+1)
	case Ptr:
		if a.IsNil() != b.IsNil() {
			d.setNilDiff(fieldPath, a, b)
			return
		}
		if a.Pointer() != b.Pointer() {
			d.doCompare(a.Elem(), b.Elem(), fieldPath, depth)
		}
	case Struct:
		for i, n := 0, a.NumField(); i < n; i++ {
			d.doCompare(a.Field(i), b.Field(i), concat(fieldPath, ".", a.Type().Field(i).Name), depth+1)
		}
	case Map:
		if a.IsNil() != b.IsNil() {
			d.setNilDiff(fieldPath, a, b)
			return
		}
		if a.Len() != b.Len() {
			d.setLenDiff(fieldPath, a, b)
		}
		for _, k := range a.MapKeys() {
			v1, v2 := a.MapIndex(k), b.MapIndex(k)
			d.doCompare(v1, v2, concat(fieldPath, "[", toString(k.Interface()), "]"), depth)
		}
	case String:
		for _, ts := range d.trimSpaces {
			if ts.MatchString(fieldPath) {
				if !DeepEqual(strings.TrimSpace(a.String()), strings.TrimSpace(b.String())) {
					d.setDiff(fieldPath, a, b)
				}
				return
			}
		}
		for _, tt := range d.trimTags {
			if tt.fieldRegexp.MatchString(fieldPath) {
				if !DeepEqual(tt.Trim(a.String()), tt.Trim(b.String())) {
					d.setDiff(fieldPath, a, b)
				}
				return
			}
		}
		fallthrough
	default:
		if !DeepEqual(a.Interface(), b.Interface()) {
			d.setDiff(fieldPath, a, b)
			return
		}
	}
}

func (d *Differ) setNilDiff(fieldName string, a, b Value) {
	d.setDiff(fieldName, iF(a.IsNil(), null, notNull), iF(b.IsNil(), null, notNull))
}

func (d *Differ) setLenDiff(fieldName string, a, b Value) {
	d.setDiff(fieldName+"[Length]", a.Len(), b.Len())
}

func (d *Differ) setDiff(fieldName string, va, vb interface{}) {
	switch d.getDiffMode() {
	case includeMode:
		if !d.isIncludedField(fieldName) {
			return
		}
	case ignoreMode:
		if d.isIgnoredField(fieldName) {
			return
		}
	}
	d.diffs = append(d.diffs, newDiff(fieldName, va, vb))
}

func (d *Differ) getDiffMode() diffMode {
	if len(d.includes) > 0 {
		return includeMode
	}
	if len(d.ignores) > 0 {
		return ignoreMode
	}
	return allDiffMode
}

func (d *Differ) isIncludedField(fieldName string) bool {
	for _, ic := range d.includes {
		if ic.MatchString(fieldName) {
			return true
		}
	}
	return false
}

func (d *Differ) isIgnoredField(fieldName string) bool {
	for _, ig := range d.ignores {
		if ig.MatchString(fieldName) {
			return true
		}
	}
	return false
}

func typeMismatchPanic(a, b interface{}) {
	panic("type mismatch: " + newDiff("type", a, b).String())
}
