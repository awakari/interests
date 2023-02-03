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
	Id  string `bson:"id"`
	Not bool   `bson:"not"`
}

const conditionAttrBase = "base"
const conditionAttrId = "id"
const conditionAttrNot = "not"

var _ Condition = (*ConditionBase)(nil)

func encodeCondition(src model.Condition) (dst Condition, kiwis []kiwiSearchData) {
	bc := ConditionBase{
		Id:  src.GetId(),
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
		kc := kiwiCondition{
			Base:    bc,
			Partial: c.IsPartial(),
			Key:     c.GetKey(),
			Pattern: c.GetPattern(),
		}
		kd := kiwiSearchData{
			Partial: c.IsPartial(),
			Key:     c.GetKey(),
			Pattern: c.GetPattern(),
		}
		kiwis = append(kiwis, kd)
		dst = kc
	}
	return
}

func decodeRawCondition(raw bson.M) (result Condition, err error) {
	base, isBase := raw[conditionAttrBase]
	if !isBase {
		err = fmt.Errorf("%w: value is not a condition instance: %v", storage.ErrInternal, raw)
	} else {
		id := base.(bson.M)[conditionAttrId].(string)
		not := base.(bson.M)[conditionAttrNot].(bool)
		baseCond := ConditionBase{
			Id:  id,
			Not: not,
		}
		group, isGroup := raw[groupConditionAttrGroup].(bson.A)
		if isGroup {
			result, err = decodeRawGroupCondition(baseCond, group, raw)
		} else {
			result, err = decodeKiwiCondition(baseCond, raw)
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
		dstBase := model.NewConditionWithId(c.Base.Not, c.Base.Id)
		dst = model.NewGroupCondition(dstBase, model.GroupLogic(c.Logic), children)
	case kiwiCondition:
		dstBase := model.NewConditionWithId(c.Base.Not, c.Base.Id)
		dstKey := model.NewKeyCondition(dstBase, c.Key)
		dst = model.NewKiwiCondition(dstKey, c.Partial, c.Pattern)
	}
	return dst
}
