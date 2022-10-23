package model

import "regexp"

type (

	// PatternCode is a pattern identifier. Generally, not equal to the source pattern string.
	PatternCode []byte

	Pattern struct {
		Code  PatternCode
		Regex string
		Src   string
	}
)

func (p Pattern) Matches(input string, partial bool) (matches bool, err error) {
	var r *regexp.Regexp
	if partial {
		r, err = regexp.Compile(p.Regex)
	} else {
		r, err = regexp.Compile("^" + p.Regex + "$")
	}
	if err == nil {
		matches = r.MatchString(input)
	}
	return
}
