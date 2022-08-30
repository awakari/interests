package storage

import (
	"context"
	"io"
	"subscriptions/patterns"
)

type (

	// Storage represents the subscriptions storage
	Storage interface {
		io.Closer

		// Create a subscription means subscribing.
		Create(ctx context.Context, s Subscription) (string, error)

		// Read the specified subscription details.
		Read(ctx context.Context, id string) (Subscription, error)

		// Update the specified subscription with new details.
		Update(ctx context.Context, id string, s Subscription) error

		// Delete a subscription means unsubscribing.
		Delete(ctx context.Context, id string) error

		// List returns all known Subscription ids with the pagination support that match the specified query.
		List(ctx context.Context, limit uint32, cursor *string) ([]string, error)

		// Resolve returns all known Subscription ids where the specified patterns are mentioned under the specified key.
		Resolve(ctx context.Context, limit uint32, cursor *string, key string, patternIds []patterns.Id) ([]string, error)
	}
)
