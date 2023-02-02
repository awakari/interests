package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubscription_Validate(t *testing.T) {
	cases := map[string]struct {
		sub SubscriptionData
		err error
	}{
		"empty condition group": {
			sub: SubscriptionData{
				Routes: []string{
					"destination",
				},
				Condition: NewGroupCondition(NewCondition(false), GroupLogicAnd, []Condition{}),
			},
			err: ErrInvalidGroupCondition,
		},
		"empty routes": {
			sub: SubscriptionData{
				Condition: NewKiwiCondition(NewKeyCondition(NewCondition(false), "key0"), false, ""),
			},
			err: ErrInvalidSubscription,
		},
		"ok": {
			sub: SubscriptionData{
				Routes: []string{
					"destination",
				},
				Condition: NewKiwiCondition(NewKeyCondition(NewCondition(false), "key0"), false, ""),
			},
		},
		"negation only condition": {
			sub: SubscriptionData{
				Routes: []string{
					"destination",
				},
				Condition: NewKiwiCondition(NewKeyCondition(NewCondition(true), "key0"), false, ""),
			},
			err: ErrInvalidSubscription,
		},
		"non pattern neither group root condition": {
			sub: SubscriptionData{
				Routes: []string{
					"destination",
				},
				Condition: NewKeyCondition(NewCondition(true), "key0"),
			},
			err: ErrInvalidSubscription,
		},
		"valid group root condition": {
			sub: SubscriptionData{
				Routes: []string{
					"destination",
				},
				Condition: NewGroupCondition(
					NewCondition(false),
					GroupLogicAnd,
					[]Condition{
						NewKiwiCondition(NewKeyCondition(NewCondition(false), "key0"), false, ""),
						NewKiwiCondition(NewKeyCondition(NewCondition(true), "key1"), false, ""),
					},
				),
			},
		},
		"invalid group root condition: negation": {
			sub: SubscriptionData{
				Routes: []string{
					"destination",
				},
				Condition: NewGroupCondition(
					NewCondition(true),
					GroupLogicAnd,
					[]Condition{
						NewKiwiCondition(NewKeyCondition(NewCondition(false), "key0"), false, ""),
						NewKiwiCondition(NewKeyCondition(NewCondition(true), "key1"), false, ""),
					},
				),
			},
			err: ErrInvalidSubscription,
		},
		"invalid group root condition: contains negation only child rules": {
			sub: SubscriptionData{
				Routes: []string{
					"destination",
				},
				Condition: NewGroupCondition(
					NewCondition(false),
					GroupLogicAnd,
					[]Condition{
						NewKiwiCondition(NewKeyCondition(NewCondition(true), "key0"), false, ""),
						NewKiwiCondition(NewKeyCondition(NewCondition(true), "key1"), false, ""),
					},
				),
			},
			err: ErrInvalidSubscription,
		},
		"invalid group root condition: contains more than 2 child rules": {
			sub: SubscriptionData{
				Routes: []string{
					"destination",
				},
				Condition: NewGroupCondition(
					NewCondition(false),
					GroupLogicAnd,
					[]Condition{
						NewKiwiCondition(NewKeyCondition(NewCondition(true), "key0"), false, ""),
						NewKiwiCondition(NewKeyCondition(NewCondition(false), "key1"), false, ""),
						NewKiwiCondition(NewKeyCondition(NewCondition(false), "key2"), false, ""),
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
