package subscription

// ConditionMatch represents a subscription routing data with addition of a condition id that matched.
type ConditionMatch struct {

	// Id represents the unique Subscription Id.
	Id string

	// Route represents the Subscription routing data.
	Route Route

	// ConditionId represents the id of a Condition that matched.
	ConditionId string
}
