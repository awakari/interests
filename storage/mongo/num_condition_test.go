package mongo

import (
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func Test_decodeNumCondition(t *testing.T) {
	cases := map[string]struct {
		base ConditionBase
		raw  bson.M
		out  numCondition
		err  error
	}{
		"ok": {
			base: ConditionBase{
				Not: true,
			},
			raw: bson.M{
				"id":                "cond0",
				numConditionAttrKey: "key0",
				numConditionAttrOp:  int32(1),
				numConditionAttrVal: -3.1415926,
			},
			out: numCondition{
				Base: ConditionBase{
					Not: true,
				},
				Id:  "cond0",
				Key: "key0",
				Op:  int32(condition.NumOpGt),
				Val: -3.1415926,
			},
		},
		"fails to decode \"op\" attribute": {
			base: ConditionBase{},
			raw: bson.M{
				"id":                "cond0",
				numConditionAttrKey: "key0",
				numConditionAttrOp:  6,
				numConditionAttrVal: 1.2e-3,
			},
			err: storage.ErrInternal,
		},
		"fails due to missing \"key\" attribute": {
			base: ConditionBase{},
			raw: bson.M{
				"val": 0.0,
			},
			err: storage.ErrInternal,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := decodeNumCondition(c.base, c.raw["val"].(float64), c.raw)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.out, out)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
