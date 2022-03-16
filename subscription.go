package synapse

import (
	"synapse/util"
)

type (

	// Subscription connects source topic(s) with the destination message handler(s).
	Subscription struct {

		// Id is the unique Subscription identifier
		Id SubscriptionId

		SubscriptionData
	}

	// SubscriptionId is the unique Subscription identifier
	SubscriptionId string

	// SubscriptionData contains the subscription details
	SubscriptionData struct {

		// Description of the Subscription
		Description string

		// TopicIds are Subscription source topics
		TopicIds []TopicId

		// HandlerFactoryName is the initialization function unique name
		HandlerFactoryName string

		// HandlerConfig is the specific Handler configuration (e.g. specific e-mail address, phone number, queue name)
		HandlerConfig HandlerConfig
	}

	SubscriptionsPageCursor SubscriptionId

	SubscriptionsPage struct {
		util.Page[Subscription, SubscriptionsPageCursor]
	}

	SubscriptionsQuery struct {

		// TopicIdRef is the optional topic id to match
		TopicIdRef *TopicId

		util.PageQuery[SubscriptionsPageCursor]
	}
)
