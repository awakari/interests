package mongo

import (
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func Test_encodeCondition(t *testing.T) {
	cases := map[string]struct {
		src   model.Condition
		dst   Condition
		kiwis []kiwiSearchData
	}{
		"single Kiwi condition": {
			src: model.NewKiwiTreeCondition(
				model.NewKiwiCondition(
					model.NewKeyCondition(
						model.NewCondition(true),
						"cond0",
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
			src: model.NewGroupCondition(
				model.NewCondition(false),
				model.GroupLogicOr,
				[]model.Condition{
					model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(true),
								"cond1",
								"key0",
							),
							true,
							"pattern0",
						),
					),
					model.NewKiwiTreeCondition(
						model.NewKiwiCondition(
							model.NewKeyCondition(
								model.NewCondition(false),
								"cond2",
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
				Logic: model.GroupLogicOr,
			},
			kiwis: []kiwiSearchData{
				{
					Key:     "key0",
					Pattern: "pattern0",
					Partial: true,
				},
				{
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
			assert.Equal(t, c.dst, dst)
			assert.Equal(t, len(c.kiwis), len(kiwis))
			assert.ElementsMatch(t, c.kiwis, kiwis)
		})
	}
}

func Test_decodeRawCondition(t *testing.T) {
	cases := map[string]struct {
		raw bson.M
		out Condition
		err error
	}{
		"fail on non condition": {
			raw: bson.M{
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
					"not": false,
				},
				"logic": int32(model.GroupLogicAnd),
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
						"logic": int32(model.GroupLogicXor),
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
		dst model.Condition
		src Condition
	}{
		"single Kiwi condition": {
			dst: model.NewKiwiCondition(
				model.NewKeyCondition(
					model.NewCondition(true),
					"cond0",
					"key0",
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
			dst: model.NewGroupCondition(
				model.NewCondition(false),
				model.GroupLogicOr,
				[]model.Condition{
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(true),
							"cond1",
							"key0",
						),
						true,
						"pattern0",
					),
					model.NewKiwiCondition(
						model.NewKeyCondition(
							model.NewCondition(false),
							"cond2",
							"key1",
						),
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
				Logic: model.GroupLogicOr,
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
