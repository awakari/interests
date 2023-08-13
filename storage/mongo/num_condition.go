package mongo

import (
	"fmt"
	"github.com/awakari/subscriptions/model/condition"
	"github.com/awakari/subscriptions/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type numCondition struct {
	Base ConditionBase `bson:"base"`
	Id   string        `bson:"id"`
	Key  string        `bson:"key"`
	Op   int32         `bson:"op"`
	Val  float64       `bson:"val"`
}

const numConditionAttrId = "id"
const numConditionAttrKey = "key"
const numConditionAttrOp = "op"
const numConditionAttrVal = "val"

var _ Condition = (*numCondition)(nil)

func encodeNumCondition(src condition.NumberCondition) (dst numCondition, ids []string) {
	id := src.GetId()
	ids = append(ids, id)
	key := src.GetKey()
	op := src.GetOperation()
	val := src.GetValue()
	dst = numCondition{
		Base: ConditionBase{
			Not: src.IsNot(),
		},
		Id:  id,
		Key: key,
		Op:  int32(op),
		Val: val,
	}
	return
}

func decodeNumCondition(baseCond ConditionBase, val float64, raw bson.M) (nc numCondition, err error) {
	nc.Base = baseCond
	nc.Val = val
	var ok bool
	nc.Id, ok = raw[numConditionAttrId].(string)
	if ok {
		nc.Key, ok = raw[numConditionAttrKey].(string)
	}
	if ok {
		nc.Op, ok = raw[numConditionAttrOp].(int32)
	}
	if !ok {
		err = fmt.Errorf("%w: failed to decode the number condition %v", storage.ErrInternal, raw)
	}
	return
}

func decodeNumOp(src int32) (dst condition.NumOp) {
	switch src {
	case 1:
		dst = condition.NumOpGt
	case 2:
		dst = condition.NumOpGte
	case 3:
		dst = condition.NumOpEq
	case 4:
		dst = condition.NumOpLte
	case 5:
		dst = condition.NumOpLt
	default:
		dst = condition.NumOpUndefined
	}
	return
}
