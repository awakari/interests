package mongo

import (
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func Test_encodeCondition(t *testing.T) {
	cases := map[string]struct {
		src     condition.Condition
		dst     Condition
		condIds []string
	}{
		"single text condition": {
			src: condition.NewTextCondition(
				condition.NewKeyCondition(
					condition.NewCondition(true), "cond0",
					"key0",
				),
				"pattern0",
				true,
			),
			dst: textCondition{
				Id:    "cond0",
				Key:   "key0",
				Term:  "pattern0",
				Exact: true,
				Base: ConditionBase{
					Not: true,
				},
			},
			condIds: []string{
				"cond0",
			},
		},
		"group condition": {
			src: condition.NewGroupCondition(
				condition.NewCondition(false),
				condition.GroupLogicOr,
				[]condition.Condition{
					condition.NewTextCondition(
						condition.NewKeyCondition(
							condition.NewCondition(true), "cond1",
							"key0",
						),
						"pattern0",
						false,
					),
					condition.NewTextCondition(
						condition.NewKeyCondition(
							condition.NewCondition(false), "cond2",
							"key1",
						),
						"pattern1",
						true,
					),
				},
			),
			dst: groupCondition{
				Base: ConditionBase{
					Not: false,
				},
				Group: []Condition{
					textCondition{
						Id:    "cond1",
						Key:   "key0",
						Term:  "pattern0",
						Exact: false,
						Base: ConditionBase{
							Not: true,
						},
					},
					textCondition{
						Id:    "cond2",
						Key:   "key1",
						Term:  "pattern1",
						Exact: true,
						Base: ConditionBase{
							Not: false,
						},
					},
				},
				Logic: condition.GroupLogicOr,
			},
			condIds: []string{
				"cond1",
				"cond2",
			},
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dst, ids := encodeCondition(c.src)
			assert.True(t, conditionRecsEqual(c.dst, dst))
			assert.Equal(t, len(c.condIds), len(ids))
			assert.ElementsMatch(t, c.condIds, ids)
		})
	}
}

func conditionRecsEqual(a, b Condition) (equal bool) {
	switch at := a.(type) {
	case groupCondition:
		var bg groupCondition
		bg, equal = b.(groupCondition)
		if equal {
			equal = at.Base.Not == bg.Base.Not
		}
		if equal {
			equal = at.Logic == bg.Logic
		}
		if equal {
			for i, child := range at.Group {
				equal = conditionRecsEqual(child, bg.Group[i])
				if !equal {
					break
				}
			}
		}
	case textCondition:
		var bk textCondition
		bk, equal = b.(textCondition)
		if equal {
			equal = at.Base.Not == bk.Base.Not
		}
		if equal {
			equal = at.Key == bk.Key
		}
		if equal {
			equal = at.Term == bk.Term
		}
	}
	return
}

