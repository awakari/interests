package mongo

import (
	"github.com/awakari/interests/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func Test_decodeTextCondition(t *testing.T) {
	cases := map[string]struct {
		base ConditionBase
		raw  bson.M
		out  textCondition
		err  error
	}{
		"ok": {
			base: ConditionBase{
				Not: true,
			},
			raw: bson.M{
				"id":                   "cond0",
				textConditionAttrKey:   "key0",
				textConditionAttrTerm:  "pattern0",
				textConditionAttrExact: true,
			},
			out: textCondition{
				Base: ConditionBase{
					Not: true,
				},
				Id:    "cond0",
				Key:   "key0",
				Term:  "pattern0",
				Exact: true,
			},
		},
		"fails to decode \"partial\" attribute": {
			base: ConditionBase{},
			raw: bson.M{
				textConditionAttrKey:  "key0",
				textConditionAttrTerm: "pattern0",
			},
			err: storage.ErrInternal,
		},
		"fails due to missing \"key\" attribute": {
			base: ConditionBase{},
			raw: bson.M{
				"term": "term0",
			},
			err: storage.ErrInternal,
		},
		"fails due to nil": {
			base: ConditionBase{
				Not: true,
			},
			raw: bson.M{
				textConditionAttrKey:  "key0",
				textConditionAttrTerm: "pattern0",
			},
			err: storage.ErrInternal,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := decodeTextCondition(c.base, c.raw["term"].(string), c.raw)
			if c.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, c.out, out)
			} else {
				assert.ErrorIs(t, err, c.err)
			}
		})
	}
}
