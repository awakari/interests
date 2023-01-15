package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
)

type kiwiCondition struct {
	Base         ConditionBase `bson:"base"`
	Key          string        `bson:"key"`
	Partial      bool          `bson:"partial"`
	ValuePattern pattern       `bson:"value_pattern"`
}

const kiwiConditionAttrKey = "key"
const kiwiConditionAttrPartial = "partial"
const kiwiConditionAttrValuePattern = "value_pattern"

var _ Condition = (*kiwiCondition)(nil)

func decodeKiwiCondition(baseCond ConditionBase, key any, raw bson.M) (kc kiwiCondition, err error) {
	kc.Base = baseCond
	kc.Key = key.(string)
	kc.Partial = raw[kiwiConditionAttrPartial].(bool)
	kc.ValuePattern, err = decodePattern(raw[kiwiConditionAttrValuePattern].(bson.M))
	return
}
