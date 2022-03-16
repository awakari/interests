package subscription

import (
	"errors"
	"synapse"
)

var (
	ErrConflict = errors.New("subscription already exists")

	ErrNotFound = errors.New("subscription was not found")
)

type (

	// Service is a subscription service
	Service interface {

		// Create a subscription means Subscribing
		Create(data synapse.SubscriptionData) (synapse.SubscriptionId, error)

		// Read the specified subscription details
		Read(id synapse.SubscriptionId) (*synapse.Subscription, error)

		// Delete a subscription means Unsubscribing
		Delete(id synapse.SubscriptionId) error

		// List returns all known Subscriptions with the pagination support that match the specified query
		List(query synapse.SubscriptionsQuery) (synapse.SubscriptionsPage, error)
	}
)
