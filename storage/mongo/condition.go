package mongo

import (
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
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

func encodeCondition(src condition.Condition) (dst Condition, kiwis []kiwiSearchData) {
	switch c := src.(type) {
	case condition.GroupCondition:
		dst, kiwis = encodeGroupCondition(c)
	case condition.KiwiCondition:
		dst, kiwis = encodeKiwiCondition(c)
	}
	return
}

func decodeRawCondition(raw bson.M) (result Condition, err error) {
	base, isBase := raw[conditionAttrBase]
	if !isBase {
		err = fmt.Errorf("%w: value is not a condition instance: %v", storage.ErrInternal, raw)
	} else {
		not := base.(bson.M)[conditionAttrNot].(bool)
		baseCond := ConditionBase{
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

func decodeCondition(src Condition) (dst condition.Condition) {
	switch c := src.(type) {
	case groupCondition:
		var children []condition.Condition
		for _, childCond := range c.Group {
			children = append(children, decodeCondition(childCond))
		}
		dstBase := condition.NewCondition(c.Base.Not)
		dst = condition.NewGroupCondition(dstBase, condition.GroupLogic(c.Logic), children)
	case kiwiCondition:
		dstBase := condition.NewCondition(c.Base.Not)
		dstKey := condition.NewKeyCondition(dstBase, c.Id, c.Key)
		dst = condition.NewKiwiCondition(dstKey, c.Partial, c.Pattern)
	}
	return dst
}
