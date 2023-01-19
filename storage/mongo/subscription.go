package mongo

import (
	"github.com/awakari/subscriptions/model"
	"go.mongodb.org/mongo-driver/bson"
)

type subscriptionWrite struct {
	Name string `bson:"name"`

	Description string `bson:"description"`

	Routes []string `bson:"routes"`

	Condition Condition `bson:"condition"`

	// Kiwis contains the list of copies of all key-pattern pairs. The Kiwis field is necessary to support the
	// subscriptions search by a "Kiwi".
	Kiwis []kiwiCondition `bson:"kiwis"`
}

// intermediate read result that contains the condition not decoded yet
type subscription struct {
	Name string `bson:"name"`

	Description string `bson:"description"`

	Routes []string `bson:"routes"`

	RawCondition bson.M `bson:"condition"`
}

const attrName = "name"
const attrDescription = "description"
const attrRoutes = "routes"
const attrKiwis = "kiwis"
const attrCondition = "condition"

func (rec subscription) decodeSubscription(sub *model.Subscription) (err error) {
	sub.Name = rec.Name
	sub.Description = rec.Description
	sub.Routes = rec.Routes
	var condRec Condition
	condRec, err = decodeRawCondition(rec.RawCondition)
	if err == nil {
		sub.Condition = decodeCondition(condRec)
	}
	return
}
