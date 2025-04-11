package mongo

import (
	"fmt"
	"github.com/awakari/interests/model/condition"
	"github.com/awakari/interests/storage"
	"go.mongodb.org/mongo-driver/bson"
)

type semCondition struct {
	Base          ConditionBase `bson:"base"`
	Id            string        `bson:"id"`
	Query         string        `bson:"q"`
	SimilarityMin float32       `bson:"similarity"`
}

const semConditionAttrId = "id"
const semConditionAttrQuery = "q"
const semConditionAttrSimilarityMin = "similarity"

const similarityMinDefault = 0.85

var _ Condition = (*semCondition)(nil)

func encodeSemCondition(src condition.SemanticCondition) (dst semCondition, ids []string) {
	id := src.GetId()
	ids = append(ids, id)
	q := src.Query()
	similarityMin := src.SimilarityMin()
	if similarityMin == 0 {
		similarityMin = similarityMinDefault
	}
	dst = semCondition{
		Base: ConditionBase{
			Not: src.IsNot(),
		},
		Id:            id,
		Query:         q,
		SimilarityMin: similarityMin,
	}
	return
}

func decodeSemCondition(baseCond ConditionBase, q string, raw bson.M) (sc semCondition, err error) {
	sc.Base = baseCond
	sc.Query = q
	similarityMinRaw, similarityMinOk := raw[semConditionAttrSimilarityMin]
	if similarityMinOk {
		similarityMinRaw, similarityMinOk = similarityMinRaw.(float64)
	}
	switch similarityMinOk {
	case true:
		sc.SimilarityMin = float32(similarityMinRaw.(float64))
	default:
		sc.SimilarityMin = similarityMinDefault
	}
	var ok bool
	sc.Id, ok = raw[semConditionAttrId].(string)
	if !ok {
		err = fmt.Errorf("%w: failed to decode the semantic condition %v", storage.ErrInternal, raw)
	}
	return
}
