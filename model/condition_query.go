package model

type ConditionQuery struct {

	// Limit defines a results page size limit.
	Limit uint32

	// Condition represents the Subscription search criteria: it should contain the specified Condition.
	Condition Condition
}
