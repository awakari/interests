package mongo

import (
	"fmt"
	"github.com/awakari/interests/model/condition"
	"github.com/awakari/interests/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type textCondition struct {
	Base  ConditionBase `bson:"base"`
	Id    string        `bson:"id"`
	Key   string        `bson:"key"`
	Term  string        `bson:"term"`
	Exact bool          `bson:"exact"`
}

const textConditionAttrId = "id"
const textConditionAttrKey = "key"
const textConditionAttrTerm = "term"
const textConditionAttrExact = "exact"

var _ Condition = (*textCondition)(nil)

func encodeTextCondition(src condition.TextCondition) (dst textCondition, ids []string) {
	id := src.GetId()
	ids = append(ids, id)
	key := src.GetKey()
	term := src.GetTerm()
	dst = textCondition{
		Base: ConditionBase{
			Not: src.IsNot(),
		},
		Id:    id,
		Key:   key,
		Term:  term,
		Exact: src.IsExact(),
	}
	return
}

func decodeTextCondition(baseCond ConditionBase, term string, raw bson.M) (tc textCondition, err error) {
	tc.Base = baseCond
	tc.Term = term
	var ok bool
	tc.Id, ok = raw[textConditionAttrId].(string)
	if ok {
		tc.Key, ok = raw[textConditionAttrKey].(string)
	}
	if ok {
		tc.Exact, _ = raw[textConditionAttrExact].(bool)
	}
	if !ok {
		err = fmt.Errorf("%w: failed to decode the text condition %v", storage.ErrInternal, raw)
	}
	return
}
