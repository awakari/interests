package mongo

import (
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/model/subscription"
	"github.com/awakari/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func Test_decodeSubscription(t *testing.T) {
	cases := map[string]struct {
		in  subscriptionRec
		out subscription.Subscription
		err error
	}{
		"ok": {
			in: subscriptionRec{
				Id: "sub0",
				Metadata: map[string]string{
					"description": "description0",
				},
				Destinations: []string{
					"route0",
					"route1",
				},
				RawCondition: bson.M{
					conditionAttrBase: bson.M{
						conditionAttrNot: false,
					},
					kiwiConditionAttrId:      "cond0",
					kiwiConditionAttrPartial: true,
					kiwiConditionAttrKey:     "key0",
					kiwiConditionAttrPattern: "pattern0",
				},
			},
			out: subscription.Subscription{
				Id: "sub0",
				Data: subscription.Data{
					Metadata: map[string]string{
						"description": "description0",
					},
					Route: subscription.Route{
						Destinations: []string{
							"route0",
							"route1",
						},
						Condition: condition.NewKiwiCondition(
							condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
							true,
							"pattern0",
						),
					},
				},
			},
		},
		"fail": {
			in: subscriptionRec{
				Id: "sub0",
				Metadata: map[string]string{
					"description": "description0",
				},
				Destinations: []string{
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
			var out subscription.Subscription
			err := c.in.decodeSubscription(&out)
			assert.ErrorIs(t, err, c.err)
			if c.err == nil {
				assert.Equal(t, c.out, out)
			}
		})
	}
}
