package mongo

type (
	matcherGroup struct {
		All      bool      `bson:"all"`
		Matchers []matcher `bson:"matchers,omitempty"`
	}
)

const (
	attrAll      = "all"
	attrMatchers = "matchers"
)
