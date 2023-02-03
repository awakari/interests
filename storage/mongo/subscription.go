package mongo

import (
	"github.com/awakari/subscriptions/model"
	"go.mongodb.org/mongo-driver/bson"
)

type subscriptionWrite struct {
	Id string `bson:"id"`

	Metadata map[string]string `bson:"metadata"`

	Routes []string `bson:"routes"`

	Condition Condition `bson:"condition"`

	// Kiwis contains the list of copies of all key-pattern pairs. The Kiwis field is necessary to support the
	// subscriptions search by a "Kiwi".
	Kiwis []kiwiSearchData `bson:"kiwis"`
}

type kiwiSearchData struct {
	Partial bool   `bson:"partial"`
	Key     string `bson:"key"`
	Pattern string `bson:"pattern"`
}

// intermediate read result that contains the condition not decoded yet
type subscription struct {
	Id string `bson:"id"`

	Description string `bson:"description"`

	Metadata map[string]string `bson:"metadata"`

	Routes []string `bson:"routes"`

	RawCondition bson.M `bson:"condition"`
}

const attrId = "id"
const attrMetadata = "metadata"
const attrRoutes = "routes"
const attrKiwis = "kiwis"
const attrCondition = "condition"

func (rec subscription) decodeSubscription(sub *model.Subscription) (err error) {
	sub.Id = rec.Id
	err = rec.decodeSubscriptionData(&sub.Data)
	return
}

func (rec subscription) decodeSubscriptionData(sd *model.SubscriptionData) (err error) {
	sd.Metadata = rec.Metadata
	sd.Routes = rec.Routes
	var condRec Condition
	condRec, err = decodeRawCondition(rec.RawCondition)
	if err == nil {
		sd.Condition = decodeCondition(condRec)
	}
	return
}
