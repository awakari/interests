package mongo

import (
	"fmt"
	"github.com/awakari/interests/model/condition"
	"github.com/awakari/interests/storage"
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

func encodeCondition(src condition.Condition) (dst Condition, ids []string) {
	switch c := src.(type) {
	case condition.GroupCondition:
		dst, ids = encodeGroupCondition(c)
	case condition.TextCondition:
		dst, ids = encodeTextCondition(c)
	case condition.NumberCondition:
		dst, ids = encodeNumCondition(c)
	case condition.SemanticCondition:
		dst, ids = encodeSemCondition(c)

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
		term, isText := raw[textConditionAttrTerm].(string)
		num, isNum := raw[numConditionAttrVal].(float64)
		sem, isSem := raw[semConditionAttrQuery].(string)
		switch {
		case isGroup:
			result, err = decodeRawGroupCondition(baseCond, group, raw)
		case isText:
			result, err = decodeTextCondition(baseCond, term, raw)
		case isNum:
			result, err = decodeNumCondition(baseCond, num, raw)
		case isSem:
			result, err = decodeSemCondition(baseCond, sem, raw)
		default:
			err = fmt.Errorf("%w: undefined condition type: %v", storage.ErrInternal, raw)
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
	case textCondition:
		dstBase := condition.NewCondition(c.Base.Not)
		dstKey := condition.NewKeyCondition(dstBase, c.Id, c.Key)
		dst = condition.NewTextCondition(dstKey, c.Term, c.Exact)
	case numCondition:
		dstBase := condition.NewCondition(c.Base.Not)
		dstKey := condition.NewKeyCondition(dstBase, c.Id, c.Key)
		op := decodeNumOp(c.Op)
		dst = condition.NewNumberCondition(dstKey, op, c.Val)
	case semCondition:
		dstBase := condition.NewCondition(c.Base.Not)
		dst = condition.NewSemanticCondition(dstBase, c.Id, c.Query)
	}
	return dst
}
