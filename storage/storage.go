package storage

import (
	"context"
	"errors"
	"github.com/meandros-messaging/subscriptions/service/patterns"
	"io"
)

type (

	// Storage represents the subscriptions storage
	Storage interface {
		io.Closer

		// Create a subscription means subscribing.
		Create(ctx context.Context, sub Subscription) (err error)

		// Read the specified subscription details.
		Read(ctx context.Context, name string) (sub Subscription, err error)

		// Update the specified subscription with new details.
		Update(ctx context.Context, sub Subscription) (err error)

		// Delete a subscription means unsubscribing.
		Delete(ctx context.Context, name string) (err error)

		// List returns all known Subscription.Name with the pagination support that match the specified query.
		List(ctx context.Context, limit uint32, cursor *string) (page []string, err error)

		// FindCandidates returns candidate subscriptions page where:<br/>
		// * Subscription.Name is greater than the cursor:<br/>
		// * and that have any Matcher in the Subscription.Includes MatcherGroup where:<br/>
		//		* Matcher.Key equals to the specified one:<br/>
		// 		* and that Matcher.PatternCode equals to the specified one.<br/>
		FindCandidates(ctx context.Context, limit uint32, cursor *string, key string, patternCode patterns.PatternCode) (page []Subscription, err error)
	}
)

var (
	ErrNotFound = errors.New("subscription was not found")
)
