package subscription

import (
	"github.com/awakari/subscriptions/model/condition"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCondition_Validate(t *testing.T) {
	cases := map[string]struct {
		d   Data
		err error
	}{
		"empty condition group": {
			d: Data{
				Condition: condition.NewGroupCondition(condition.NewCondition(false), condition.GroupLogicAnd, []condition.Condition{}),
			},
			err: condition.ErrInvalidGroupCondition,
		},
		"ok": {
			d: Data{
				Condition: condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition(false), "", "key0"), false, ""),
			},
		},
		"negation only condition": {
			d: Data{
				Condition: condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition(true), "", "key0"), false, ""),
			},
			err: ErrInvalidSubscriptionCondition,
		},
		"non pattern neither group root condition": {
			d: Data{
				Condition: condition.NewKeyCondition(condition.NewCondition(true), "", "key0"),
			},
			err: ErrInvalidSubscriptionCondition,
		},
		"valid group root condition": {
			d: Data{
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition(false), "", "key0"), false, ""),
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition(true), "", "key1"), false, ""),
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition(false), "", "key2"), false, ""),
					},
				),
			},
		},
		"invalid group root condition: negation": {
			d: Data{
				Condition: condition.NewGroupCondition(
					condition.NewCondition(true),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition(false), "", "key0"), false, ""),
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition(true), "", "key1"), false, ""),
					},
				),
			},
			err: ErrInvalidSubscriptionCondition,
		},
		"invalid group root condition: contains negation only child rules": {
			d: Data{
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition(true), "", "key0"), false, ""),
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition(true), "", "key1"), false, ""),
					},
				),
			},
			err: condition.ErrInvalidGroupCondition,
		},
		"invalid group root condition: contains less than 2 child rules": {
			d: Data{
				Condition: condition.NewGroupCondition(
					condition.NewCondition(false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition(true), "", "key0"), false, ""),
					},
				),
			},
			err: condition.ErrInvalidGroupCondition,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := c.d.Validate()
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
