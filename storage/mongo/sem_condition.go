package mongo

import (
	"fmt"
	"github.com/awakari/interests/model/condition"
	"github.com/awakari/interests/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type semCondition struct {
	Base  ConditionBase `bson:"base"`
	Id    string        `bson:"id"`
	Query string        `bson:"q"`
}

const semConditionAttrQuery = "q"

var _ Condition = (*semCondition)(nil)

func encodeSemCondition(src condition.SemanticCondition) (dst semCondition, ids []string) {
	id := src.GetId()
	ids = append(ids, id)
	q := src.Query()
	dst = semCondition{
		Base: ConditionBase{
			Not: src.IsNot(),
		},
		Id:    id,
		Query: q,
	}
	return
}

func decodeSemCondition(baseCond ConditionBase, q string, raw bson.M) (sc semCondition, err error) {
	sc.Base = baseCond
	sc.Query = q
	var ok bool
	sc.Id, ok = raw[numConditionAttrId].(string)
	if !ok {
		err = fmt.Errorf("%w: failed to decode the semantic condition %v", storage.ErrInternal, raw)
	}
	return
}
