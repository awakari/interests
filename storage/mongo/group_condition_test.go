package mongo

import (
	"github.com/awakari/interests/model/condition"
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
			base: ConditionBase{},
			raw: bson.M{
				"group": bson.A{
					bson.M{
						"base": bson.M{
							"not": true,
						},
						"id":   "cond0",
						"key":  "k0",
						"term": "p0",
					},
					bson.M{
						"base": bson.M{
							"not": false,
						},
						"group": bson.A{
							bson.M{
								"base": bson.M{
									"not": false,
								},
								"id":   "cond2",
								"key":  "k1",
								"term": "p1",
							},
							bson.M{
								"base": bson.M{
									"not": false,
								},
								"id":    "cond3",
								"key":   "k2",
								"term":  "p2",
								"exact": true,
							},
						},
						"logic": int32(condition.GroupLogicXor),
					},
				},
				"logic": int32(condition.GroupLogicAnd),
			},
			out: groupCondition{
				Base: ConditionBase{
					Not: false,
				},
				Group: []Condition{
					textCondition{
						Id:   "cond0",
						Key:  "k0",
						Term: "p0",
						Base: ConditionBase{
							Not: true,
						},
					},
					groupCondition{
						Base: ConditionBase{
							Not: false,
						},
						Group: []Condition{
							textCondition{
								Id:   "cond2",
								Key:  "k1",
								Term: "p1",
								Base: ConditionBase{
									Not: false,
								},
							},
							textCondition{
								Id:    "cond3",
								Key:   "k2",
								Term:  "p2",
								Exact: true,
								Base: ConditionBase{
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
