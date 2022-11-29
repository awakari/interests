package mongo

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
