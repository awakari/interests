package model

// SubscriptionMatch represents a subscription routing data with addition of a condition id that matched.
type SubscriptionMatch struct {

	// Id represents the unique Subscription Id.
	Id string

	// Routes represents a list of routes associated with the Subscription.
	Routes []string

	// Condition represents the certain criteria to select the Subscription.
	Condition Condition

	// MatchId represents the id of a Condition that matched.
	MatchId string
}
