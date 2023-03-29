package subscription

import "github.com/awakari/subscriptions/model/condition"

// ConditionMatch represents a subscription id and a condition id that matched.
type ConditionMatch struct {

	// Id represents the unique Subscription Id.
	Id string

	// Account represents the Subscription owner id.
	Account string

	// ConditionId represents the id of a Condition that matched.
	ConditionId string

	// Condition represents the root Subscription condition.
	Condition condition.Condition
}
