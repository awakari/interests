package mongo

type (
	subscription struct {
		Name        string       `bson:"name"`
		Description string       `bson:"description"`
		Includes    matcherGroup `bson:"includes,omitempty"`
		Excludes    matcherGroup `bson:"excludes,omitempty"`
	}
)

const (
	attrName        = "name"
	attrDescription = "description"
	attrIncludes    = "includes"
	attrExcludes    = "excludes"
)
