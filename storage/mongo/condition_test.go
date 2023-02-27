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
		src   condition.Condition
		dst   Condition
		kiwis []kiwiSearchData
	}{
		"single Kiwi condition": {
			src: condition.NewKiwiTreeCondition(
				condition.NewKiwiCondition(
					condition.NewKeyCondition(
						condition.NewCondition(true), "cond0",
						"key0",
					),
					true,
					"pattern0",
				),
			),
			dst: kiwiCondition{
				Id:      "cond0",
				Key:     "key0",
				Pattern: "pattern0",
				Partial: true,
				Base: ConditionBase{
					Not: true,
				},
			},
			kiwis: []kiwiSearchData{
				{
					Id:      "cond0",
					Key:     "key0",
					Pattern: "pattern0",
					Partial: true,
				},
			},
		},
		"group condition": {
			src: condition.NewGroupCondition(
				condition.NewCondition(false),
				condition.GroupLogicOr,
				[]condition.Condition{
					condition.NewKiwiTreeCondition(
						condition.NewKiwiCondition(
							condition.NewKeyCondition(
								condition.NewCondition(true), "cond1",
								"key0",
							),
							true,
							"pattern0",
						),
					),
					condition.NewKiwiTreeCondition(
						condition.NewKiwiCondition(
							condition.NewKeyCondition(
								condition.NewCondition(false), "cond2",
								"key1",
							),
							false,
							"pattern1",
						),
					),
				},
			),
			dst: groupCondition{
				Base: ConditionBase{
					Not: false,
				},
				Group: []Condition{
					kiwiCondition{
						Id:      "cond1",
						Key:     "key0",
						Pattern: "pattern0",
						Partial: true,
						Base: ConditionBase{
							Not: true,
						},
					},
					kiwiCondition{
						Id:      "cond2",
						Key:     "key1",
						Pattern: "pattern1",
						Partial: false,
						Base: ConditionBase{
							Not: false,
						},
					},
				},
				Logic: condition.GroupLogicOr,
			},
			kiwis: []kiwiSearchData{
				{
					Id:      "cond1",
					Key:     "key0",
					Pattern: "pattern0",
					Partial: true,
				},
				{
					Id:      "cond2",
					Key:     "key1",
					Pattern: "pattern1",
					Partial: false,
				},
			},
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dst, kiwis := encodeCondition(c.src)
			assert.True(t, conditionRecsEqual(c.dst, dst))
			assert.Equal(t, len(c.kiwis), len(kiwis))
			for i, k := range c.kiwis {
				assert.True(t, conditionRecsEqual(k, kiwis[i]))
			}
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
	case kiwiCondition:
		var bk kiwiCondition
		bk, equal = b.(kiwiCondition)
		if equal {
			equal = at.Base.Not == bk.Base.Not
		}
		if equal {
			equal = at.Partial == bk.Partial
		}
		if equal {
			equal = at.Key == bk.Key
		}
		if equal {
			equal = at.Pattern == bk.Pattern
		}
	case kiwiSearchData:
		var bk kiwiSearchData
		bk, equal = b.(kiwiSearchData)
		if equal {
			equal = at.Partial == bk.Partial
		}
		if equal {
			equal = at.Key == bk.Key
		}
		if equal {
			equal = at.Pattern == bk.Pattern
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
				"id":      "cond0",
				"key":     "k0",
				"partial": false,
				"pattern": "p0",
			},
			err: storage.ErrInternal,
		},
		"fail on unknown condition type": {
			raw: bson.M{
				"base": bson.M{
					"not": false,
				},
				"id":      "cond0",
				"partial": false,
				"pattern": "p0",
			},
			err: storage.ErrInternal,
		},
		"kiwi condition ok": {
			raw: bson.M{
				"base": bson.M{
					"not": false,
				},
				"id":      "cond0",
				"key":     "k0",
				"pattern": "p0",
				"partial": false,
			},
			out: kiwiCondition{
				Id:      "cond0",
				Key:     "k0",
				Pattern: "p0",
				Partial: false,
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
						"id":      "cond1",
						"key":     "k0",
						"pattern": "p0",
						"partial": false,
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
								"id":      "cond3",
								"key":     "k1",
								"pattern": "p1",
								"partial": true,
							},
							bson.M{
								"base": bson.M{
									"not": false,
								},
								"id":      "cond4",
								"key":     "k2",
								"pattern": "p2",
								"partial": false,
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
					kiwiCondition{
						Id:      "cond1",
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
								Id:      "cond3",
								Key:     "k1",
								Pattern: "p1",
								Partial: true,
								Base: ConditionBase{
									Not: false,
								},
							},
							kiwiCondition{
								Id:      "cond4",
								Key:     "k2",
								Pattern: "p2",
								Partial: false,
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
		"single Kiwi condition": {
			dst: condition.NewKiwiCondition(
				condition.NewKeyCondition(
					condition.NewCondition(true),
					"cond0", "key0",
				),
				true,
				"pattern0",
			),
			src: kiwiCondition{
				Id:      "cond0",
				Key:     "key0",
				Pattern: "pattern0",
				Partial: true,
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
					condition.NewKiwiCondition(
						condition.NewKeyCondition(condition.NewCondition(true), "cond1", "key0"),
						true,
						"pattern0",
					),
					condition.NewKiwiCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "cond2", "key1"),
						false,
						"pattern1",
					),
				},
			),
			src: groupCondition{
				Base: ConditionBase{
					Not: false,
				},
				Group: []Condition{
					kiwiCondition{
						Id:      "cond1",
						Key:     "key0",
						Pattern: "pattern0",
						Partial: true,
						Base: ConditionBase{
							Not: true,
						},
					},
					kiwiCondition{
						Id:      "cond2",
						Key:     "key1",
						Pattern: "pattern1",
						Partial: false,
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
