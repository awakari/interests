package mongo

import (
	"github.com/awakari/subscriptions/model/subscription"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type subscriptionWrite struct {
	Id string `bson:"id"`

	GroupId string `bson:"groupId"`

	UserId string `bson:"userId"`

	Description string `bson:"descr"`

	Enabled bool `bson:"enabled"`

	Expires time.Time `bson:"expires,omitempty"`

	Created time.Time `bson:"created,omitempty"`

	Updated time.Time `bson:"updated,omitempty"`

	Public bool `bson:"public,omitempty"`

	Followers int64 `bson:"followers"`

	Condition Condition `bson:"cond"`

	// CondIds contains a flat list of all condition ids.
	// The CondIds field is necessary to support the subscriptions search by a condition id.
	CondIds []string `bson:"condIds"`
}

// intermediate read result that contains the condition not decoded yet
type subscriptionRec struct {
	Id string `bson:"id"`

	GroupId string `bson:"groupId"`

	UserId string `bson:"userId"`

	Description string `bson:"descr"`

	Enabled bool `bson:"enabled"`

	Expires time.Time `bson:"expires,omitempty"`

	Created time.Time `bson:"created,omitempty"`

	Updated time.Time `bson:"updated,omitempty"`

	Public bool `bson:"public,omitempty"`

	Followers int64 `bson:"followers,omitempty"`

	RawCondition bson.M `bson:"cond"`

	// CondIds contains a flat list of all condition ids.
	// The CondIds field is necessary to support the subscriptions search by a condition id.
	CondIds []string `bson:"condIds"`
}

const attrId = "id"
const attrGroupId = "groupId"
const attrUserId = "userId"
const attrDescr = "descr"
const attrEnabled = "enabled"
const attrExpires = "expires"
const attrCreated = "created"
const attrUpdated = "updated"
const attrPublic = "public"
const attrFollowers = "followers"
const attrCondIds = "condIds"
const attrCond = "cond"

func (rec subscriptionRec) decodeSubscription(sub *subscription.Subscription) (err error) {
	sub.Id = rec.Id
	sub.GroupId = rec.GroupId
	sub.UserId = rec.UserId
	err = rec.decodeSubscriptionData(&sub.Data)
	return
}

func (rec subscriptionRec) decodeSubscriptionData(sd *subscription.Data) (err error) {
	sd.Description = rec.Description
	sd.Enabled = rec.Enabled
	sd.Expires = rec.Expires
	sd.Created = rec.Created
	sd.Updated = rec.Updated
	sd.Public = rec.Public
	sd.Followers = rec.Followers
	var condRec Condition
	condRec, err = decodeRawCondition(rec.RawCondition)
	if err == nil {
		sd.Condition = decodeCondition(condRec)
	}
	return
}

func (rec subscriptionRec) decodeSubscriptionConditionMatch(cm *subscription.ConditionMatch) (err error) {
	cm.SubscriptionId = rec.Id
	var condRec Condition
	condRec, err = decodeRawCondition(rec.RawCondition)
	if err == nil {
		cm.Condition = decodeCondition(condRec)
	}
	return
}
