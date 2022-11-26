package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatcher_Equal(t *testing.T) {
	cases := map[string]struct {
		m1    Matcher
		m2    Matcher
		equal bool
	}{
		"": {
			equal: true,
		},
		"are equal": {
			m1: Matcher{
				MatcherData: MatcherData{
					Key: "key0",
					Pattern: Pattern{
						Src: "pattern0",
					},
				},
				Partial: true,
			},
			m2: Matcher{
				MatcherData: MatcherData{
					Key: "key0",
					Pattern: Pattern{
						Src: "pattern0",
					},
				},
				Partial: true,
			},
			equal: true,
		},
		"partial field mismatch": {
			m1: Matcher{
				MatcherData: MatcherData{
					Key: "key0",
					Pattern: Pattern{
						Src: "pattern0",
					},
				},
				Partial: true,
			},
			m2: Matcher{
				MatcherData: MatcherData{
					Key: "key0",
					Pattern: Pattern{
						Src: "pattern0",
					},
				},
				Partial: false,
			},
			equal: false,
		},
		"data mismatch": {
			m1: Matcher{
				MatcherData: MatcherData{
					Key: "key0",
					Pattern: Pattern{
						Src: "pattern0",
					},
				},
				Partial: true,
			},
			m2: Matcher{
				MatcherData: MatcherData{
					Key: "key1",
					Pattern: Pattern{
						Src: "pattern0",
					},
				},
				Partial: true,
			},
			equal: false,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			equal := c.m1.Equal(c.m2)
			assert.Equal(t, c.equal, equal)
		})
	}
}
