package storage

import (
	"io"
	"subscriptions/patterns"
)

type (
	Metadata map[string]string

	Storage interface {
		io.Closer

		// Create a subscription means subscribing.
		Create(s Subscription) (string, error)

		// Read the specified subscription details.
		Read(id string) (Subscription, error)

		// Update the specified subscription with new details.
		Update(id string, s Subscription) error

		// Delete a subscription means unsubscribing.
		Delete(id string) error

		// List returns all known Subscription ids with the pagination support that match the specified query.
		List(limit uint32, cursor *string) ([]string, error)

		// Resolve returns all known Subscription ids whene the specified patterns are mentioned.
		Resolve(limit uint32, cursor *string, patternIds []patterns.Id) ([]string, error)
	}
)
