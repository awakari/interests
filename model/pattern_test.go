package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPattern_Equal(t *testing.T) {
	cases := map[string]struct {
		p1    Pattern
		p2    Pattern
		equal bool
	}{
		"are equal": {
			p1: Pattern{
				Src: "foo",
			},
			p2: Pattern{
				Src: "foo",
			},
			equal: true,
		},
		"are not equal": {
			p1: Pattern{
				Src: "foo",
			},
			p2: Pattern{
				Src: "bar",
			},
			equal: false,
		},
		"codes are not compared": {
			p1: Pattern{
				Code: []byte{1, 2, 3},
			},
			p2: Pattern{
				Code: []byte{4, 5, 6},
			},
			equal: true,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			equal := c.p1.Equal(c.p2)
			assert.Equal(t, c.equal, equal)
		})
	}
}
