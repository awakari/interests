package subscription

// Subscription represents the subscription entry.
type Subscription struct {

	// Id represents the unique Subscription Id.
	Id string

	GroupId string

	// UserId represents an id of the Subscription owner.
	UserId string

	// Data contains the Subscription payload, mutable and immutable parts.
	Data Data
}
