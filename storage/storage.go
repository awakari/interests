package storage

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
	"io"
)

type (

	// Query represents the search query to use in Storage.Find
	Query struct {

		// Limit defines a results page size limit.
		Limit uint32

		// Includes defines if it's necessary to find a model.Subscription with same model.Matcher in the "Excludes"
		// model.MatcherGroup
		Includes bool

		// Excludes defines if it's necessary to find a model.Subscription with same model.Matcher in the "Excludes"
		// model.MatcherGroup
		Excludes bool

		// Matcher represents a model.Matcher that should be present in the model.Subscription to include into the
		// search results.
		Matcher model.Matcher
	}

	// Storage represents the subscriptions storage
	Storage interface {
		io.Closer

		// Create a subscription means subscribing.
		Create(ctx context.Context, sub model.Subscription) (err error)

		// Read the specified subscription details.
		Read(ctx context.Context, name string) (sub model.Subscription, err error)

		// Update the specified subscription with new details.
		Update(ctx context.Context, sub model.Subscription) (err error)

		// DeleteVersion the same as Delete but removes the subscription only if version matches.
		DeleteVersion(ctx context.Context, subKey model.SubscriptionKey) (err error)

		// ListNames returns all known subscription names with the pagination support that match the specified query.
		ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error)

		// Find returns subscriptions page where:<br/>
		// * model.Subscription name is greater than the one specified by the cursor<br/>
		// *
		Find(ctx context.Context, q Query, cursor string) (page []model.Subscription, err error)
	}
)

var (
	ErrNotFound = errors.New("subscription was not found")
)
