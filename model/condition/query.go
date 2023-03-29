package condition

type Query struct {

	// Limit defines a results page size limit.
	Limit uint32

	// Condition represents the subscription.Subscription search criteria: it should contain the specified Condition.
	Condition Condition
}
