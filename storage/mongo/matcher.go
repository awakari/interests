package mongo

type (
	matcher struct {
		Partial bool    `bson:"partial"`
		Key     string  `bson:"key"`
		Pattern pattern `bson:"pattern"`
	}
)

const (
	attrPartial = "partial"
	attrKey     = "key"
	attrPattern = "pattern"
)
