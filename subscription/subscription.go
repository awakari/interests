package subscription

import (
	"synapse/handler"
	"synapse/util"
)

type (

	// Subscription connects source metadata filters with the destination message handler(s).
	Subscription struct {

		// Id is the unique Subscription identifier
		Id Id

		Data
	}

	// Id is the unique Subscription identifier
	Id string

	// Data contains the subscription details
	Data struct {

		// Description of the Subscription
		Description string

		// MetadataFilter is the message.Message's metadata filter. It should contain regex by key to match.
		MetadataFilter map[string]string

		// HandlerFactoryName is the initialization function unique name
		HandlerFactoryName string

		// HandlerConfig is the specific Handler configuration (e.g. specific e-mail address, phone number, queue name)
		HandlerConfig handler.Config
	}

	MetadataFilter

	SubscriptionsPageCursor Id

	SubscriptionsPage struct {
		util.Page[Subscription, SubscriptionsPageCursor]
	}

	SubscriptionsQuery struct {

		// TopicIdRef is the optional topic id to match
		MetadataFilters []

		util.PageQuery[SubscriptionsPageCursor]
	}
)
