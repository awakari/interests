package mongo

import (
	"github.com/awakari/interests/model/interest"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type interestWrite struct {
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
	// The CondIds field is necessary to support the interests search by a condition id.
	CondIds []string `bson:"condIds"`

	RateLimit int64 `bson:"rateLimit,omitempty"`
}

// intermediate read result that contains the condition not decoded yet
type interestRec struct {
	Id string `bson:"id"`

	GroupId string `bson:"groupId"`

	UserId string `bson:"userId"`

	Description string `bson:"descr"`

	Enabled bool `bson:"enabled"`

	EnabledSince time.Time `bson:"enabledSince,omitempty"`

	Expires time.Time `bson:"expires,omitempty"`

	Created time.Time `bson:"created,omitempty"`

	Updated time.Time `bson:"updated,omitempty"`

	Result time.Time `bson:"result,omitempty"`

	Public bool `bson:"public,omitempty"`

	Followers int64 `bson:"followers,omitempty"`

	RawCondition bson.M `bson:"cond"`

	// CondIds contains a flat list of all condition ids.
	// The CondIds field is necessary to support the interests search by a condition id.
	CondIds []string `bson:"condIds"`

	RateLimit int64 `bson:"rateLimit,omitempty"`
}

const attrId = "id"
const attrGroupId = "groupId"
const attrUserId = "userId"
const attrDescr = "descr"
const attrEnabled = "enabled"
const attrEnabledSince = "enabledSince"
const attrExpires = "expires"
const attrCreated = "created"
const attrUpdated = "updated"
const attrResult = "result"
const attrPublic = "public"
const attrFollowers = "followers"
const attrCondIds = "condIds"
const attrCond = "cond"
const attrRateLimit = "rateLimit"

func (rec interestRec) decodeInterest(sub *interest.Interest) (err error) {
	sub.Id = rec.Id
	sub.GroupId = rec.GroupId
	sub.UserId = rec.UserId
	err = rec.decodeInterestData(&sub.Data)
	return
}

func (rec interestRec) decodeInterestData(sd *interest.Data) (err error) {
	sd.Description = rec.Description
	sd.Enabled = rec.Enabled
	sd.EnabledSince = rec.EnabledSince
	sd.Expires = rec.Expires
	sd.Created = rec.Created
	sd.Updated = rec.Updated
	sd.Result = rec.Result
	sd.Public = rec.Public
	sd.Followers = rec.Followers
	sd.RateLimit = rec.RateLimit
	var condRec Condition
	condRec, err = decodeRawCondition(rec.RawCondition)
	if err == nil {
		sd.Condition = decodeCondition(condRec)
	}
	return
}

func (rec interestRec) decodeInterestConditionMatch(cm *interest.ConditionMatch) (err error) {
	cm.InterestId = rec.Id
	var condRec Condition
	condRec, err = decodeRawCondition(rec.RawCondition)
	if err == nil {
		cm.Condition = decodeCondition(condRec)
	}
	return
}
