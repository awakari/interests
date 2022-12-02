package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubscription_Validate(t *testing.T) {
	cases := map[string]struct {
		sub Subscription
		err error
	}{
		"empty name": {
			sub: Subscription{
				Routes: []string{
					"destination",
				},
				Includes: MatcherGroup{
					Matchers: []Matcher{
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
			},
			err: ErrInvalidSubscription,
		},
		"empty matcher groups": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
			},
			err: ErrInvalidSubscription,
		},
		"empty routes": {
			sub: Subscription{
				Name: "sub0",
				Excludes: MatcherGroup{
					Matchers: []Matcher{
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
			},
			err: ErrInvalidSubscription,
		},
		"ok": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Excludes: MatcherGroup{
					Matchers: []Matcher{
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
			},
		},
		"dup matcher in includes group": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Includes: MatcherGroup{
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
			},
			err: ErrInvalidSubscription,
		},
		"dup matcher in excludes group": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Excludes: MatcherGroup{
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
			},
			err: ErrInvalidSubscription,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := c.sub.Validate()
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
