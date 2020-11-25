package sdiffer

import (
	"regexp"
	"strings"
)

type trimTag struct {
	fieldRegexp *regexp.Regexp
	cutset      string
}

func newTrimTag(exp, cutset string) *trimTag {
	return &trimTag{
		fieldRegexp: regexp.MustCompile(exp),
		cutset:      cutset,
	}
}

func (tt *trimTag) Trim(s string) string {
	return strings.Trim(s, tt.cutset)
}
