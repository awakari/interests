package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatcherGroup_Validate(t *testing.T) {
	//
	cases := map[string]struct {
		mg  MatcherGroup
		err error
	}{
		"success": {
			mg: MatcherGroup{
				Matchers: []Matcher{
					{
						MatcherData: MatcherData{
							Key: "key0",
							Pattern: Pattern{
								Src: "pattern0",
							},
						},
					},
					{
						Partial: true,
						MatcherData: MatcherData{
							Key: "key0",
							Pattern: Pattern{
								Src: "pattern0",
							},
						},
					},
					{
						MatcherData: MatcherData{
							Key: "key1",
							Pattern: Pattern{
								Src: "pattern0",
							},
						},
					},
					{
						MatcherData: MatcherData{
							Key: "key0",
							Pattern: Pattern{
								Src: "pattern1",
							},
						},
					},
				},
			},
		},
		"fail": {
			mg: MatcherGroup{
				Matchers: []Matcher{
					{
						MatcherData: MatcherData{
							Key: "key0",
							Pattern: Pattern{
								Src: "pattern0",
							},
						},
					},
					{
						MatcherData: MatcherData{
							Key: "key0",
							Pattern: Pattern{
								Src: "pattern0",
							},
						},
					},
				},
			},
			err: ErrInvalidMatcherGroup,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := c.mg.Validate()
			assert.ErrorIs(t, err, c.err)
		})
	}
}
