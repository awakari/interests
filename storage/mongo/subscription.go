package mongo

type (
	subscription struct {
		Name        string       `bson:"name"`
		Description string       `bson:"description"`
		Routes      []string     `bson:"routes"`
		Includes    matcherGroup `bson:"includes,omitempty"`
		Excludes    matcherGroup `bson:"excludes,omitempty"`
	}
)

const (
	attrName        = "name"
	attrDescription = "description"
	attrRoutes      = "routes"
	attrIncludes    = "includes"
	attrExcludes    = "excludes"
)
