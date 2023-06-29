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
				Id:          "sub0",
				GroupId:     "group0",
				UserId:      "acc0",
				Description: "description0",
				RawCondition: bson.M{
					conditionAttrBase: bson.M{
						conditionAttrNot: false,
					},
					textConditionAttrId:   "cond0",
					textConditionAttrKey:  "key0",
					textConditionAttrTerm: "pattern0",
				},
			},
			out: subscription.Subscription{
				Id:      "sub0",
				GroupId: "group0",
				UserId:  "acc0",
				Data: subscription.Data{
					Description: "description0",
					Condition: condition.NewTextCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
						"pattern0",
					),
				},
			},
		},
		"fail": {
			in: subscriptionRec{
				Id:           "sub0",
				GroupId:      "group0",
				UserId:       "acc0",
				Description:  "description0",
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
