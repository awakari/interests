package mongo

import (
	"fmt"
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

func decodeCondition(raw bson.M) (result Condition, err error) {
	base, isBase := raw[conditionAttrBase]
	fmt.Printf("%v", base)
	if !isBase {
		err = fmt.Errorf("%w: value is not a condition instance: %v", storage.ErrInternal, raw)
	} else {
		not := base.(bson.M)[conditionAttrNot].(bool)
		baseCond := ConditionBase{Not: not}
		group, isGroup := raw[groupConditionAttrGroup].(bson.A)
		if isGroup {
			result, err = decodeGroupCondition(baseCond, group, raw)
		} else {
			key, present := raw[kiwiConditionAttrKey]
			if !present {
				err = fmt.Errorf("%w: value doesn't contain neither \"group\" attribute nor \"key\": %v", storage.ErrInternal, raw)
			} else {
				result, err = decodeKiwiCondition(baseCond, key, raw)
			}
		}
	}
	return
}
