package mongo

import (
	"fmt"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/storage"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type kiwiCondition struct {
	Base    ConditionBase `bson:"base"`
	Id      string        `bson:"id"`
	Partial bool          `bson:"partial"`
	Key     string        `bson:"key"`
	Pattern string        `bson:"pattern"`
}

const kiwiConditionAttrId = "id"
const kiwiConditionAttrPartial = "partial"
const kiwiConditionAttrKey = "key"
const kiwiConditionAttrPattern = "pattern"

var _ Condition = (*kiwiCondition)(nil)

func encodeKiwiCondition(src model.KiwiCondition) (dst kiwiCondition, kiwis []kiwiSearchData) {
	id := uuid.NewString()
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
			Not: src.IsNot(),
		},
		Id:      id,
		Partial: partial,
		Key:     key,
		Pattern: pattern,
	}
	return
}

func decodeKiwiCondition(baseCond ConditionBase, raw bson.M) (kc kiwiCondition, err error) {
	kc.Base = baseCond
	var ok bool
	kc.Id, ok = raw[kiwiConditionAttrId].(string)
	if ok {
		kc.Partial, ok = raw[kiwiConditionAttrPartial].(bool)
	}
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
