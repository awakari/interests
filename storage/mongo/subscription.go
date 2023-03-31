package mongo

import (
	"github.com/awakari/subscriptions/model/subscription"
	"go.mongodb.org/mongo-driver/bson"
)

type subscriptionWrite struct {
	Id string `bson:"id"`

	Account string `bson:"acc"`

	Description string `bson:"descr"`

	Priority uint32 `bson:"prio"`

	Condition Condition `bson:"cond"`

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

	Account string `bson:"acc"`

	Description string `bson:"descr"`

	Priority uint32 `bson:"prio"`

	RawCondition bson.M `bson:"cond"`

	// Kiwis contains a flat list of copies of all kiwi conditions.
	// The Kiwis field is necessary to support the subscriptions search by a "Kiwi".
	Kiwis []kiwiSearchData `bson:"kiwis"`
}

const attrId = "id"
const attrAcc = "acc"
const attrDescr = "descr"
const attrPrio = "prio"
const attrKiwis = "kiwis"
const attrCond = "cond"

func (rec subscriptionRec) decodeSubscription(sub *subscription.Subscription) (err error) {
	sub.Id = rec.Id
	sub.Account = rec.Account
	err = rec.decodeSubscriptionData(&sub.Data)
	return
}

func (rec subscriptionRec) decodeSubscriptionData(sd *subscription.Data) (err error) {
	rec.decodeSubscriptionMetadata(&sd.Metadata)
	var condRec Condition
	condRec, err = decodeRawCondition(rec.RawCondition)
	if err == nil {
		sd.Condition = decodeCondition(condRec)
	}
	return
}

func (rec subscriptionRec) decodeSubscriptionMetadata(smd *subscription.Metadata) {
	smd.Description = rec.Description
	smd.Priority = rec.Priority
}

func (rec subscriptionRec) decodeSubscriptionConditionMatch(cm *subscription.ConditionMatch) (err error) {
	cm.Key = subscription.ConditionMatchKey{
		Id:       rec.Id,
		Priority: rec.Priority,
	}
	cm.Account = rec.Account
	var condRec Condition
	condRec, err = decodeRawCondition(rec.RawCondition)
	if err == nil {
		cm.Condition = decodeCondition(condRec)
	}
	return
}
