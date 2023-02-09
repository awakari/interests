package mongo

import (
	"github.com/awakari/subscriptions/model/condition"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func Test_decodeGroupCondition(t *testing.T) {
	cases := map[string]struct {
		base ConditionBase
		raw  bson.M
		out  groupCondition
		err  error
	}{
		"ok": {
			base: ConditionBase{
				Id: "cond1",
			},
			raw: bson.M{
				"group": bson.A{
					bson.M{
						"base": bson.M{
							"id":  "cond0",
							"not": true,
						},
						"key":     "k0",
						"pattern": "p0",
						"partial": false,
					},
					bson.M{
						"base": bson.M{
							"id":  "cond4",
							"not": false,
						},
						"group": bson.A{
							bson.M{
								"base": bson.M{
									"id":  "cond2",
									"not": false,
								},
								"key":     "k1",
								"pattern": "p1",
								"partial": true,
							},
							bson.M{
								"base": bson.M{
									"id":  "cond3",
									"not": false,
								},
								"key":     "k2",
								"pattern": "p2",
								"partial": false,
							},
						},
						"logic": int32(condition.GroupLogicXor),
					},
				},
				"logic": int32(condition.GroupLogicAnd),
			},
			out: groupCondition{
				Base: ConditionBase{
					Id:  "cond1",
					Not: false,
				},
				Group: []Condition{
					kiwiCondition{
						Key:     "k0",
						Pattern: "p0",
						Partial: false,
						Base: ConditionBase{
							Id:  "cond0",
							Not: true,
						},
					},
					groupCondition{
						Base: ConditionBase{
							Id:  "cond4",
							Not: false,
						},
						Group: []Condition{
							kiwiCondition{
								Key:     "k1",
								Pattern: "p1",
								Partial: true,
								Base: ConditionBase{
									Id:  "cond2",
									Not: false,
								},
							},
							kiwiCondition{
								Key:     "k2",
								Pattern: "p2",
								Partial: false,
								Base: ConditionBase{
									Id:  "cond3",
									Not: false,
								},
							},
						},
						Logic: condition.GroupLogicXor,
					},
				},
				Logic: condition.GroupLogicAnd,
			},
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := decodeRawGroupCondition(c.base, c.raw[groupConditionAttrGroup].(bson.A), c.raw)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.out, out)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
