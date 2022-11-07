package aggregator

import (
	"context"
	"github.com/meandros-messaging/subscriptions/model"
)

type (

	// MatchInGroup represents the reduced information about the model.MatcherGroup necessary to match it on the
	// aggregator side.
	MatchInGroup struct {

		// All represents the model.MatcherGroup matching constraint. See the corresponding field in the
		// model.MatcherGroup for details.
		All bool

		// MatcherCount represents the total model.Matcher count in the corresponding model.MatcherGroup.
		MatcherCount uint32
	}

	// Match represents an event of some model.Matcher matched a model.Message metadata for a certain
	// model.Subscription. Either Includes or Excludes should be set to a non-nil value (but not both simultaneously).
	Match struct {

		// MessageId represents the model.Message for which the current Match event occurred.
		MessageId model.MessageId

		// SubscriptionKey represents the model.Subscription for which the current Match event occurred.
		SubscriptionKey model.SubscriptionKey

		// Includes represents the corresponding model.Subscription includes model.Matcher group.
		// Nil if match occurred in another (Excludes) group.
		Includes *MatchInGroup

		// Includes represents the corresponding model.Subscription excludes model.Matcher group.
		// Nil if match occured in another (Includes) group.
		Excludes *MatchInGroup
	}

	// Service represents the aggregator service.
	Service interface {

		// Update sends the Match event for a further aggregation into a single record defined by the supplied
		// model.MessageId and model.SubscriptionKey pair in the event.
		Update(ctx context.Context, m Match) (err error)
	}
)
