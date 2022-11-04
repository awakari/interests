package model

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp/syntax"
	"testing"
)

func TestPattern_Matches(t *testing.T) {
	cases := []struct {
		input   string
		p       Pattern
		partial bool
		matches bool
		err     error
	}{
		{
			input: "",
			p: Pattern{
				Regex: "",
			},
			partial: true,
			matches: true,
		},
		{
			input: "foo",
			p: Pattern{
				Regex: "foo",
			},
			partial: false,
			matches: true,
		},
		{
			input: "foo",
			p: Pattern{
				Regex: "bar",
			},
			partial: false,
			matches: false,
		},
		{
			input: "foo",
			p: Pattern{
				Regex: "f..",
			},
			partial: false,
			matches: true,
		},
		{
			input: "foo bar",
			p: Pattern{
				Regex: "f..",
			},
			partial: false,
			matches: false,
		},
		{
			input: "foo bar yohoho",
			p: Pattern{
				Regex: "b..",
			},
			partial: true,
			matches: true,
		},
		{
			input: "foo bar",
			p: Pattern{
				Regex: "[",
			},
			partial: true,
			matches: false,
			err:     &syntax.Error{Code: syntax.ErrMissingBracket, Expr: `[`},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s, %s", c.input, c.p.Regex), func(t *testing.T) {
			matches, err := c.p.Matches(c.input, c.partial)
			assert.Equal(t, c.matches, matches)
			if c.err != nil {
				assert.Equal(t, c.err.Error(), err.Error())
			}
		})
	}
}
