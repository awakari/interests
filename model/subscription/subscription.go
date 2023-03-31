package subscription

// Subscription represents the subscription entry.
type Subscription struct {

	// Id represents the unique Subscription Id.
	Id string

	// Account represents an id of the Subscription owner.
	Account string

	// Data contains the Subscription payload, mutable and immutable parts.
	Data Data
}
