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
				Rule: NewMetadataPatternRule(NewMetadataRule(NewRule(true), "key0"), false, Pattern{}),
			},
			err: ErrInvalidSubscription,
		},
		"empty rule group": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Rule: NewGroupRule(NewRule(false), GroupAnd, []Rule{}),
			},
			err: ErrInvalidGroupRule,
		},
		"empty routes": {
			sub: Subscription{
				Name: "sub0",
				Rule: NewMetadataPatternRule(NewMetadataRule(NewRule(false), "key0"), false, Pattern{}),
			},
			err: ErrInvalidSubscription,
		},
		"ok": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Rule: NewMetadataPatternRule(NewMetadataRule(NewRule(false), "key0"), false, Pattern{}),
			},
		},
		"negation only rule": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Rule: NewMetadataPatternRule(NewMetadataRule(NewRule(true), "key0"), false, Pattern{}),
			},
			err: ErrInvalidSubscription,
		},
		"non pattern neither group root rule": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Rule: NewMetadataRule(NewRule(true), "key0"),
			},
			err: ErrInvalidSubscription,
		},
		"valid group root rule": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Rule: NewGroupRule(
					NewRule(false),
					GroupAnd,
					[]Rule{
						NewMetadataPatternRule(NewMetadataRule(NewRule(false), "key0"), false, Pattern{}),
						NewMetadataPatternRule(NewMetadataRule(NewRule(true), "key1"), false, Pattern{}),
					},
				),
			},
		},
		"invalid group root rule: negation": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Rule: NewGroupRule(
					NewRule(true),
					GroupAnd,
					[]Rule{
						NewMetadataPatternRule(NewMetadataRule(NewRule(false), "key0"), false, Pattern{}),
						NewMetadataPatternRule(NewMetadataRule(NewRule(true), "key1"), false, Pattern{}),
					},
				),
			},
			err: ErrInvalidSubscription,
		},
		"invalid group root rule: contains negation only child rules": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Rule: NewGroupRule(
					NewRule(false),
					GroupAnd,
					[]Rule{
						NewMetadataPatternRule(NewMetadataRule(NewRule(true), "key0"), false, Pattern{}),
						NewMetadataPatternRule(NewMetadataRule(NewRule(true), "key1"), false, Pattern{}),
					},
				),
			},
			err: ErrInvalidSubscription,
		},
		"invalid group root rule: contains more than 2 child rules": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Rule: NewGroupRule(
					NewRule(false),
					GroupAnd,
					[]Rule{
						NewMetadataPatternRule(NewMetadataRule(NewRule(true), "key0"), false, Pattern{}),
						NewMetadataPatternRule(NewMetadataRule(NewRule(false), "key1"), false, Pattern{}),
						NewMetadataPatternRule(NewMetadataRule(NewRule(false), "key2"), false, Pattern{}),
					},
				),
			},
			err: ErrInvalidSubscription,
		},
		"invalid group root rule: contains non pattern/group child rule": {
			sub: Subscription{
				Name: "sub0",
				Routes: []string{
					"destination",
				},
				Rule: NewGroupRule(
					NewRule(false),
					GroupAnd,
					[]Rule{
						NewMetadataPatternRule(NewMetadataRule(NewRule(true), "key0"), false, Pattern{}),
						NewGroupRule(),
					},
				),
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
