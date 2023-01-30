package mongo

import (
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func Test_decodeSubscription(t *testing.T) {
	cases := map[string]struct {
		in  subscription
		out model.Subscription
		err error
	}{
		"ok": {
			in: subscription{
				Name: "sub0",
				Routes: []string{
					"route0",
					"route1",
				},
				RawCondition: bson.M{
					conditionAttrBase: bson.M{
						conditionAttrNot: false,
					},
					kiwiConditionAttrPartial: true,
					kiwiConditionAttrKey:     "key0",
					kiwiConditionAttrPattern: "pattern0",
				},
			},
			out: model.Subscription{
				Name:        "sub0",
				Description: "",
				Routes: []string{
					"route0",
					"route1",
				},
				Condition: model.NewKiwiCondition(
					model.NewKeyCondition(
						model.NewCondition(false),
						"key0",
					),
					true,
					"pattern0",
				),
			},
		},
		"fail": {
			in: subscription{
				Name: "sub0",
				Routes: []string{
					"route0",
					"route1",
				},
				RawCondition: bson.M{},
			},
			err: storage.ErrInternal,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var out model.Subscription
			err := c.in.decodeSubscription(&out)
			assert.ErrorIs(t, err, c.err)
			if c.err == nil {
				assert.Equal(t, c.out, out)
			}
		})
	}
}
