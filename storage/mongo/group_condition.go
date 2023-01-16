package mongo

import (
	"github.com/awakari/subscriptions/model"
	"go.mongodb.org/mongo-driver/bson"
)

type groupCondition struct {
	Base  ConditionBase    `bson:"base"`
	Group []Condition      `bson:"group"`
	Logic model.GroupLogic `bson:"logic"`
}

const groupConditionAttrGroup = "group"
const groupConditionAttrLogic = "logic"

var _ Condition = (*groupCondition)(nil)

func decodeGroupCondition(baseCond ConditionBase, rawGroup bson.A, rawData bson.M) (gc groupCondition, err error) {
	gc.Base = baseCond
	gc.Logic = model.GroupLogic(rawData[groupConditionAttrLogic].(int32))
	for _, rawChild := range rawGroup {
		var child Condition
		child, err = decodeRawCondition(rawChild.(bson.M))
		gc.Group = append(gc.Group, child)
	}
	return
}
