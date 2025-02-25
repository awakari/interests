package mongo

import (
	"github.com/awakari/interests/model/condition"
	"github.com/awakari/interests/model/interest"
	"github.com/awakari/interests/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"
)

func Test_decodeInterest(t *testing.T) {
	cases := map[string]struct {
		in  interestRec
		out interest.Interest
		err error
	}{
		"ok": {
			in: interestRec{
				Id:             "sub0",
				GroupId:        "group0",
				UserId:         "acc0",
				Description:    "description0",
				Followers:      42,
				LimitPerMinute: 1,
				RawCondition: bson.M{
					conditionAttrBase: bson.M{
						conditionAttrNot: false,
					},
					numConditionAttrId:    "cond0",
					numConditionAttrKey:   "key0",
					textConditionAttrTerm: "pattern0",
				},
			},
			out: interest.Interest{
				Id:      "sub0",
				GroupId: "group0",
				UserId:  "acc0",
				Data: interest.Data{
					Description:    "description0",
					Followers:      42,
					LimitPerMinute: 1,
					Condition: condition.NewTextCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
						"pattern0", false,
					),
				},
			},
		},
		"ok w/ created, updated, read dates": {
			in: interestRec{
				Id:           "sub0",
				GroupId:      "group0",
				UserId:       "acc0",
				Description:  "description0",
				Enabled:      false,
				EnabledSince: time.Date(2025, 2, 1, 6, 53, 10, 0, time.UTC),
				Expires:      time.Date(2025, 2, 2, 6, 53, 10, 0, time.UTC),
				Created:      time.Date(2024, 4, 9, 6, 53, 10, 0, time.UTC),
				Updated:      time.Date(2024, 4, 9, 6, 53, 20, 0, time.UTC),
				Public:       true,
				RawCondition: bson.M{
					conditionAttrBase: bson.M{
						conditionAttrNot: false,
					},
					numConditionAttrId:    "cond0",
					numConditionAttrKey:   "key0",
					textConditionAttrTerm: "pattern0",
				},
			},
			out: interest.Interest{
				Id:      "sub0",
				GroupId: "group0",
				UserId:  "acc0",
				Data: interest.Data{
					Description:  "description0",
					Enabled:      false,
					EnabledSince: time.Date(2025, 2, 1, 6, 53, 10, 0, time.UTC),
					Expires:      time.Date(2025, 2, 2, 6, 53, 10, 0, time.UTC),
					Created:      time.Date(2024, 4, 9, 6, 53, 10, 0, time.UTC),
					Updated:      time.Date(2024, 4, 9, 6, 53, 20, 0, time.UTC),
					Public:       true,
					Condition: condition.NewTextCondition(
						condition.NewKeyCondition(condition.NewCondition(false), "cond0", "key0"),
						"pattern0", false,
					),
				},
			},
		},
		"fail": {
			in: interestRec{
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
			var out interest.Interest
			err := c.in.decodeInterest(&out)
			assert.ErrorIs(t, err, c.err)
			if c.err == nil {
				assert.Equal(t, c.out, out)
			}
		})
	}
}
