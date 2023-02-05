package mongo

import (
	"github.com/awakari/subscriptions/model"
	"go.mongodb.org/mongo-driver/bson"
)

type groupCondition struct {
	Base  ConditionBase `bson:"base"`
	Group []Condition   `bson:"group"`
	Logic int32         `bson:"logic"`
}

const groupConditionAttrGroup = "group"
const groupConditionAttrLogic = "logic"

var _ Condition = (*groupCondition)(nil)

func encodeGroupCondition(src model.GroupCondition) (dst groupCondition, kiwis []kiwiSearchData) {
	var group []Condition
	for _, childSrc := range src.GetGroup() {
		childDst, childKiwis := encodeCondition(childSrc)
		group = append(group, childDst)
		kiwis = append(kiwis, childKiwis...)
	}
	dst = groupCondition{
		Base: ConditionBase{
			Not: src.IsNot(),
		},
		Group: group,
		Logic: int32(src.GetLogic()),
	}
	return
}

func decodeRawGroupCondition(baseCond ConditionBase, rawGroup bson.A, rawData bson.M) (gc groupCondition, err error) {
	gc.Base = baseCond
	gc.Logic = rawData[groupConditionAttrLogic].(int32)
	for _, rawChild := range rawGroup {
		var child Condition
		child, err = decodeRawCondition(rawChild.(bson.M))
		gc.Group = append(gc.Group, child)
	}
	return
}
