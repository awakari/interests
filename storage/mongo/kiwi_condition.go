package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
)

type kiwiCondition struct {
	kiwi
	Base    ConditionBase `bson:"base"`
	Partial bool          `bson:"partial"`
}

const kiwiConditionAttrKey = "key"
const kiwiConditionAttrPartial = "partial"
const kiwiConditionAttrPattern = "pattern"

var _ Condition = (*kiwiCondition)(nil)

func decodeKiwiCondition(baseCond ConditionBase, key any, raw bson.M) (kc kiwiCondition, err error) {
	kc.Base = baseCond
	kc.Key = key.(string)
	kc.Partial = raw[kiwiConditionAttrPartial].(bool)
	kc.Pattern = raw[kiwiConditionAttrPattern].(string)
	return
}
