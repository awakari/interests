package mongo

import (
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
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

func encodeKiwiCondition(src condition.KiwiCondition, id string) (dst kiwiCondition, kiwis []kiwiSearchData) {
	partial := src.IsPartial()
	key := src.GetKey()
	pattern := src.GetPattern()
	kd := kiwiSearchData{
		Id:      id,
		Partial: partial,
		Key:     key,
		Pattern: pattern,
	}
	kiwis = append(kiwis, kd)
	dst = kiwiCondition{
		Base: ConditionBase{
			Id:  id,
			Not: src.IsNot(),
		},
		Partial: partial,
		Key:     key,
		Pattern: pattern,
	}
	return
}

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
