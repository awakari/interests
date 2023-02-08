package mongo

import (
	"github.com/awakari/subscriptions/model/subscription"
	"go.mongodb.org/mongo-driver/bson"
)

type subscriptionWrite struct {
	Id string `bson:"id"`

	Metadata map[string]string `bson:"metadata"`

	Destinations []string `bson:"destinations"`

	Condition Condition `bson:"condition"`

	// Kiwis contains a flat list of copies of all kiwi conditions.
	// The Kiwis field is necessary to support the subscriptions search by a "Kiwi".
	Kiwis []kiwiSearchData `bson:"kiwis"`
}

type kiwiSearchData struct {
	Id      string `bson:"id"`
	Partial bool   `bson:"partial"`
	Key     string `bson:"key"`
	Pattern string `bson:"pattern"`
}

// intermediate read result that contains the condition not decoded yet
type subscriptionRec struct {
	Id string `bson:"id"`

	Metadata map[string]string `bson:"metadata"`

	Destinations []string `bson:"destinations"`

	RawCondition bson.M `bson:"condition"`

	// Kiwis contains a flat list of copies of all kiwi conditions.
	// The Kiwis field is necessary to support the subscriptions search by a "Kiwi".
	Kiwis []kiwiSearchData `bson:"kiwis"`
}

const attrId = "id"
const attrMetadata = "metadata"
const attrRoutes = "routes"
const attrKiwis = "kiwis"
const attrCondition = "condition"

func (rec subscriptionRec) decodeSubscription(sub *subscription.Subscription) (err error) {
	sub.Id = rec.Id
	err = rec.decodeSubscriptionData(&sub.Data)
	return
}

func (rec subscriptionRec) decodeSubscriptionData(sd *subscription.Data) (err error) {
	sd.Metadata = rec.Metadata
	err = rec.decodeSubscriptionRoute(&sd.Route)
	return
}

func (rec subscriptionRec) decodeSubscriptionRoute(sr *subscription.Route) (err error) {
	sr.Destinations = rec.Destinations
	var condRec Condition
	condRec, err = decodeRawCondition(rec.RawCondition)
	if err == nil {
		sr.Condition = decodeCondition(condRec)
	}
	return
}

func (rec subscriptionRec) decodeSubscriptionConditionMatch(cm *subscription.ConditionMatch) (err error) {
	cm.Id = rec.Id
	err = rec.decodeSubscriptionRoute(&cm.Route)
	return
}
