package aggregator

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
)

type (

	// MatchGroup represents the reduced information about the model.MatcherGroup necessary to match it on the
	// aggregator side.
	MatchGroup struct {

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

		// SubscriptionName represents the model.Subscription unique name for which the current Match event occurred.
		SubscriptionName string

		// InExcludes defines whether the match occurred in the Excludes matcher group.
		// Match occurred in Includes matcher group otherwise (when false).
		InExcludes bool

		// Includes represents the corresponding model.Subscription includes model.Matcher group.
		Includes MatchGroup

		// Includes represents the corresponding model.Subscription excludes model.Matcher group.
		Excludes MatchGroup
	}

	// Service represents the aggregator service.
	Service interface {

		// Update sends the Match event for a further aggregation into a single record defined by the supplied
		// model.MessageId and model.Subscription name pair in the event.
		Update(ctx context.Context, m Match) (err error)
	}
)

var (

	// ErrInternal indicates some unexpected internal failure.
	ErrInternal = errors.New("internal failure")
)
