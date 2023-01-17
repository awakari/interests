package mongo

import (
	"fmt"
	"github.com/awakari/subscriptions/model"
	"github.com/awakari/subscriptions/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type Condition interface {
}

type ConditionBase struct {
	Not bool `bson:"not"`
}

const conditionAttrBase = "base"
const conditionAttrNot = "not"

var _ Condition = (*ConditionBase)(nil)

func encodeCondition(src model.Condition) (dst Condition, kiwis []Kiwi) {
	bc := ConditionBase{
		Not: src.IsNot(),
	}
	switch c := src.(type) {
	case model.GroupCondition:
		var group []Condition
		for _, childSrc := range c.GetGroup() {
			childDst, childKiwis := encodeCondition(childSrc)
			group = append(group, childDst)
			kiwis = append(kiwis, childKiwis...)
		}
		dst = groupCondition{
			Base:  bc,
			Group: group,
			Logic: int32(c.GetLogic()),
		}
	case model.KiwiCondition:
		k := Kiwi{
			Key:     c.GetKey(),
			Pattern: c.GetPattern(),
		}
		kc := kiwiCondition{
			Kiwi:    k,
			Base:    bc,
			Partial: c.IsPartial(),
		}
		kiwis = append(kiwis, k)
		dst = kc
	}
	return
}

func decodeRawCondition(raw bson.M) (result Condition, err error) {
	base, isBase := raw[conditionAttrBase]
	if !isBase {
		err = fmt.Errorf("%w: value is not a condition instance: %v", storage.ErrInternal, raw)
	} else {
		not := base.(bson.M)[conditionAttrNot].(bool)
		baseCond := ConditionBase{Not: not}
		group, isGroup := raw[groupConditionAttrGroup].(bson.A)
		if isGroup {
			result, err = decodeRawGroupCondition(baseCond, group, raw)
		} else {
			rawKiwi, present := raw[kiwiConditionAttrKiwi]
			if !present {
				err = fmt.Errorf("%w: value doesn't contain neither \"group\" nor \"kiwi\" attribute: %v", storage.ErrInternal, raw)
			} else {
				result, err = decodeKiwiCondition(baseCond, rawKiwi, raw)
			}
		}
	}
	return
}

func decodeCondition(src Condition) (dst model.Condition) {
	switch c := src.(type) {
	case groupCondition:
		var children []model.Condition
		for _, childCond := range c.Group {
			children = append(children, decodeCondition(childCond))
		}
		dstBase := model.NewCondition(c.Base.Not)
		dst = model.NewGroupCondition(dstBase, model.GroupLogic(c.Logic), children)
	case kiwiCondition:
		dstBase := model.NewCondition(c.Base.Not)
		dstKey := model.NewKeyCondition(dstBase, c.Kiwi.Key)
		dst = model.NewKiwiCondition(dstKey, c.Partial, c.Kiwi.Pattern)
	}
	return dst
}
