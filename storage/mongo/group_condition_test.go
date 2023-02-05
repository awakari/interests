package mongo

import (
	"github.com/awakari/subscriptions/model"
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
						"id":      "cond0",
						"key":     "k0",
						"pattern": "p0",
						"partial": false,
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
								"id":      "cond2",
								"key":     "k1",
								"pattern": "p1",
								"partial": true,
							},
							bson.M{
								"base": bson.M{
									"not": false,
								},
								"id":      "cond3",
								"key":     "k2",
								"pattern": "p2",
								"partial": false,
							},
						},
						"logic": int32(model.GroupLogicXor),
					},
				},
				"logic": int32(model.GroupLogicAnd),
			},
			out: groupCondition{
				Base: ConditionBase{
					Not: false,
				},
				Group: []Condition{
					kiwiCondition{
						Id:      "cond0",
						Key:     "k0",
						Pattern: "p0",
						Partial: false,
						Base: ConditionBase{
							Not: true,
						},
					},
					groupCondition{
						Base: ConditionBase{
							Not: false,
						},
						Group: []Condition{
							kiwiCondition{
								Id:      "cond2",
								Key:     "k1",
								Pattern: "p1",
								Partial: true,
								Base: ConditionBase{
									Not: false,
								},
							},
							kiwiCondition{
								Id:      "cond3",
								Key:     "k2",
								Pattern: "p2",
								Partial: false,
								Base: ConditionBase{
									Not: false,
								},
							},
						},
						Logic: model.GroupLogicXor,
					},
				},
				Logic: model.GroupLogicAnd,
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
