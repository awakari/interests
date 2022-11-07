package model

type (

	// SubscriptionKey is a key part of the Subscription that uniquely defines it.
	SubscriptionKey struct {

		// Name represents a unique subscription name.
		Name string

		// Version represents a Subscription entry version for the optimistic lock purpose.
		Version uint64
	}
)
