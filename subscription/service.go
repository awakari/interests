package subscription

import (
	"errors"
)

var (
	ErrConflict = errors.New("subscription already exists")

	ErrNotFound = errors.New("subscription was not found")
)

type (

	// Service is a subscription service
	Service interface {

		// Create a subscription means Subscribing
		Create(data Data) (Id, error)

		// Read the specified subscription details
		Read(id Id) (*Subscription, error)

		// Delete a subscription means Unsubscribing
		Delete(id Id) error

		// List returns all known Subscriptions with the pagination support that match the specified query
		List(query SubscriptionsQuery) (SubscriptionsPage, error)
	}
)
