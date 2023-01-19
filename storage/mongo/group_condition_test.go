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
								"key":     "k1",
								"pattern": "p1",
								"partial": true,
							},
							bson.M{
								"base": bson.M{
									"not": false,
								},
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
						Key:     "k0",
						Pattern: "p0",
						Base: ConditionBase{
							Not: true,
						},
						Partial: false,
					},
					groupCondition{
						Base: ConditionBase{
							Not: false,
						},
						Group: []Condition{
							kiwiCondition{
								Key:     "k1",
								Pattern: "p1",
								Base: ConditionBase{
									Not: false,
								},
								Partial: true,
							},
							kiwiCondition{
								Key:     "k2",
								Pattern: "p2",
								Base: ConditionBase{
									Not: false,
								},
								Partial: false,
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
