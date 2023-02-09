package subscription

import (
	"github.com/awakari/subscriptions/model/condition"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoute_Validate(t *testing.T) {
	cases := map[string]struct {
		route Route
		err   error
	}{
		"empty condition group": {
			route: Route{
				Destinations: []string{
					"destination",
				},
				Condition: condition.NewGroupCondition(condition.NewCondition("", false), condition.GroupLogicAnd, []condition.Condition{}),
			},
			err: condition.ErrInvalidGroupCondition,
		},
		"empty routes": {
			route: Route{
				Condition: condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", false), "key0"), false, ""),
			},
			err: ErrInvalidSubscriptionRoute,
		},
		"ok": {
			route: Route{
				Destinations: []string{
					"destination",
				},
				Condition: condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", false), "key0"), false, ""),
			},
		},
		"negation only condition": {
			route: Route{
				Destinations: []string{
					"destination",
				},
				Condition: condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", true), "key0"), false, ""),
			},
			err: ErrInvalidSubscriptionRoute,
		},
		"non pattern neither group root condition": {
			route: Route{
				Destinations: []string{
					"destination",
				},
				Condition: condition.NewKeyCondition(condition.NewCondition("", true), "key0"),
			},
			err: ErrInvalidSubscriptionRoute,
		},
		"valid group root condition": {
			route: Route{
				Destinations: []string{
					"destination",
				},
				Condition: condition.NewGroupCondition(
					condition.NewCondition("", false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", false), "key0"), false, ""),
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", true), "key1"), false, ""),
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", false), "key2"), false, ""),
					},
				),
			},
		},
		"invalid group root condition: negation": {
			route: Route{
				Destinations: []string{
					"destination",
				},
				Condition: condition.NewGroupCondition(
					condition.NewCondition("", true),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", false), "key0"), false, ""),
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", true), "key1"), false, ""),
					},
				),
			},
			err: ErrInvalidSubscriptionRoute,
		},
		"invalid group root condition: contains negation only child rules": {
			route: Route{
				Destinations: []string{
					"destination",
				},
				Condition: condition.NewGroupCondition(
					condition.NewCondition("", false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", true), "key0"), false, ""),
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", true), "key1"), false, ""),
					},
				),
			},
			err: condition.ErrInvalidGroupCondition,
		},
		"invalid group root condition: contains less than 2 child rules": {
			route: Route{
				Destinations: []string{
					"destination",
				},
				Condition: condition.NewGroupCondition(
					condition.NewCondition("", false),
					condition.GroupLogicAnd,
					[]condition.Condition{
						condition.NewKiwiCondition(condition.NewKeyCondition(condition.NewCondition("", true), "key0"), false, ""),
					},
				),
			},
			err: condition.ErrInvalidGroupCondition,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := c.route.Validate()
			if c.err == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
