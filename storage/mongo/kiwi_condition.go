package mongo

import (
	"fmt"
	"github.com/awakari/subscriptions/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type kiwiCondition struct {
	Base    ConditionBase `bson:"base"`
	Kiwi    Kiwi          `bson:"kiwi"`
	Partial bool          `bson:"partial"`
}

const kiwiConditionAttrKiwi = "kiwi"
const kiwiConditionAttrPartial = "partial"

var _ Condition = (*kiwiCondition)(nil)

func decodeKiwiCondition(baseCond ConditionBase, rawKiwi any, raw bson.M) (kc kiwiCondition, err error) {
	kc.Base = baseCond
	var ok bool
	kc.Partial, ok = raw[kiwiConditionAttrPartial].(bool)
	var rawKiwiRec bson.M
	if ok {
		rawKiwiRec, ok = rawKiwi.(bson.M)
	}
	if ok {
		kc.Kiwi, err = decodeRawKiwi(rawKiwiRec)
	} else {
		err = fmt.Errorf("%w: failed to decode the kiwi condition %v", storage.ErrInternal, raw)
	}
	return
}
