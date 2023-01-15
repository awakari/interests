package mongo

import "go.mongodb.org/mongo-driver/bson"

type (
	subscriptionWrite struct {
		Name string `bson:"name"`

		Description string `bson:"description"`

		Routes []string `bson:"routes"`

		Condition Condition `bson:"condition"`

		// KiwiConditions contains the list of copies of all "kiwi" (key, value pattern) conditions under the Condition
		// field. The KiwiConditions field is necessary to support the subscriptions search by a "kiwi" condition.
		KiwiConditions []kiwiCondition `bson:"kiwi_conditions"`
	}

	// intermediate search result that contains the condition not decoded yet
	subscriptionRaw struct {
		Name string `bson:"name"`

		Routes []string `bson:"routes"`

		Condition bson.M `bson:"condition"`
	}
)

const (
	attrName           = "name"
	attrDescription    = "description"
	attrRoutes         = "routes"
	attrKiwiConditions = "kiwi_conditions"
	attrCondition      = "condition"
	attrIncludes       = "includes"
	attrExcludes       = "excludes"
)

func decodeSubscriptionSearchResult(raw subscriptionRaw) (result subscriptionSearchResult) {
	result.Name = raw.Name
	result.Routes = raw.Routes
	result.Condition, _ = decodeCondition(raw.Condition)
	return
}
