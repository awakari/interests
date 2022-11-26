package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatcherData_Equal(t *testing.T) {
	cases := map[string]struct {
		md1   MatcherData
		md2   MatcherData
		equal bool
	}{
		"": {
			equal: true,
		},
		"are equal": {
			md1: MatcherData{
				Key: "key0",
				Pattern: Pattern{
					Src: "pattern0",
				},
			},
			md2: MatcherData{
				Key: "key0",
				Pattern: Pattern{
					Src: "pattern0",
				},
			},
			equal: true,
		},
		"keys mismatch": {
			md1: MatcherData{
				Key: "key0",
				Pattern: Pattern{
					Src: "pattern0",
				},
			},
			md2: MatcherData{
				Key: "key1",
				Pattern: Pattern{
					Src: "pattern0",
				},
			},
			equal: false,
		},
		"patterns mismatch": {
			md1: MatcherData{
				Key: "key0",
				Pattern: Pattern{
					Src: "pattern0",
				},
			},
			md2: MatcherData{
				Key: "key0",
			},
			equal: false,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			equal := c.md1.Equal(c.md2)
			assert.Equal(t, c.equal, equal)
		})
	}
}
