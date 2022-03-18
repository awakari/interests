package subscription

import "synapse"

// Service is a subscription service
type Service interface {

	// Create a subscription means Subscribing
	Create(data synapse.SubscriptionData) (synapse.SubscriptionId, error)

	// Read the specified subscription details
	Read(id synapse.SubscriptionId) (synapse.SubscriptionData, error)

	// Delete a subscription means Unsubscribing
	Delete(id synapse.SubscriptionId) error

	// ListAll returns all known Subscriptions with the pagination support
	ListAll(cursor synapse.SubscriptionsPageCursor) (synapse.SubscriptionsPage, error)

	// ListByTopicIds returns all known Subscriptions whose connected to any of the specified Topics
	ListByTopicIds([]synapse.TopicId, synapse.SubscriptionsPageCursor) (synapse.SubscriptionsPage, error)
}
