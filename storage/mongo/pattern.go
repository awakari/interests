package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	pattern struct {
		Code []byte `bson:"code"`
		Src  string `bson:"src"`
	}
)

const (
	attrCode = "code"
	attrSrc  = "src"
)

func decodePattern(raw bson.M) (p pattern, err error) {
	p.Code = raw[attrCode].(primitive.Binary).Data
	p.Src = raw[attrSrc].(string)
	return
}
