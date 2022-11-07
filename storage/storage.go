package storage

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/model"
	"io"
)

type (

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

		// ListNames returns all known model.Subscription names with the pagination support that match the specified query.
		ListNames(ctx context.Context, limit uint32, cursor string) (page []string, err error)

		// FindByMatcherData returns candidate subscriptions page where:<br/>
		// * model.Subscription name is greater than the cursor:<br/>
		// * and that have any model.Matcher in any associated model.MatcherGroup
		FindByMatcherData(ctx context.Context, limit uint32, cursor string, md model.MatcherData) (page []model.Subscription, err error)
	}
)

var (
	ErrNotFound = errors.New("subscription was not found")
)