func Test_decodeRawCondition(t *testing.T) {
	cases := map[string]struct {
		raw bson.M
		out Condition
		err error
	}{
		"fail on non condition": {
			raw: bson.M{
				"id":   "cond0",
				"key":  "k0",
				"term": "p0",
			},
			err: storage.ErrInternal,
		},
		"fail on unknown condition type": {
			raw: bson.M{
				"base": bson.M{
					"not": false,
				},
				"id":   "cond0",
				"term": "p0",
			},
			err: storage.ErrInternal,
		},
		"text condition ok": {
			raw: bson.M{
				"base": bson.M{
					"not": false,
				},
				"id":   "cond0",
				"key":  "k0",
				"term": "p0",
			},
			out: textCondition{
				Id:   "cond0",
				Key:  "k0",
				Term: "p0",
				Base: ConditionBase{
					Not: false,
				},
			},
		},
		"num condition ok": {
			raw: bson.M{
				"base": bson.M{
					"not": false,
				},
				"id":  "cond0",
				"key": "k0",
				"op":  int32(2),
				"val": 42.0,
			},
			out: numCondition{
				Id:  "cond0",
				Key: "k0",
				Op:  int32(2),
				Val: 42,
				Base: ConditionBase{
					Not: false,
				},
			},
		},
		"fail on invalid group condition": {
			raw: bson.M{
				"base": bson.M{
					"not": false,
				},
				"group": bson.M{
					"1": true,
				},
			},
			err: storage.ErrInternal,
		},
		"group condition ok": {
			raw: bson.M{
				"base": bson.M{
					"id":  "cond6",
					"not": false,
				},
				"logic": int32(condition.GroupLogicAnd),
				"group": bson.A{
					bson.M{
						"base": bson.M{
							"not": true,
						},
						"id":   "cond1",
						"key":  "k0",
						"term": "p0",
					},
					bson.M{
						"base": bson.M{
							"not": false,
						},
						"id":    "cond5",
						"logic": int32(condition.GroupLogicXor),
						"group": bson.A{
							bson.M{
								"base": bson.M{
									"not": false,
								},
								"id":   "cond3",
								"key":  "k1",
								"term": "p1",
							},
							bson.M{
								"base": bson.M{
									"not": false,
								},
								"id":  "cond4",
								"key": "k2",
								"op":  int32(3),
								"val": 4.5,
							},
						},
					},
				},
			},
			out: groupCondition{
				Base: ConditionBase{
					Not: false,
				},
				Group: []Condition{
					textCondition{
						Id:   "cond1",
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
								Id:   "cond3",
								Key:  "k1",
								Term: "p1",
								Base: ConditionBase{
									Not: false,
								},
							},
							numCondition{
								Id:  "cond4",
								Key: "k2",
								Op:  int32(3),
								Val: 4.5,
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
			out, err := decodeRawCondition(c.raw)
			assert.ErrorIs(t, err, c.err)
			if c.err == nil {
				assert.Equal(t, c.out, out)
			}
		})
	}
}

func Test_decodeCondition(t *testing.T) {
	cases := map[string]struct {
		dst condition.Condition
		src Condition
	}{
		"single text condition": {
			dst: condition.NewTextCondition(
				condition.NewKeyCondition(
					condition.NewCondition(true),
					"cond0", "key0",
				),
				"pattern0",
				false,
			),
			src: textCondition{
				Id:   "cond0",
				Key:  "key0",
				Term: "pattern0",
				Base: ConditionBase{
					Not: true,
				},
			},
		},
		"single num condition": {
			dst: condition.NewNumberCondition(
				condition.NewKeyCondition(
					condition.NewCondition(true),
					"cond0", "key0",
				),
				condition.NumOpLt,
				-1.2e-3,
			),
			src: numCondition{
				Id:  "cond0",
				Key: "key0",
				Op:  int32(5),
				Val: -1.2e-3,
				Base: ConditionBase{
					Not: true,
				},
			},
		},
		"invalid num condition": {
			dst: condition.NewNumberCondition(
				condition.NewKeyCondition(
					condition.NewCondition(true),
					"cond0", "key0",
				),
				condition.NumOpUndefined,
				-1.2e-3,
			),
			src: numCondition{
				Id:  "cond0",
				Key: "key0",
				Op:  int32(6),
				Val: -1.2e-3,
				Base: ConditionBase{
					Not: true,
				},
			},
		},
		"group condition": {
			dst: condition.NewGroupCondition(
				condition.NewCondition(false),
				condition.GroupLogicOr,
				[]condition.Condition{
					condition.NewNumberCondition(
						condition.NewKeyCondition(condition.NewCondition(true), "cond1", "key0"),
						condition.NumOpLte, 5,
					),
					condition.NewTextCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "cond2", "key1"),
						"pattern1", true,
					),
				},
			),
			src: groupCondition{
				Base: ConditionBase{
					Not: false,
				},
				Group: []Condition{
					numCondition{
						Id:  "cond1",
						Key: "key0",
						Op:  int32(4),
						Val: 5,
						Base: ConditionBase{
							Not: true,
						},
					},
					textCondition{
						Id:    "cond2",
						Key:   "key1",
						Term:  "pattern1",
						Exact: true,
						Base: ConditionBase{
							Not: false,
						},
					},
				},
				Logic: condition.GroupLogicOr,
			},
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dst := decodeCondition(c.src)
			assert.Equal(t, c.dst, dst)
		})
	}
}
