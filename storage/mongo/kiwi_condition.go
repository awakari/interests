package mongo

import (
	"fmt"
	"github.com/awakari/subscriptions/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type kiwiCondition struct {
	Base    ConditionBase `bson:"base"`
	Partial bool          `bson:"partial"`
	Key     string        `bson:"key"`
	Pattern string        `bson:"pattern"`
}

const kiwiConditionAttrPartial = "partial"
const kiwiConditionAttrKey = "key"
const kiwiConditionAttrPattern = "pattern"

var _ Condition = (*kiwiCondition)(nil)

func decodeKiwiCondition(baseCond ConditionBase, raw bson.M) (kc kiwiCondition, err error) {
	kc.Base = baseCond
	var ok bool
	kc.Partial, ok = raw[kiwiConditionAttrPartial].(bool)
	if ok {
		kc.Key, ok = raw[kiwiConditionAttrKey].(string)
	}
	if ok {
		kc.Pattern, ok = raw[kiwiConditionAttrPattern].(string)
	}
	if !ok {
		err = fmt.Errorf("%w: failed to decode kiwi %v", storage.ErrInternal, raw)
	}
	return
}
