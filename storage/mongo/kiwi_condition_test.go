package mongo

import (
	"github.com/awakari/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func Test_decodeKiwiCondition(t *testing.T) {
	cases := map[string]struct {
		base ConditionBase
		raw  bson.M
		out  kiwiCondition
		err  error
	}{
		"ok": {
			base: ConditionBase{
				Not: true,
			},
			raw: bson.M{
				kiwiConditionAttrPartial: true,
				kiwiConditionAttrKey:     "key0",
				kiwiConditionAttrPattern: "pattern0",
			},
			out: kiwiCondition{
				Base: ConditionBase{
					Not: true,
				},
				Key:     "key0",
				Pattern: "pattern0",
				Partial: true,
			},
		},
		"fails to decode \"partial\" attribute": {
			base: ConditionBase{},
			raw: bson.M{
				kiwiConditionAttrPartial: 1,
				kiwiConditionAttrKey:     "key0",
				kiwiConditionAttrPattern: "pattern0",
			},
			err: storage.ErrInternal,
		},
		"fails due to missing \"kiwi\" attribute": {
			base: ConditionBase{},
			raw: bson.M{
				kiwiConditionAttrPartial: false,
			},
			err: storage.ErrInternal,
		},
		"fails due to nil": {
			base: ConditionBase{
				Not: true,
			},
			raw: bson.M{
				kiwiConditionAttrPartial: nil,
				kiwiConditionAttrKey:     "key0",
				kiwiConditionAttrPattern: "pattern0",
			},
			err: storage.ErrInternal,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := decodeKiwiCondition(c.base, c.raw)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.out, out)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
