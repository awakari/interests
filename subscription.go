package synapse

import (
	"synapse/util"
)

// SubscriptionId is the unique Subscription identifier
type SubscriptionId int

// SubscriptionData contains the subscription details
type SubscriptionData struct {

	// Description of the Subscription
	Description string

	// TopicIds are Subscription source topics
	TopicIds []TopicId

	// Handler is a generic and persistent message processing function
	HandlerName string

	// Config is the specific Handler configuration (e.g. specific e-mail address, phone number, queue name)
	Config map[string]interface{}
}

// Subscription connects source topic(s) with the destination message handler(s).
// When created
type Subscription struct {

	// Id is the unique Subscription identifier
	Id SubscriptionId

	SubscriptionData
}

type SubscriptionsPageCursor int

type SubscriptionsPage struct {
	util.ResultsPage[Subscription, SubscriptionsPageCursor]
}

type SubscriptionRuntime struct {
	Subscription

	HandleFunc util.HandleFunc[Message]

	HandleBatchFunc util.HandleBatchFunc[Message]
}
