package mongo

import (
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/storage"
	"github.com/google/uuid"
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

func encodeCondition(src condition.Condition) (dst Condition, kiwis []kiwiSearchData) {
	id := uuid.NewString()
	switch c := src.(type) {
	case condition.GroupCondition:
		dst, kiwis = encodeGroupCondition(c, id)
	case condition.KiwiCondition:
		dst, kiwis = encodeKiwiCondition(c, id)
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

func decodeCondition(src Condition) (dst condition.Condition) {
	switch c := src.(type) {
	case groupCondition:
		var children []condition.Condition
		for _, childCond := range c.Group {
			children = append(children, decodeCondition(childCond))
		}
		dstBase := condition.NewCondition(c.Base.Id, c.Base.Not)
		dst = condition.NewGroupCondition(dstBase, condition.GroupLogic(c.Logic), children)
	case kiwiCondition:
		dstBase := condition.NewCondition(c.Base.Id, c.Base.Not)
		dstKey := condition.NewKeyCondition(dstBase, c.Key)
		dst = condition.NewKiwiCondition(dstKey, c.Partial, c.Pattern)
	}
	return dst
}
